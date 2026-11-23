package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/events"
	"github.com/albertson/albertsonmissionplan/internal/mail"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

// RegisterEventRoutes adds the public events calendar and the leader event manager endpoints.
func (s *Server) RegisterEventRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/events", s.APIEvents)
	mux.HandleFunc("GET /api/events/{id}/calendar.ics", s.APIEventICS)
	mux.HandleFunc("GET /api/admin/events", s.APIAdminEvents)
	mux.HandleFunc("POST /api/admin/events", s.APIAdminEventCreate)
	mux.HandleFunc("POST /api/admin/events/{id}", s.APIAdminEventUpdate)
	mux.HandleFunc("POST /api/admin/events/{id}/delete", s.APIAdminEventDelete)
}

// eventsNow is swapped in tests.
var eventsNow = time.Now

// pastEventsWindow is how far back "Show past events" reaches.
const pastEventsWindow = 12 // months

type eventDTO struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	AllDay      bool   `json:"allDay"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Link        string `json:"link"`

	DateLabel  string `json:"dateLabel"`
	TimeLabel  string `json:"timeLabel"`
	Zone       string `json:"zone"`
	MonthKey   string `json:"monthKey"`
	MonthLabel string `json:"monthLabel"`
	Upcoming   bool   `json:"upcoming"`
	ICSURL     string `json:"icsUrl"`
	GoogleURL  string `json:"googleCalendarUrl"`
	MapURL     string `json:"mapUrl"`
}

func (s *Server) eventsPageURL() string {
	if s.PublicBaseURL == "" {
		return ""
	}
	return s.PublicBaseURL + "/events"
}

func (s *Server) toEventDTO(e store.WardEvent, loc *time.Location, today string) eventDTO {
	return eventDTO{
		ID: e.ID, Title: e.Title, Date: e.Date, AllDay: e.AllDay, StartTime: e.StartTime, EndTime: e.EndTime,
		Location: e.Location, Description: e.Description, Link: e.LinkURL,
		DateLabel:  events.DateLabel(e),
		TimeLabel:  events.TimeLabel(e),
		Zone:       events.ZoneAbbrev(e, loc),
		MonthKey:   events.MonthKey(e),
		MonthLabel: events.MonthLabel(e),
		Upcoming:   e.Date >= today,
		ICSURL:     fmt.Sprintf("/api/events/%d/calendar.ics", e.ID),
		GoogleURL:  events.GoogleCalendarURL(e, loc, s.eventsPageURL()),
		MapURL:     events.MapURL(e.Location),
	}
}

func (s *Server) eventDTOs(list []store.WardEvent, loc *time.Location, today string) []eventDTO {
	out := make([]eventDTO, 0, len(list))
	for _, e := range list {
		out = append(out, s.toEventDTO(e, loc, today))
	}
	return out
}

// APIEvents lists upcoming events (default) or, with ?view=past, the last year's past events
// (most recent first) for the public calendar.
func (s *Server) APIEvents(w http.ResponseWriter, r *http.Request) {
	loc := mail.WardLocation()
	now := eventsNow()
	today := events.Today(loc, now)
	var (
		list []store.WardEvent
		err  error
	)
	if r.URL.Query().Get("view") == "past" {
		since := now.In(loc).AddDate(0, -pastEventsWindow, 0).Format(events.DateLayout)
		list, err = store.ListPastEvents(r.Context(), s.DB, since, today)
	} else {
		list, err = store.ListUpcomingEvents(r.Context(), s.DB, today)
	}
	if err != nil {
		log.Printf("api events: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load events."})
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	writeJSON(w, http.StatusOK, map[string]any{
		"events":   s.eventDTOs(list, loc, today),
		"today":    today,
		"timeZone": loc.String(),
	})
}

func eventIDFromPath(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

// APIEventICS downloads one event as an .ics file (Apple Calendar, Outlook, Google import).
func (s *Server) APIEventICS(w http.ResponseWriter, r *http.Request) {
	id, ok := eventIDFromPath(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	e, err := store.GetEvent(r.Context(), s.DB, id)
	if errors.Is(err, store.ErrEventNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("event ics %d: %v", id, err)
		http.Error(w, "Could not load the event", http.StatusInternalServerError)
		return
	}
	host := r.Host
	if u, perr := url.Parse(s.PublicBaseURL); perr == nil && u.Host != "" {
		host = u.Host
	}
	body := events.ICS(*e, mail.WardLocation(), host, s.eventsPageURL(), eventsNow())
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="albertson-ward-event-%d.ics"`, id))
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(body)
}

// APIAdminEvents lists every event (past and upcoming) for the leader manager.
func (s *Server) APIAdminEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	list, err := store.ListAllEvents(r.Context(), s.DB)
	if err != nil {
		log.Printf("api admin events: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load events."})
		return
	}
	loc := mail.WardLocation()
	today := events.Today(loc, eventsNow())
	writeJSON(w, http.StatusOK, map[string]any{
		"events":   s.eventDTOs(list, loc, today),
		"today":    today,
		"timeZone": loc.String(),
	})
}

func writeEventInputError(w http.ResponseWriter, err error) bool {
	var ve *events.ValidationError
	if errors.As(err, &ve) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": ve.Message, "field": ve.Field})
		return true
	}
	return false
}

func (s *Server) writeSavedEvent(w http.ResponseWriter, r *http.Request, id int64) {
	e, err := store.GetEvent(r.Context(), s.DB, id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
		return
	}
	loc := mail.WardLocation()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "event": s.toEventDTO(*e, loc, events.Today(loc, eventsNow()))})
}

// APIAdminEventCreate adds an event.
func (s *Server) APIAdminEventCreate(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	var in events.Input
	if err := readJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	e, err := events.Normalize(in)
	if err != nil {
		writeEventInputError(w, err)
		return
	}
	id, err := store.InsertEvent(r.Context(), s.DB, e, email)
	if err != nil {
		log.Printf("event insert: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the event."})
		return
	}
	log.Printf("event %d created by %s", id, email)
	s.writeSavedEvent(w, r, id)
}

// APIAdminEventUpdate replaces an event's details.
func (s *Server) APIAdminEventUpdate(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := eventIDFromPath(r)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found."})
		return
	}
	var in events.Input
	if err := readJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	e, err := events.Normalize(in)
	if err != nil {
		writeEventInputError(w, err)
		return
	}
	e.ID = id
	if err := store.UpdateEvent(r.Context(), s.DB, e); err != nil {
		if errors.Is(err, store.ErrEventNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found. It may have been deleted."})
			return
		}
		log.Printf("event update %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the event."})
		return
	}
	log.Printf("event %d updated by %s", id, email)
	s.writeSavedEvent(w, r, id)
}

// APIAdminEventDelete permanently removes an event.
func (s *Server) APIAdminEventDelete(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := eventIDFromPath(r)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found."})
		return
	}
	if err := store.DeleteEvent(r.Context(), s.DB, id); err != nil {
		if errors.Is(err, store.ErrEventNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found."})
			return
		}
		log.Printf("event delete %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not delete the event."})
		return
	}
	log.Printf("event %d deleted by %s", id, email)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
