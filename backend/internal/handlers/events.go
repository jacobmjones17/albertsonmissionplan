package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/events"
	"github.com/albertson/albertsonmissionplan/internal/gallery"
	"github.com/albertson/albertsonmissionplan/internal/mail"
	"github.com/albertson/albertsonmissionplan/internal/spa"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

// RegisterEventRoutes adds the public events calendar and the leader event manager endpoints.
func (s *Server) RegisterEventRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /events/{id}", s.ServeEventPage)
	mux.HandleFunc("GET /api/events", s.APIEvents)
	mux.HandleFunc("GET /api/events/{id}/calendar.ics", s.APIEventICS)
	mux.HandleFunc("GET /api/events/{id}/flyer/{size}", s.APIEventFlyer)
	mux.HandleFunc("GET /api/admin/events", s.APIAdminEvents)
	mux.HandleFunc("POST /api/admin/events", s.APIAdminEventCreate)
	mux.HandleFunc("POST /api/admin/events/{id}", s.APIAdminEventUpdate)
	mux.HandleFunc("POST /api/admin/events/{id}/delete", s.APIAdminEventDelete)
	mux.HandleFunc("POST /api/admin/events/{id}/flyer", s.APIAdminEventFlyerUpload)
	mux.HandleFunc("POST /api/admin/events/{id}/flyer/delete", s.APIAdminEventFlyerDelete)
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

	HasFlyer      bool   `json:"hasFlyer"`
	FlyerURL      string `json:"flyerUrl,omitempty"`
	FlyerThumbURL string `json:"flyerThumbUrl,omitempty"`
	FlyerWidth    int    `json:"flyerWidth,omitempty"`
	FlyerHeight   int    `json:"flyerHeight,omitempty"`

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
	return s.trimmedPublicBaseURL() + "/events"
}

func (s *Server) toEventDTO(e store.WardEvent, loc *time.Location, today string) eventDTO {
	d := eventDTO{
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
	if e.HasFlyer() {
		d.HasFlyer = true
		d.FlyerURL = fmt.Sprintf("/api/events/%d/flyer/image?v=%s", e.ID, e.FlyerHash)
		d.FlyerThumbURL = fmt.Sprintf("/api/events/%d/flyer/thumb?v=%s", e.ID, e.FlyerHash)
		d.FlyerWidth = e.FlyerWidth
		d.FlyerHeight = e.FlyerHeight
	}
	return d
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
	host := "albertsonwardplan.com"
	if u, perr := url.Parse(s.trimmedPublicBaseURL()); perr == nil && u.Hostname() != "" {
		host = u.Hostname()
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

// APIEventFlyer serves an event's uploaded flyer (display image or thumbnail) from the database.
// URLs carry the content hash (?v=), so matching requests are cached for a year.
func (s *Server) APIEventFlyer(w http.ResponseWriter, r *http.Request) {
	id, ok := eventIDFromPath(r)
	size := r.PathValue("size")
	if !ok || (size != "image" && size != "thumb") {
		http.NotFound(w, r)
		return
	}
	data, hash, err := store.EventFlyerImage(r.Context(), s.DB, id, size == "thumb")
	if errors.Is(err, store.ErrEventNotFound) || errors.Is(err, store.ErrEventFlyerNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("event flyer %d: %v", id, err)
		http.Error(w, "Could not load the flyer", http.StatusInternalServerError)
		return
	}
	etag := `"` + hash + "-" + size + `"`
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", etag)
	if r.URL.Query().Get("v") == hash {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func writeGalleryProcessError(w http.ResponseWriter, err error) bool {
	status := http.StatusUnsupportedMediaType
	switch {
	case errors.Is(err, gallery.ErrTooLarge):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, gallery.ErrTooManyPx), errors.Is(err, gallery.ErrCorrupt):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, gallery.ErrHEIC), errors.Is(err, gallery.ErrUnsupported):
	default:
		return false
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
	return true
}

func readMultipartImage(w http.ResponseWriter, r *http.Request) (data []byte, ok bool) {
	r.Body = http.MaxBytesReader(w, r.Body, gallery.MaxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": gallery.ErrTooLarge.Error()})
			return nil, false
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not read the upload."})
		return nil, false
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	f, _, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Choose a picture to upload."})
		return nil, false
	}
	defer f.Close()
	data, err = io.ReadAll(io.LimitReader(f, gallery.MaxUploadBytes+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not read the upload."})
		return nil, false
	}
	return data, true
}

// APIAdminEventFlyerUpload stores a flyer photo on an event (multipart field "image").
// The file is validated and processed the same way as Ward Gallery drawings (15 MB max,
// JPEG/PNG/WebP, EXIF/GPS dropped, display + thumbnail JPEGs).
func (s *Server) APIAdminEventFlyerUpload(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := eventIDFromPath(r)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found."})
		return
	}
	data, ok := readMultipartImage(w, r)
	if !ok {
		return
	}

	galleryProcessSlot <- struct{}{}
	start := time.Now()
	res, err := gallery.Process(data)
	<-galleryProcessSlot
	if err != nil {
		if writeGalleryProcessError(w, err) {
			return
		}
		log.Printf("event flyer process %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process the flyer."})
		return
	}
	if err := store.SetEventFlyer(r.Context(), s.DB, id, store.GalleryImages{
		Width: res.DisplayWidth, Height: res.DisplayHeight,
		ThumbWidth: res.ThumbWidth, ThumbHeight: res.ThumbHeight,
		Hash: res.Hash, Image: res.Display, Thumb: res.Thumb,
	}); err != nil {
		if errors.Is(err, store.ErrEventNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found. It may have been deleted."})
			return
		}
		log.Printf("event flyer save %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the flyer."})
		return
	}
	log.Printf("event %d flyer uploaded by %s: %d KB in → %dx%d (%d KB) + thumb (%d KB) in %s", id, email,
		len(data)>>10, res.DisplayWidth, res.DisplayHeight, len(res.Display)>>10, len(res.Thumb)>>10,
		time.Since(start).Round(time.Millisecond))
	s.writeSavedEvent(w, r, id)
}

// APIAdminEventFlyerDelete removes an uploaded flyer from an event. The optional link is left alone.
func (s *Server) APIAdminEventFlyerDelete(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	id, ok := eventIDFromPath(r)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found."})
		return
	}
	if err := store.ClearEventFlyer(r.Context(), s.DB, id); err != nil {
		if errors.Is(err, store.ErrEventNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Event not found."})
			return
		}
		log.Printf("event flyer delete %d: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not remove the flyer."})
		return
	}
	log.Printf("event %d flyer removed by %s", id, email)
	s.writeSavedEvent(w, r, id)
}

// ServeEventPage serves the events SPA for a per-event share URL (/events/{id}) and
// rewrites Open Graph tags so iMessage and social previews show this event.
func (s *Server) ServeEventPage(w http.ResponseWriter, r *http.Request) {
	id, ok := spa.EventSharePath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	html := s.IndexHTML
	if len(html) == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	status := http.StatusOK
	e, err := store.GetEvent(r.Context(), s.DB, id)
	switch {
	case errors.Is(err, store.ErrEventNotFound):
		status = http.StatusNotFound
	case err != nil:
		log.Printf("event page %d: %v", id, err)
		// Still serve the app shell so a visitor can use the calendar.
	default:
		loc := mail.WardLocation()
		today := events.Today(loc, eventsNow())
		d := s.toEventDTO(*e, loc, today)
		pageURL := s.absoluteURL(fmt.Sprintf("/events/%d", id))
		og := spa.OpenGraph{
			Title:       d.Title + " | Albertson Ward",
			Description: eventShareText(*e),
			URL:         pageURL,
			Image:       s.absoluteURL("/site/og-share.jpg"),
			ImageAlt:    d.Title,
			ImageWidth:  1200,
			ImageHeight: 1600,
		}
		if d.HasFlyer && d.FlyerURL != "" {
			og.Image = s.absoluteURL(d.FlyerURL)
			og.ImageAlt = "Flyer for " + d.Title
			if d.FlyerWidth > 0 {
				og.ImageWidth = d.FlyerWidth
			}
			if d.FlyerHeight > 0 {
				og.ImageHeight = d.FlyerHeight
			}
		}
		html = spa.ApplyOpenGraph(html, og)
	}
	spa.ServeIndexHTML(w, html, status)
}

// eventShareText is the short blurb used in the Web Share sheet and Open Graph description.
func eventShareText(e store.WardEvent) string {
	parts := []string{events.DateLabel(e)}
	if t := events.TimeLabel(e); t != "" {
		parts = append(parts, t)
	}
	if loc := strings.TrimSpace(e.Location); loc != "" {
		parts = append(parts, loc)
	}
	return strings.Join(parts, " · ")
}

func (s *Server) absoluteURL(p string) string {
	base := s.trimmedPublicBaseURL()
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return base + p
}
