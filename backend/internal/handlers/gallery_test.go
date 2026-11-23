package handlers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type galleryHarness struct {
	t      *testing.T
	mux    *http.ServeMux
	leader []*http.Cookie
}

func newGalleryHarness(t *testing.T) *galleryHarness {
	t.Helper()
	s := newTestServer(t, nil)
	mux := http.NewServeMux()
	s.RegisterGalleryRoutes(mux)
	// First registration is auto-approved and signed in; keep its session cookie.
	body, _ := json.Marshal(map[string]string{"email": "lead@example.org", "password": "pw-lead-123", "firstName": "Lea", "lastName": "Der"})
	rec := httptest.NewRecorder()
	s.APIAuthRegister(rec, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}
	return &galleryHarness{t: t, mux: mux, leader: rec.Result().Cookies()}
}

func (h *galleryHarness) do(req *http.Request, asLeader bool) *httptest.ResponseRecorder {
	if asLeader {
		for _, c := range h.leader {
			req.AddCookie(c)
		}
	}
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	return rec
}

func (h *galleryHarness) postJSON(path string, v any, asLeader bool) *httptest.ResponseRecorder {
	b, _ := json.Marshal(v)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return h.do(req, asLeader)
}

func sampleJPEG(t *testing.T, w, hgt int) []byte {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, w, hgt))
	for y := 0; y < hgt; y++ {
		for x := 0; x < w; x++ {
			im.Set(x, y, color.RGBA{uint8(x), uint8(y), 180, 255})
		}
	}
	var b bytes.Buffer
	_ = jpeg.Encode(&b, im, nil)
	return b.Bytes()
}

func (h *galleryHarness) upload(data []byte, fields map[string]string, asLeader bool) *httptest.ResponseRecorder {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	fw, _ := mw.CreateFormFile("image", "scan.jpg")
	_, _ = fw.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/gallery", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return h.do(req, asLeader)
}

type itemsResp struct {
	Items []galleryItemDTO `json:"items"`
}

func (h *galleryHarness) list(path string, asLeader bool) []galleryItemDTO {
	rec := h.do(httptest.NewRequest(http.MethodGet, path, nil), asLeader)
	if rec.Code != http.StatusOK {
		h.t.Fatalf("GET %s: %d %s", path, rec.Code, rec.Body.String())
	}
	var out itemsResp
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		h.t.Fatal(err)
	}
	return out.Items
}

func TestGalleryUploadListServeEditDelete(t *testing.T) {
	h := newGalleryHarness(t)

	if got := h.list("/api/gallery", false); len(got) != 0 {
		t.Fatalf("empty gallery returned %d items", len(got))
	}

	// Any text fields a client sends are ignored: the gallery stores no names or labels.
	rec := h.upload(sampleJPEG(t, 2400, 1800), map[string]string{
		"title": "Jesus and the children", "caption": "Drawn in Primary.", "artist": "Mia", "ageGroup": "Sunbeams",
	}, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body.String())
	}
	var up struct {
		Item galleryItemDTO `json:"item"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &up)
	it := up.Item
	for _, leak := range []string{"Mia", "Sunbeams", "Jesus and the children", "Drawn in Primary", "artist", "title", "caption"} {
		if strings.Contains(rec.Body.String(), leak) {
			t.Errorf("upload response contains %q: %s", leak, rec.Body.String())
		}
	}
	if it.Width != 1600 || it.Height != 1200 || it.ThumbWidth != 640 {
		t.Errorf("sizes: %dx%d thumb %d", it.Width, it.Height, it.ThumbWidth)
	}
	if it.Alt != "Drawing by a ward member" {
		t.Errorf("alt = %q", it.Alt)
	}

	// The API lists newest first (the public page shuffles client-side).
	rec = h.upload(sampleJPEG(t, 300, 400), nil, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload 2: %d", rec.Code)
	}
	pub := h.list("/api/gallery", false)
	if len(pub) != 2 || pub[1].ID != it.ID {
		t.Fatalf("public order: %+v", pub)
	}
	templeID := pub[0].ID

	// Image endpoint: JPEG bytes, year-long cache for the versioned URL.
	img := h.do(httptest.NewRequest(http.MethodGet, it.ImageURL, nil), false)
	if img.Code != http.StatusOK || img.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("image: %d %q", img.Code, img.Header().Get("Content-Type"))
	}
	if cc := img.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("cache-control = %q", cc)
	}
	if _, err := jpeg.Decode(bytes.NewReader(img.Body.Bytes())); err != nil {
		t.Errorf("served image isn't a JPEG: %v", err)
	}
	cond := httptest.NewRequest(http.MethodGet, it.ThumbURL, nil)
	thumb := h.do(httptest.NewRequest(http.MethodGet, it.ThumbURL, nil), false)
	cond.Header.Set("If-None-Match", thumb.Header().Get("ETag"))
	if rec := h.do(cond, false); rec.Code != http.StatusNotModified {
		t.Errorf("conditional thumb: %d", rec.Code)
	}

	// Hide: gone from the public list and its image is 404 for visitors, still there for leaders.
	edit := map[string]any{"hidden": true}
	if rec := h.postJSON("/api/admin/gallery/"+itoa(it.ID)+"/visibility", edit, true); rec.Code != http.StatusOK {
		t.Fatalf("hide: %d %s", rec.Code, rec.Body.String())
	}
	if pub := h.list("/api/gallery", false); len(pub) != 1 {
		t.Errorf("hidden item still public: %d items", len(pub))
	}
	if rec := h.do(httptest.NewRequest(http.MethodGet, it.ImageURL, nil), false); rec.Code != http.StatusNotFound {
		t.Errorf("hidden image for visitor: %d", rec.Code)
	}
	leaderImg := h.do(httptest.NewRequest(http.MethodGet, it.ImageURL, nil), true)
	if leaderImg.Code != http.StatusOK || !strings.Contains(leaderImg.Header().Get("Cache-Control"), "no-store") {
		t.Errorf("hidden image for leader: %d %q", leaderImg.Code, leaderImg.Header().Get("Cache-Control"))
	}
	if all := h.list("/api/admin/gallery", true); len(all) != 2 {
		t.Errorf("admin list = %d items", len(all))
	}

	// Unhide: back on the public list, which returns every visible drawing (no paging).
	edit["hidden"] = false
	if rec := h.postJSON("/api/admin/gallery/"+itoa(it.ID)+"/visibility", edit, true); rec.Code != http.StatusOK {
		t.Fatalf("unhide: %d", rec.Code)
	}
	if pub := h.list("/api/gallery", false); len(pub) != 2 || pub[0].ID != templeID || pub[1].ID != it.ID {
		t.Errorf("unhidden list: %+v", pub)
	}
	if strings.Contains(h.do(httptest.NewRequest(http.MethodGet, "/api/gallery", nil), false).Body.String(), "featured") {
		t.Errorf("public API still exposes featured")
	}

	// Delete.
	if rec := h.postJSON("/api/admin/gallery/"+itoa(it.ID)+"/delete", nil, true); rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	if rec := h.do(httptest.NewRequest(http.MethodGet, it.ImageURL, nil), true); rec.Code != http.StatusNotFound {
		t.Errorf("deleted image: %d", rec.Code)
	}
	if rec := h.postJSON("/api/admin/gallery/"+itoa(it.ID)+"/delete", nil, true); rec.Code != http.StatusNotFound {
		t.Errorf("second delete: %d", rec.Code)
	}
}

func TestGalleryAdminEndpointsRequireLeader(t *testing.T) {
	h := newGalleryHarness(t)
	checks := []*httptest.ResponseRecorder{
		h.do(httptest.NewRequest(http.MethodGet, "/api/admin/gallery", nil), false),
		h.upload(sampleJPEG(t, 10, 10), nil, false),
		h.postJSON("/api/admin/gallery/1/visibility", map[string]any{"hidden": true}, false),
		h.postJSON("/api/admin/gallery/1/delete", nil, false),
	}
	for i, rec := range checks {
		if rec.Code != http.StatusForbidden {
			t.Errorf("check %d: status %d, want 403", i, rec.Code)
		}
	}
}

func TestGalleryUploadRejectsBadFiles(t *testing.T) {
	h := newGalleryHarness(t)
	cases := []struct {
		name string
		data []byte
		code int
	}{
		{"text", []byte("not an image at all"), http.StatusUnsupportedMediaType},
		{"heic", append([]byte{0, 0, 0, 0x18}, []byte("ftypheic\x00\x00\x00\x00mif1heic")...), http.StatusUnsupportedMediaType},
		{"too big", append([]byte{0xFF, 0xD8, 0xFF}, make([]byte, 16<<20)...), http.StatusRequestEntityTooLarge},
	}
	for _, c := range cases {
		rec := h.upload(c.data, nil, true)
		if rec.Code != c.code {
			t.Errorf("%s: %d %s, want %d", c.name, rec.Code, rec.Body.String(), c.code)
		}
	}
	if items := h.list("/api/admin/gallery", true); len(items) != 0 {
		t.Errorf("rejected uploads were stored: %d", len(items))
	}
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
