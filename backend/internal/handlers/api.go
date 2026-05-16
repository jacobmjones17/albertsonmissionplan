package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"

	"github.com/albertson/albertsonmissionplan/internal/sanitizehtml"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<22)) // 4 MiB
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

type bootstrapResponse struct {
	CSRFToken      string   `json:"csrfToken"`
	User           *userDTO `json:"user"`
	GoogleSignInOK bool     `json:"googleSignInOk"`
}

type userDTO struct {
	Email    string `json:"email"`
	IsLeader bool   `json:"isLeader"`
}

// APIBootstrap returns CSRF token and session info for the SPA.
func (s *Server) APIBootstrap(w http.ResponseWriter, r *http.Request) {
	email := s.Auth.SessionEmail(r)
	var u *userDTO
	if email != "" {
		u = &userDTO{Email: email, IsLeader: s.Auth.IsLeader(email)}
	}
	writeJSON(w, http.StatusOK, bootstrapResponse{
		CSRFToken:      csrf.Token(r),
		User:           u,
		GoogleSignInOK: s.Auth.OAuth != nil,
	})
}

type experiencesDTO struct {
	Experiences []experienceDTO `json:"experiences"`
}

type experienceDTO struct {
	Body   string `json:"body"`
	Author string `json:"author"`
	When   string `json:"when,omitempty"`
}

func mapApprovedDTO(in []store.Testimonial) []experienceDTO {
	out := make([]experienceDTO, 0, len(in))
	for _, t := range in {
		who := strings.TrimSpace(t.AuthorLabel)
		if who == "" {
			who = "A ward member"
		}
		out = append(out, experienceDTO{
			Body:   t.Body,
			Author: who,
			When:   formatWhen(t.CreatedAt),
		})
	}
	return out
}

// APIHomeExperiences returns recent approved experiences for the home page.
func (s *Server) APIHomeExperiences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	approved, err := store.ListApprovedTestimonials(ctx, s.DB, 5)
	if err != nil {
		log.Printf("api home testimonials: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load experiences."})
		return
	}
	writeJSON(w, http.StatusOK, experiencesDTO{Experiences: mapApprovedDTO(approved)})
}

// APIExperiences lists approved experiences.
func (s *Server) APIExperiences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	approved, err := store.ListApprovedTestimonials(ctx, s.DB, 100)
	if err != nil {
		log.Printf("api experiences list: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load experiences."})
		return
	}
	writeJSON(w, http.StatusOK, experiencesDTO{Experiences: mapApprovedDTO(approved)})
}

type postExperienceRequest struct {
	Body    string `json:"body"`
	Author  string `json:"author"`
	Website string `json:"website"` // honeypot
}

type okMessage struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// APIPostExperience accepts a new experience (JSON body).
func (s *Server) APIPostExperience(w http.ResponseWriter, r *http.Request) {
	var req postExperienceRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	if strings.TrimSpace(req.Website) != "" {
		writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Thanks for sharing!"})
		return
	}
	body := strings.TrimSpace(req.Body)
	author := strings.TrimSpace(req.Author)
	switch {
	case len(body) < 20:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please write a bit more (at least 20 characters)."})
		return
	case len(body) > 8000:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please shorten your experience to 8,000 characters or less."})
		return
	}
	ctx := r.Context()
	if _, err := store.InsertTestimonial(ctx, s.DB, body, author); err != nil {
		log.Printf("insert testimonial: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save your experience."})
		return
	}
	writeJSON(w, http.StatusOK, okMessage{
		OK:      true,
		Message: "Thank you! Leaders will review your experience before it appears here.",
	})
}

type orgDTO struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Bullets []string `json:"bullets"`
}

type wardPlanDTO struct {
	WardGoals []string `json:"wardGoals"`
	Orgs      []orgDTO `json:"orgs"`
}

// APIWardPlan returns public ward plan data.
func (s *Server) APIWardPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgs, err := store.ListOrgSections(ctx, s.DB)
	if err != nil {
		log.Printf("api ward plan orgs: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load ward plan."})
		return
	}
	goals, err := store.GetWardGoals(ctx, s.DB)
	if err != nil {
		log.Printf("api ward goals: %v", err)
	}
	out := wardPlanDTO{WardGoals: goals, Orgs: make([]orgDTO, 0, len(orgs))}
	for _, o := range orgs {
		out.Orgs = append(out.Orgs, orgDTO{Slug: o.Slug, Title: o.Title, Bullets: o.Bullets})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) requireLeader(w http.ResponseWriter, r *http.Request) (email string, ok bool) {
	email = s.Auth.SessionEmail(r)
	if email == "" || !s.Auth.IsLeader(email) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden."})
		return "", false
	}
	return email, true
}

type pendingDTO struct {
	ID     int64  `json:"id"`
	Body   string `json:"body"`
	Author string `json:"author"`
	When   string `json:"when"`
}

type pendingListDTO struct {
	Pending []pendingDTO `json:"pending"`
}

// APIAdminPending lists testimonial queue.
func (s *Server) APIAdminPending(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireLeader(w, r); !ok {
		return
	}
	ctx := r.Context()
	pending, err := store.ListPendingTestimonials(ctx, s.DB)
	if err != nil {
		log.Printf("api pending: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load queue."})
		return
	}
	out := pendingListDTO{Pending: make([]pendingDTO, 0, len(pending))}
	for _, t := range pending {
		out.Pending = append(out.Pending, pendingDTO{
			ID:     t.ID,
			Body:   t.Body,
			Author: strings.TrimSpace(t.AuthorLabel),
			When:   formatWhen(t.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type moderateRequest struct {
	ID     int64  `json:"id"`
	Action string `json:"action"`
	Body   string `json:"body"`
	Author string `json:"author"`
	Note   string `json:"note"`
}

// APIAdminModerate approves or rejects a submission.
func (s *Server) APIAdminModerate(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireLeader(w, r)
	if !ok {
		return
	}
	var req moderateRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	if req.ID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid id."})
		return
	}
	ctx := r.Context()
	switch strings.TrimSpace(strings.ToLower(req.Action)) {
	case "approve":
		body := strings.TrimSpace(req.Body)
		author := strings.TrimSpace(req.Author)
		if err := store.ApproveTestimonial(ctx, s.DB, req.ID, body, author, email); err != nil {
			log.Printf("moderate approve: %v", err)
			msg := "Could not publish. Edit the text so it is at least a short sentence (10+ characters) and try again."
			if errors.Is(err, store.ErrApproveBodyTooLong) {
				msg = "Edited text is too long (8,000 characters max)."
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
			return
		}
		writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Published."})
	case "reject":
		note := strings.TrimSpace(req.Note)
		if note == "" {
			note = "Does not meet ward publishing guidelines."
		}
		if err := store.SetTestimonialStatus(ctx, s.DB, req.ID, "rejected", email, note); err != nil {
			log.Printf("moderate reject: %v", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not update that submission (it may already be handled)."})
			return
		}
		writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Rejected."})
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown action."})
	}
}

// APIAdminWardPlan returns editable ward plan snapshot.
func (s *Server) APIAdminWardPlan(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireLeader(w, r); !ok {
		return
	}
	ctx := r.Context()
	orgs, err := store.ListOrgSections(ctx, s.DB)
	if err != nil {
		log.Printf("api admin orgs: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load organizations."})
		return
	}
	goals, err := store.GetWardGoals(ctx, s.DB)
	if err != nil {
		log.Printf("api admin ward goals: %v", err)
	}
	out := wardPlanDTO{WardGoals: goals, Orgs: make([]orgDTO, 0, len(orgs))}
	for _, o := range orgs {
		out.Orgs = append(out.Orgs, orgDTO{Slug: o.Slug, Title: o.Title, Bullets: o.Bullets})
	}
	writeJSON(w, http.StatusOK, out)
}

type wardPlanSaveRequest struct {
	WardGoals string            `json:"wardGoals"`
	Orgs      map[string]string `json:"orgs"` // slug -> multiline bullets
}

// APIAdminWardPlanSave persists ward plan edits.
func (s *Server) APIAdminWardPlanSave(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireLeader(w, r)
	if !ok {
		return
	}
	var req wardPlanSaveRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	ctx := r.Context()

	req.WardGoals = sanitizehtml.WardBullets(req.WardGoals)

	if err := store.SetWardGoals(ctx, s.DB, req.WardGoals, email); err != nil {
		log.Printf("save ward goals: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save ward goals."})
		return
	}

	orgs, err := store.ListOrgSections(ctx, s.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not reload organizations."})
		return
	}
	for _, o := range orgs {
		body := sanitizehtml.WardBullets(req.Orgs[o.Slug])
		if err := store.UpdateOrgBullets(ctx, s.DB, o.Slug, body); err != nil {
			log.Printf("save org %s: %v", o.Slug, err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving " + o.Slug})
			return
		}
	}
	writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Saved."})
}
