package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/albertson/albertsonmissionplan/internal/challenges"
	"github.com/albertson/albertsonmissionplan/internal/gallery"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

// Limits for the Monthly Challenges editor (kept generous; the card layout wraps).
const (
	maxChallengeHeading = 60
	maxChallengeGoal    = 200
	maxChallengeBullets = 8
	maxChallengeWho     = 30
	maxChallengeText    = 400
)

// RegisterChallengeRoutes adds the public and leader Monthly Challenges endpoints.
func (s *Server) RegisterChallengeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/challenges", s.APIChallenges)
	mux.HandleFunc("GET /api/challenges/{month}/{size}", s.APIChallengeImage)
	mux.HandleFunc("GET /api/admin/challenges", s.APIAdminChallenges)
	mux.HandleFunc("POST /api/admin/challenges/{month}", s.APIAdminChallengeSave)
	mux.HandleFunc("POST /api/admin/challenges/{month}/image", s.APIAdminChallengeImage)
	mux.HandleFunc("POST /api/admin/challenges/{month}/reset", s.APIAdminChallengeReset)
	mux.HandleFunc("POST /api/admin/challenges/{month}/move", s.APIAdminChallengeMove)
}

type challengeDTO struct {
	Month       int                 `json:"month"`
	Heading     string              `json:"heading"`
	Goal        string              `json:"goal"`
	Bullets     []challenges.Bullet `json:"bullets"`
	ImageURL    string              `json:"imageUrl"`
	ThumbURL    string              `json:"thumbUrl"`
	ImageWidth  int                 `json:"imageWidth"`
	ImageHeight int                 `json:"imageHeight"`
	ThumbWidth  int                 `json:"thumbWidth"`
	ThumbHeight int                 `json:"thumbHeight"`
}

type challengeOriginalDTO struct {
	Heading string              `json:"heading"`
	Goal    string              `json:"goal"`
	Bullets []challenges.Bullet `json:"bullets"`
}

type adminChallengeDTO struct {
	challengeDTO
	Hidden          bool                 `json:"hidden"`
	IsOriginalText  bool                 `json:"isOriginalText"`
	IsOriginalImage bool                 `json:"isOriginalImage"`
	UpdatedAt       string               `json:"updatedAt,omitempty"`
	UpdatedBy       string               `json:"updatedBy,omitempty"`
	Original        challengeOriginalDTO `json:"original"`
}

func toChallengeDTO(c store.MonthlyChallenge) challengeDTO {
	b := c.Bullets
	if b == nil {
		b = []challenges.Bullet{}
	}
	return challengeDTO{
		Month: c.Month, Heading: c.Heading, Goal: c.Goal, Bullets: b,
		ImageURL:   fmt.Sprintf("/api/challenges/%d/image?v=%s", c.Month, c.ImageHash),
		ThumbURL:   fmt.Sprintf("/api/challenges/%d/thumb?v=%s", c.Month, c.ImageHash),
		ImageWidth: challenges.ImageWidth, ImageHeight: challenges.ImageHeight,
		ThumbWidth: challenges.ThumbWidth, ThumbHeight: challenges.ThumbHeight,
	}
}

func toAdminChallengeDTO(c store.MonthlyChallenge) adminChallengeDTO {
	d, _ := challenges.ForMonth(c.Month)
	out := adminChallengeDTO{
		challengeDTO:    toChallengeDTO(c),
		Hidden:          c.Hidden,
		IsOriginalImage: store.IsDefaultArt(c.Month, c.ImageHash),
		Original:        challengeOriginalDTO{Heading: d.Heading, Goal: d.Goal, Bullets: d.Bullets},
		UpdatedBy:       c.UpdatedBy,
	}
	out.IsOriginalText = c.Heading == d.Heading && c.Goal == d.Goal && reflect.DeepEqual(out.Bullets, d.Bullets)
	if c.UpdatedBy != "" && !c.UpdatedAt.IsZero() {
		out.UpdatedAt = c.UpdatedAt.In(time.Local).Format("January 2, 2006")
	}
	return out
}

// APIChallenges lists the published challenges (January → December) for the public page and
// the printable sheet.
func (s *Server) APIChallenges(w http.ResponseWriter, r *http.Request) {
	list, err := store.ListMonthlyChallenges(r.Context(), s.DB, false)
	if err != nil {
		log.Printf("api challenges: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load the monthly challenges."})
		return
	}
	out := make([]challengeDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toChallengeDTO(c))
	}
	w.Header().Set("Cache-Control", "no-cache")
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

// APIChallengeImage serves a month's picture from the database. URLs carry the content hash
// (?v=), so matching requests are cached for a year; hidden months are leader-only.
func (s *Server) APIChallengeImage(w http.ResponseWriter, r *http.Request) {
	month, err := strconv.Atoi(r.PathValue("month"))
	size := r.PathValue("size")
	if err != nil || month < 1 || month > 12 || (size != "image" && size != "thumb") {
		http.NotFound(w, r)
		return
	}
	data, ctype, hash, hidden, err := store.MonthlyChallengeImage(r.Context(), s.DB, month, size == "thumb")
	if errors.Is(err, store.ErrChallengeNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("challenge image %d: %v", month, err)
		http.Error(w, "Could not load image", http.StatusInternalServerError)
		return
	}
	if hidden && !s.isModerator(w, r) {
		http.NotFound(w, r)
		return
	}
	etag := `"` + hash + "-" + size + `"`
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", etag)
	switch {
	case hidden:
		w.Header().Set("Cache-Control", "private, no-store")
	case r.URL.Query().Get("v") == hash:
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	default:
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag && !hidden {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func (s *Server) adminChallengesResponse(w http.ResponseWriter, r *http.Request, status int, extra map[string]any) {
	list, err := store.ListMonthlyChallenges(r.Context(), s.DB, true)
	if err != nil {
		log.Printf("api admin challenges: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load the monthly challenges."})
		return
	}
	out := make([]adminChallengeDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toAdminChallengeDTO(c))
	}
	body := map[string]any{"items": out}
	for k, v := range extra {
		body[k] = v
	}
	writeJSON(w, status, body)
}

// APIAdminChallenges lists all twelve months (including hidden ones) for the editor.
func (s *Server) APIAdminChallenges(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	s.adminChallengesResponse(w, r, http.StatusOK, nil)
}

func challengeMonthFromPath(w http.ResponseWriter, r *http.Request) (int, bool) {
	m, err := strconv.Atoi(r.PathValue("month"))
	if err != nil || m < 1 || m > 12 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "That month doesn't exist."})
		return 0, false
	}
	return m, true
}

type challengeSaveRequest struct {
	Heading string              `json:"heading"`
	Goal    string              `json:"goal"`
	Bullets []challenges.Bullet `json:"bullets"`
	Hidden  bool                `json:"hidden"`
}

// oneLine trims a field and folds any line breaks or runs of spaces into single spaces.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// validateChallenge cleans up the submitted text and returns a friendly message when a field
// is missing or too long.
func validateChallenge(req challengeSaveRequest) (store.ChallengeText, string) {
	t := store.ChallengeText{Heading: oneLine(req.Heading), Goal: oneLine(req.Goal), Hidden: req.Hidden}
	switch {
	case t.Heading == "":
		return t, "Give the challenge a name, like “Gratitude Challenge”."
	case utf8.RuneCountInString(t.Heading) > maxChallengeHeading:
		return t, fmt.Sprintf("The challenge name can be up to %d characters.", maxChallengeHeading)
	case t.Goal == "":
		return t, "Add the goal for this month."
	case utf8.RuneCountInString(t.Goal) > maxChallengeGoal:
		return t, fmt.Sprintf("The goal can be up to %d characters.", maxChallengeGoal)
	}
	for _, b := range req.Bullets {
		b.Who, b.Text = oneLine(b.Who), oneLine(b.Text)
		if b.Who == "" && b.Text == "" {
			continue // an empty row the leader added and never filled in
		}
		switch {
		case b.Who == "":
			return t, "Each idea needs a “who” label, like Individuals, Families, or Ward."
		case utf8.RuneCountInString(b.Who) > maxChallengeWho:
			return t, fmt.Sprintf("“Who” labels can be up to %d characters.", maxChallengeWho)
		case b.Text == "":
			return t, fmt.Sprintf("The idea for %s is empty. Fill it in or remove that row.", b.Who)
		case utf8.RuneCountInString(b.Text) > maxChallengeText:
			return t, fmt.Sprintf("Each idea can be up to %d characters.", maxChallengeText)
		}
		t.Bullets = append(t.Bullets, b)
	}
	if len(t.Bullets) == 0 {
		return t, "Add at least one idea (for example, one for Individuals)."
	}
	if len(t.Bullets) > maxChallengeBullets {
		return t, fmt.Sprintf("A challenge can have up to %d ideas.", maxChallengeBullets)
	}
	return t, ""
}

// APIAdminChallengeSave saves one month's text and visibility.
func (s *Server) APIAdminChallengeSave(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	month, ok := challengeMonthFromPath(w, r)
	if !ok {
		return
	}
	var req challengeSaveRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	t, msg := validateChallenge(req)
	if msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}
	if err := store.UpdateMonthlyChallengeText(r.Context(), s.DB, month, t, email); err != nil {
		if errors.Is(err, store.ErrChallengeNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "That month doesn't exist."})
			return
		}
		log.Printf("challenge save %d: %v", month, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save."})
		return
	}
	s.adminChallengesResponse(w, r, http.StatusOK, map[string]any{"ok": true})
}

// APIAdminChallengeImage replaces one month's picture (multipart field "image"). The upload is
// center-cropped to the card's 3:2 shape, downscaled, and re-encoded (EXIF/GPS dropped).
func (s *Server) APIAdminChallengeImage(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	month, ok := challengeMonthFromPath(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, gallery.MaxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": gallery.ErrTooLarge.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not read the upload."})
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	f, _, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Choose a picture to upload."})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, gallery.MaxUploadBytes+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not read the upload."})
		return
	}

	galleryProcessSlot <- struct{}{}
	res, err := gallery.ProcessCover(data, challenges.ImageWidth, challenges.ImageHeight,
		challenges.ThumbWidth, challenges.ThumbHeight)
	<-galleryProcessSlot
	if err != nil {
		status := http.StatusUnsupportedMediaType
		switch {
		case errors.Is(err, gallery.ErrTooLarge):
			status = http.StatusRequestEntityTooLarge
		case errors.Is(err, gallery.ErrTooManyPx), errors.Is(err, gallery.ErrCorrupt):
			status = http.StatusUnprocessableEntity
		case errors.Is(err, gallery.ErrHEIC), errors.Is(err, gallery.ErrUnsupported):
		default:
			log.Printf("challenge image process: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not process the picture."})
			return
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	art := challenges.Art{ContentType: "image/jpeg", Image: res.Display, Thumb: res.Thumb}
	if err := store.SetMonthlyChallengeImage(r.Context(), s.DB, month, art, true, email); err != nil {
		log.Printf("challenge image save %d: %v", month, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the picture."})
		return
	}
	log.Printf("challenge %d picture replaced by %s (%d KB in → %d KB)", month, email, len(data)>>10, len(res.Display)>>10)
	s.adminChallengesResponse(w, r, http.StatusOK, map[string]any{"ok": true})
}

type challengeResetRequest struct {
	// Part is "all" (default), "text", or "image".
	Part string `json:"part"`
}

// APIAdminChallengeReset restores a month's original text and/or picture.
func (s *Server) APIAdminChallengeReset(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	month, ok := challengeMonthFromPath(w, r)
	if !ok {
		return
	}
	var req challengeResetRequest
	if err := readJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	ctx := r.Context()
	var err error
	switch req.Part {
	case "", "all":
		err = store.ResetMonthlyChallenge(ctx, s.DB, month, email)
	case "text":
		d, _ := challenges.ForMonth(month)
		var cur *store.MonthlyChallenge
		cur, err = store.GetMonthlyChallenge(ctx, s.DB, month)
		if err == nil {
			err = store.UpdateMonthlyChallengeText(ctx, s.DB, month,
				store.ChallengeText{Heading: d.Heading, Goal: d.Goal, Bullets: d.Bullets, Hidden: cur.Hidden}, email)
		}
	case "image":
		var art challenges.Art
		art, err = challenges.DefaultArt(month)
		if err == nil {
			err = store.SetMonthlyChallengeImage(ctx, s.DB, month, art, false, email)
		}
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request."})
		return
	}
	if err != nil {
		log.Printf("challenge reset %d: %v", month, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not reset this month."})
		return
	}
	s.adminChallengesResponse(w, r, http.StatusOK, map[string]any{"ok": true})
}

type challengeMoveRequest struct {
	To int `json:"to"`
}

// APIAdminChallengeMove swaps two months' challenges: {"to": 11} moves this month's challenge
// to November and November's challenge here.
func (s *Server) APIAdminChallengeMove(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	month, ok := challengeMonthFromPath(w, r)
	if !ok {
		return
	}
	var req challengeMoveRequest
	if err := readJSON(r, &req); err != nil || req.To < 1 || req.To > 12 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Choose a month to move this challenge to."})
		return
	}
	if err := store.SwapMonthlyChallenges(r.Context(), s.DB, month, req.To, email); err != nil {
		log.Printf("challenge move %d→%d: %v", month, req.To, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not move the challenge."})
		return
	}
	s.adminChallengesResponse(w, r, http.StatusOK, map[string]any{"ok": true})
}
