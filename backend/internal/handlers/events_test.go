package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
