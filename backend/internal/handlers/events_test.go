package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newEventsHarness(t *testing.T) *galleryHarness {
	t.Helper()
	t.Setenv("MAIL_TIME_ZONE", "")
	prev := eventsNow
	// 9 AM on Nov 23, 2026 in Denver.
	eventsNow = func() time.Time { return time.Date(2026, 11, 23, 16, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { eventsNow = prev })
	s := newTestServer(t, nil)
	mux := http.NewServeMux()
	s.RegisterEventRoutes(mux)
	// First registration is auto-approved and signed in; keep its session cookie.
	body, _ := json.Marshal(map[string]string{"email": "lead@example.org", "password": "pw-lead-123", "firstName": "Lea", "lastName": "Der"})
	rec := httptest.NewRecorder()
	s.APIAuthRegister(rec, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}
	return &galleryHarness{t: t, mux: mux, leader: rec.Result().Cookies()}
}

type eventsResp struct {
	Events []eventDTO `json:"events"`
	Today  string     `json:"today"`
	Zone   string     `json:"timeZone"`
}

func decodeEvents(t *testing.T, rec *httptest.ResponseRecorder) eventsResp {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var out eventsResp
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestEventsRequireLeader(t *testing.T) {
	h := newEventsHarness(t)
	body := map[string]any{"title": "Fireside", "date": "2026-12-06", "startTime": "19:00"}
	for _, rec := range []*httptest.ResponseRecorder{
		h.do(httptest.NewRequest(http.MethodGet, "/api/admin/events", nil), false),
		h.postJSON("/api/admin/events", body, false),
		h.postJSON("/api/admin/events/1", body, false),
		h.postJSON("/api/admin/events/1/delete", map[string]any{}, false),
		h.uploadFlyer("/api/admin/events/1/flyer", sampleJPEG(t, 20, 20), false),
		h.postJSON("/api/admin/events/1/flyer/delete", map[string]any{}, false),
	} {
		if rec.Code != http.StatusForbidden {
			t.Fatalf("anonymous got %d: %s", rec.Code, rec.Body.String())
		}
	}
}

func TestEventsCRUDAndPublicViews(t *testing.T) {
	h := newEventsHarness(t)

	create := func(v map[string]any) eventDTO {
		t.Helper()
		rec := h.postJSON("/api/admin/events", v, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("create %v: %d %s", v, rec.Code, rec.Body.String())
		}
		var out struct {
			Event eventDTO `json:"event"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return out.Event
	}
	baptism := create(map[string]any{"title": "Baptism", "date": "2026-12-05", "startTime": "10:00", "endTime": "11:00",
		"location": "Albertson Ward Building", "description": "Everyone is welcome."})
	if baptism.ID == 0 || baptism.TimeLabel != "10:00 – 11:00 AM" || baptism.Zone != "MST" || !baptism.Upcoming ||
		baptism.MonthLabel != "December 2026" || !strings.HasPrefix(baptism.MapURL, "https://www.google.com/maps/") ||
		!strings.Contains(baptism.GoogleURL, "20261205T170000Z") {
		t.Fatalf("created dto %+v", baptism)
	}
	today := create(map[string]any{"title": "Today's open house", "date": "2026-11-23", "allDay": true})
	past := create(map[string]any{"title": "Last month's fireside", "date": "2026-10-18", "startTime": "19:00"})
	old := create(map[string]any{"title": "Two years ago", "date": "2024-10-18", "startTime": "19:00"})

	// Validation errors name the field.
	rec := h.postJSON("/api/admin/events", map[string]any{"title": "No time", "date": "2026-12-01"}, true)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"field":"startTime"`) {
		t.Fatalf("validation: %d %s", rec.Code, rec.Body.String())
	}

	up := decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/events", nil), false))
	if up.Today != "2026-11-23" || up.Zone != "America/Denver" || len(up.Events) != 2 ||
		up.Events[0].ID != today.ID || up.Events[1].ID != baptism.ID {
		t.Fatalf("upcoming %+v", up)
	}
	pastResp := decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/events?view=past", nil), false))
	if len(pastResp.Events) != 1 || pastResp.Events[0].ID != past.ID || pastResp.Events[0].Upcoming {
		t.Fatalf("past %+v (older than a year should be omitted: %d)", pastResp.Events, old.ID)
	}
	all := decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/admin/events", nil), true))
	if len(all.Events) != 4 {
		t.Fatalf("admin list %d", len(all.Events))
	}

	// Edit.
	rec = h.postJSON(fmt.Sprintf("/api/admin/events/%d", baptism.ID), map[string]any{"title": "Baptism of a new member",
		"date": "2026-12-12", "startTime": "10:00", "location": "Stake Center"}, true)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Baptism of a new member") ||
		!strings.Contains(rec.Body.String(), `"timeLabel":"10:00 AM"`) {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
	if rec := h.postJSON("/api/admin/events/999999", map[string]any{"title": "x", "date": "2026-12-12", "allDay": true}, true); rec.Code != http.StatusNotFound {
		t.Fatalf("update missing: %d", rec.Code)
	}

	// Calendar file.
	ics := h.do(httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/calendar.ics", baptism.ID), nil), false)
	if ics.Code != http.StatusOK || !strings.HasPrefix(ics.Header().Get("Content-Type"), "text/calendar") ||
		!strings.Contains(ics.Header().Get("Content-Disposition"), "attachment") ||
		!strings.Contains(ics.Body.String(), "SUMMARY:Baptism of a new member") ||
		!strings.Contains(ics.Body.String(), "DTSTART:20261212T170000Z") ||
		!strings.Contains(ics.Body.String(), "UID:ward-event-") || !strings.Contains(ics.Body.String(), "@ward.example.org") {
		t.Fatalf("ics: %d %v\n%s", ics.Code, ics.Header(), ics.Body.String())
	}
	if rec := h.do(httptest.NewRequest(http.MethodGet, "/api/events/999999/calendar.ics", nil), false); rec.Code != http.StatusNotFound {
		t.Fatalf("missing ics: %d", rec.Code)
	}

	// Delete.
	if rec := h.postJSON(fmt.Sprintf("/api/admin/events/%d/delete", baptism.ID), map[string]any{}, true); rec.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}
	if rec := h.postJSON(fmt.Sprintf("/api/admin/events/%d/delete", baptism.ID), map[string]any{}, true); rec.Code != http.StatusNotFound {
		t.Fatalf("second delete: %d", rec.Code)
	}
	up = decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/events", nil), false))
	if len(up.Events) != 1 {
		t.Fatalf("after delete %+v", up.Events)
	}
}

func (h *galleryHarness) uploadFlyer(path string, data []byte, asLeader bool) *httptest.ResponseRecorder {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("image", "flyer.jpg")
	_, _ = fw.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return h.do(req, asLeader)
}

func TestEventFlyerUploadServeDelete(t *testing.T) {
	h := newEventsHarness(t)

	// Create an event that only has a link — that link must keep working after flyer operations.
	rec := h.postJSON("/api/admin/events", map[string]any{
		"title": "Ward Christmas Party", "date": "2026-12-12", "startTime": "18:00",
		"link": "https://example.org/signup",
	}, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var created struct {
		Event eventDTO `json:"event"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	ev := created.Event
	if ev.HasFlyer || ev.FlyerURL != "" || ev.Link != "https://example.org/signup" {
		t.Fatalf("new event should have a link and no flyer: %+v", ev)
	}

	if rec := h.do(httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/events/%d/flyer/image", ev.ID), nil), false); rec.Code != http.StatusNotFound {
		t.Fatalf("no flyer yet: %d", rec.Code)
	}
	if rec := h.do(httptest.NewRequest(http.MethodGet, "/api/events/999999/flyer/image", nil), false); rec.Code != http.StatusNotFound {
		t.Fatalf("missing event flyer: %d", rec.Code)
	}

	up := h.uploadFlyer(fmt.Sprintf("/api/admin/events/%d/flyer", ev.ID), sampleJPEG(t, 1200, 1600), true)
	if up.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", up.Code, up.Body.String())
	}
	var saved struct {
		Event eventDTO `json:"event"`
	}
	_ = json.Unmarshal(up.Body.Bytes(), &saved)
	if !saved.Event.HasFlyer || saved.Event.FlyerURL == "" || saved.Event.FlyerThumbURL == "" ||
		saved.Event.FlyerWidth != 1200 || saved.Event.FlyerHeight != 1600 ||
		saved.Event.Link != "https://example.org/signup" {
		t.Fatalf("uploaded dto %+v", saved.Event)
	}

	pub := decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/events", nil), false))
	var found *eventDTO
	for i := range pub.Events {
		if pub.Events[i].ID == ev.ID {
			found = &pub.Events[i]
			break
		}
	}
	if found == nil || !found.HasFlyer || found.FlyerURL != saved.Event.FlyerURL || found.Link != "https://example.org/signup" {
		t.Fatalf("public list flyer %+v", found)
	}

	img := h.do(httptest.NewRequest(http.MethodGet, found.FlyerURL, nil), false)
	if img.Code != http.StatusOK || img.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("image: %d %q", img.Code, img.Header().Get("Content-Type"))
	}
	if cc := img.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("cache-control = %q", cc)
	}
	if _, err := jpeg.Decode(bytes.NewReader(img.Body.Bytes())); err != nil {
		t.Errorf("served image isn't a JPEG: %v", err)
	}
	thumb := h.do(httptest.NewRequest(http.MethodGet, found.FlyerThumbURL, nil), false)
	if thumb.Code != http.StatusOK {
		t.Fatalf("thumb: %d", thumb.Code)
	}
	cond := httptest.NewRequest(http.MethodGet, found.FlyerThumbURL, nil)
	cond.Header.Set("If-None-Match", thumb.Header().Get("ETag"))
	if rec := h.do(cond, false); rec.Code != http.StatusNotModified {
		t.Errorf("conditional thumb: %d", rec.Code)
	}

	// Replace the flyer: new hash/URL, same event and link.
	up2 := h.uploadFlyer(fmt.Sprintf("/api/admin/events/%d/flyer", ev.ID), sampleJPEG(t, 800, 600), true)
	if up2.Code != http.StatusOK {
		t.Fatalf("replace: %d %s", up2.Code, up2.Body.String())
	}
	_ = json.Unmarshal(up2.Body.Bytes(), &saved)
	if !saved.Event.HasFlyer || saved.Event.FlyerWidth != 800 || saved.Event.FlyerHeight != 600 ||
		saved.Event.FlyerURL == found.FlyerURL || saved.Event.Link != "https://example.org/signup" {
		t.Fatalf("replaced dto %+v (old url %s)", saved.Event, found.FlyerURL)
	}

	// A text-only update must keep the flyer.
	rec = h.postJSON(fmt.Sprintf("/api/admin/events/%d", ev.ID), map[string]any{
		"title": "Ward Christmas Party", "date": "2026-12-12", "startTime": "18:00",
		"link": "https://example.org/signup",
	}, true)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"hasFlyer":true`) {
		t.Fatalf("text update dropped flyer: %d %s", rec.Code, rec.Body.String())
	}

	// Reject a non-image; the existing flyer stays.
	bad := h.uploadFlyer(fmt.Sprintf("/api/admin/events/%d/flyer", ev.ID), []byte("not an image at all"), true)
	if bad.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("bad file: %d %s", bad.Code, bad.Body.String())
	}
	heic := append([]byte{0, 0, 0, 0x18}, []byte("ftypheic")...)
	heic = append(heic, make([]byte, 8)...)
	if rec := h.uploadFlyer(fmt.Sprintf("/api/admin/events/%d/flyer", ev.ID), heic, true); rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("heic: %d %s", rec.Code, rec.Body.String())
	}
	if rec := h.uploadFlyer(fmt.Sprintf("/api/admin/events/%d/flyer", ev.ID), append([]byte{0xFF, 0xD8, 0xFF}, make([]byte, 16<<20)...), true); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("too big: %d %s", rec.Code, rec.Body.String())
	}
	still := decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/events", nil), false))
	ok := false
	for _, e := range still.Events {
		if e.ID == ev.ID && e.HasFlyer && e.FlyerURL == saved.Event.FlyerURL {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("rejected uploads must not replace the flyer: %+v", still.Events)
	}

	// Upload to a missing event is 404.
	if rec := h.uploadFlyer("/api/admin/events/999999/flyer", sampleJPEG(t, 20, 20), true); rec.Code != http.StatusNotFound {
		t.Fatalf("upload missing: %d %s", rec.Code, rec.Body.String())
	}

	// Delete the flyer; the link remains.
	del := h.postJSON(fmt.Sprintf("/api/admin/events/%d/flyer/delete", ev.ID), map[string]any{}, true)
	if del.Code != http.StatusOK || strings.Contains(del.Body.String(), `"hasFlyer":true`) {
		t.Fatalf("delete flyer: %d %s", del.Code, del.Body.String())
	}
	if rec := h.do(httptest.NewRequest(http.MethodGet, saved.Event.FlyerURL, nil), false); rec.Code != http.StatusNotFound {
		t.Fatalf("deleted flyer still served: %d", rec.Code)
	}
	after := decodeEvents(t, h.do(httptest.NewRequest(http.MethodGet, "/api/events", nil), false))
	for _, e := range after.Events {
		if e.ID == ev.ID {
			if e.HasFlyer || e.FlyerURL != "" || e.Link != "https://example.org/signup" {
				t.Fatalf("after delete %+v", e)
			}
		}
	}
	if rec := h.postJSON("/api/admin/events/999999/flyer/delete", map[string]any{}, true); rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing: %d", rec.Code)
	}
}
