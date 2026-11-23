package handlers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/albertson/albertsonmissionplan/internal/challenges"
)

func newChallengeHarness(t *testing.T) *galleryHarness {
	t.Helper()
	return newGalleryHarness(t) // registers the challenge routes on the same server
}

type challengeList struct {
	Items []struct {
		Month           int                 `json:"month"`
		Heading         string              `json:"heading"`
		Goal            string              `json:"goal"`
		Bullets         []challenges.Bullet `json:"bullets"`
		ImageURL        string              `json:"imageUrl"`
		ThumbURL        string              `json:"thumbUrl"`
		Hidden          bool                `json:"hidden"`
		IsOriginalText  bool                `json:"isOriginalText"`
		IsOriginalImage bool                `json:"isOriginalImage"`
		UpdatedBy       string              `json:"updatedBy"`
	} `json:"items"`
	Error string `json:"error"`
}

func decodeList(t *testing.T, rec *httptest.ResponseRecorder) challengeList {
	t.Helper()
	var out challengeList
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return out
}

func (h *galleryHarness) get(path string, asLeader bool) *httptest.ResponseRecorder {
	return h.do(httptest.NewRequest(http.MethodGet, path, nil), asLeader)
}

func (h *galleryHarness) uploadTo(path string, data []byte, asLeader bool) *httptest.ResponseRecorder {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("image", "art.jpg")
	_, _ = fw.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return h.do(req, asLeader)
}

func TestChallengesAPI(t *testing.T) {
	h := newChallengeHarness(t)

	// Public list: twelve seeded months, identical to the original static content.
	rec := h.get("/api/challenges", false)
	pub := decodeList(t, rec)
	if rec.Code != 200 || len(pub.Items) != 12 {
		t.Fatalf("public list %d: %s", rec.Code, rec.Body.String())
	}
	for i, c := range pub.Items {
		d := challenges.Defaults[i]
		if c.Month != i+1 || c.Heading != d.Heading || c.Goal != d.Goal || len(c.Bullets) != len(d.Bullets) {
			t.Fatalf("month %d = %+v", i+1, c)
		}
	}
	img := h.get(pub.Items[0].ImageURL, false)
	if img.Code != 200 || img.Header().Get("Content-Type") != "image/webp" ||
		!strings.Contains(img.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("image %d %v", img.Code, img.Header())
	}
	if h.get("/api/challenges/13/image", false).Code != 404 || h.get("/api/challenges/1/big", false).Code != 404 {
		t.Fatal("bad image paths should 404")
	}

	// Leader endpoints require a signed-in leader.
	for _, path := range []string{"/api/admin/challenges/1", "/api/admin/challenges/1/reset", "/api/admin/challenges/1/move"} {
		if rec := h.postJSON(path, map[string]any{}, false); rec.Code != http.StatusForbidden {
			t.Fatalf("anonymous POST %s = %d", path, rec.Code)
		}
	}
	if rec := h.get("/api/admin/challenges", false); rec.Code != http.StatusForbidden {
		t.Fatalf("anonymous admin list = %d", rec.Code)
	}
	if rec := h.uploadTo("/api/admin/challenges/1/image", sampleJPEG(t, 50, 50), false); rec.Code != http.StatusForbidden {
		t.Fatalf("anonymous upload = %d", rec.Code)
	}

	// Validation.
	bad := []map[string]any{
		{"heading": "", "goal": "g", "bullets": []map[string]string{{"who": "A", "text": "b"}}},
		{"heading": "H", "goal": "", "bullets": []map[string]string{{"who": "A", "text": "b"}}},
		{"heading": "H", "goal": "g", "bullets": []map[string]string{}},
		{"heading": "H", "goal": "g", "bullets": []map[string]string{{"who": "", "text": "b"}}},
		{"heading": "H", "goal": "g", "bullets": []map[string]string{{"who": "A", "text": ""}}},
		{"heading": strings.Repeat("x", 61), "goal": "g", "bullets": []map[string]string{{"who": "A", "text": "b"}}},
	}
	for i, b := range bad {
		if rec := h.postJSON("/api/admin/challenges/2", b, true); rec.Code != http.StatusBadRequest || decodeList(t, rec).Error == "" {
			t.Fatalf("bad[%d] = %d %s", i, rec.Code, rec.Body.String())
		}
	}
	if rec := h.postJSON("/api/admin/challenges/13", bad[0], true); rec.Code != http.StatusNotFound {
		t.Fatalf("month 13 = %d", rec.Code)
	}

	// Save: whitespace is tidied and blank rows dropped; hidden months leave the public list.
	rec = h.postJSON("/api/admin/challenges/2", map[string]any{
		"heading": "  Kindness   Challenge ", "goal": "Be kind\nevery day", "hidden": true,
		"bullets": []map[string]string{{"who": "Everyone", "text": "See [this](https://example.org)."}, {"who": " ", "text": ""}},
	}, true)
	adm := decodeList(t, rec)
	if rec.Code != 200 || len(adm.Items) != 12 {
		t.Fatalf("save = %d %s", rec.Code, rec.Body.String())
	}
	feb := adm.Items[1]
	if feb.Heading != "Kindness Challenge" || feb.Goal != "Be kind every day" || len(feb.Bullets) != 1 || !feb.Hidden ||
		feb.IsOriginalText || !feb.IsOriginalImage || feb.UpdatedBy != "lead@example.org" {
		t.Fatalf("saved feb = %+v", feb)
	}
	if pub := decodeList(t, h.get("/api/challenges", false)); len(pub.Items) != 11 {
		t.Fatalf("public list after hiding = %d", len(pub.Items))
	}
	if h.get(feb.ImageURL, false).Code != 404 || h.get(feb.ImageURL, true).Code != 200 {
		t.Fatal("hidden month's picture should be leader-only")
	}

	// Replace a picture: a tall upload is cropped to the 3:2 card and re-encoded as JPEG.
	rec = h.uploadTo("/api/admin/challenges/2/image", sampleJPEG(t, 300, 600), true)
	adm = decodeList(t, rec)
	if rec.Code != 200 || adm.Items[1].IsOriginalImage {
		t.Fatalf("upload = %d %s", rec.Code, rec.Body.String())
	}
	got := h.get(adm.Items[1].ImageURL, true)
	cfg, format, err := image.DecodeConfig(bytes.NewReader(got.Body.Bytes()))
	if err != nil || format != "jpeg" || cfg.Width != 960 || cfg.Height != 640 {
		t.Fatalf("uploaded image %s %dx%d %v", format, cfg.Width, cfg.Height, err)
	}
	thumb := h.get(adm.Items[1].ThumbURL, true)
	if c, err := jpeg.DecodeConfig(bytes.NewReader(thumb.Body.Bytes())); err != nil || c.Width != 480 || c.Height != 320 {
		t.Fatalf("thumb %+v %v", c, err)
	}
	if rec := h.uploadTo("/api/admin/challenges/2/image", []byte("not an image"), true); rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("bad upload = %d", rec.Code)
	}

	// Move February's challenge to July (swap), then back.
	rec = h.postJSON("/api/admin/challenges/2/move", map[string]int{"to": 7}, true)
	adm = decodeList(t, rec)
	if rec.Code != 200 || adm.Items[1].Heading != "Gratitude Challenge" || adm.Items[6].Heading != "Kindness Challenge" || !adm.Items[6].Hidden {
		t.Fatalf("move = %d %+v", rec.Code, adm.Items)
	}
	if rec := h.postJSON("/api/admin/challenges/2/move", map[string]int{"to": 0}, true); rec.Code != http.StatusBadRequest {
		t.Fatalf("move to 0 = %d", rec.Code)
	}
	h.postJSON("/api/admin/challenges/7/move", map[string]int{"to": 2}, true)

	// Reset only the text keeps the uploaded picture and visibility; full reset restores all.
	adm = decodeList(t, h.postJSON("/api/admin/challenges/2/reset", map[string]string{"part": "text"}, true))
	if f := adm.Items[1]; !f.IsOriginalText || f.IsOriginalImage || !f.Hidden {
		t.Fatalf("text reset = %+v", f)
	}
	adm = decodeList(t, h.postJSON("/api/admin/challenges/2/reset", map[string]string{"part": "image"}, true))
	if f := adm.Items[1]; !f.IsOriginalImage {
		t.Fatalf("image reset = %+v", f)
	}
	h.postJSON("/api/admin/challenges/2", map[string]any{"heading": "X", "goal": "Y", "bullets": []map[string]string{{"who": "A", "text": "b"}}, "hidden": true}, true)
	adm = decodeList(t, h.postJSON("/api/admin/challenges/2/reset", map[string]string{}, true))
	if f := adm.Items[1]; !f.IsOriginalText || !f.IsOriginalImage || f.Hidden {
		t.Fatalf("full reset = %+v", f)
	}
	if rec := h.postJSON("/api/admin/challenges/2/reset", map[string]string{"part": "nope"}, true); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad reset part = %d", rec.Code)
	}
}
