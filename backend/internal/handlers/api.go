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
	CSRFToken string   `json:"csrfToken"`
	User      *userDTO `json:"user"`
}

type userDTO struct {
	Email            string `json:"email"`
	DisplayName      string `json:"displayName,omitempty"`
	IsLeader         bool   `json:"isLeader"`
	IsAdmin          bool   `json:"isAdmin"`
	CanModerate      bool   `json:"canModerate"`
	CanEditWardGoals bool   `json:"canEditWardGoals"`
}

// APIBootstrap returns CSRF token and session info for the SPA.
//
// Model: any signed-in account with an approved leader_credentials row is a full admin who
// can moderate submissions and edit every ward plan section. If the session email no longer
// maps to an approved row (e.g. admin denied them after sign-in — currently impossible since
// approved accounts are immutable, but future-proofed), the user is treated as signed-out.
func (s *Server) APIBootstrap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email, display := s.Auth.SessionIdentity(w, r)
	email = strings.ToLower(strings.TrimSpace(email))
	display = strings.TrimSpace(display)
	var u *userDTO
	if email != "" {
		priv, err := s.loadLeaderPriv(ctx, email)
		if err != nil {
			log.Printf("api bootstrap roles: %v", err)
		} else if priv.IsAdmin {
			if display == "" {
				if first, last, lerr := store.LeaderName(ctx, s.DB, email); lerr == nil {
					display = displayName(first, last, email)
				} else {
					display = displayName("", "", email)
				}
			}
			u = &userDTO{
				Email:            email,
				DisplayName:      display,
				IsLeader:         true,
				IsAdmin:          priv.IsAdmin,
				CanModerate:      priv.canModerate(),
				CanEditWardGoals: priv.canEditWardWideGoals(),
			}
		}
	}
	writeJSON(w, http.StatusOK, bootstrapResponse{
		CSRFToken: csrf.Token(r),
		User:      u,
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
	_, err := store.InsertTestimonial(ctx, s.DB, body, author)
	if err != nil {
		log.Printf("insert testimonial: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save your experience."})
		return
	}
	go s.notifyPendingExperienceSubmission(body, author)
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

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) (email string, ok bool) {
	email = strings.ToLower(strings.TrimSpace(s.Auth.SessionEmail(w, r)))
	if email == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden."})
		return "", false
	}
	priv, err := s.loadLeaderPriv(r.Context(), email)
	if err != nil {
		log.Printf("require admin roles: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not verify access."})
		return "", false
	}
	if !priv.canModerate() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden."})
		return "", false
	}
	return email, true
}

func (s *Server) requireWardPlanEditor(w http.ResponseWriter, r *http.Request) (email string, priv leaderPriv, ok bool) {
	email = strings.ToLower(strings.TrimSpace(s.Auth.SessionEmail(w, r)))
	if email == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden."})
		return "", leaderPriv{}, false
	}
	p, err := s.loadLeaderPriv(r.Context(), email)
	if err != nil {
		log.Printf("require ward plan editor roles: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not verify access."})
		return "", leaderPriv{}, false
	}
	if !p.canUseWardPlanEditor() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden."})
		return "", leaderPriv{}, false
	}
	return email, p, true
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
	if _, ok := s.requireAdmin(w, r); !ok {
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
}

// APIAdminModerate approves or rejects a submission.
func (s *Server) APIAdminModerate(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
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
		if err := store.SetTestimonialStatus(ctx, s.DB, req.ID, "rejected", email, ""); err != nil {
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
	if _, _, ok := s.requireWardPlanEditor(w, r); !ok {
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
	email, priv, ok := s.requireWardPlanEditor(w, r)
	if !ok {
		return
	}
	var req wardPlanSaveRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	ctx := r.Context()

	orgs, err := store.ListOrgSections(ctx, s.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load organizations."})
		return
	}

	if priv.canEditWardWideGoals() {
		wg := sanitizehtml.WardBullets(req.WardGoals)
		if err := store.SetWardGoals(ctx, s.DB, wg, email); err != nil {
			log.Printf("save ward goals: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save ward goals."})
			return
		}
	}

	for _, o := range orgs {
		var body string
		if priv.canEditOrg(o.Slug) {
			body = sanitizehtml.WardBullets(req.Orgs[o.Slug])
		} else {
			body = strings.Join(o.Bullets, "\n")
		}
		if err := store.UpdateOrgBullets(ctx, s.DB, o.Slug, body); err != nil {
			log.Printf("save org %s: %v", o.Slug, err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving " + o.Slug})
			return
		}
	}
	writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Saved."})
}

type adminPublishedDTO struct {
	ID     int64  `json:"id"`
	Body   string `json:"body"`
	Author string `json:"author"`
	When   string `json:"when"`
}

type adminPublishedListDTO struct {
	Published []adminPublishedDTO `json:"published"`
}

// APIAdminListPublished returns approved experiences with their IDs so admins can edit/delete
// them. The public endpoints intentionally omit IDs.
func (s *Server) APIAdminListPublished(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	approved, err := store.ListApprovedTestimonials(r.Context(), s.DB, 500)
	if err != nil {
		log.Printf("api admin list published: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load published experiences."})
		return
	}
	out := adminPublishedListDTO{Published: make([]adminPublishedDTO, 0, len(approved))}
	for _, t := range approved {
		out.Published = append(out.Published, adminPublishedDTO{
			ID:     t.ID,
			Body:   t.Body,
			Author: strings.TrimSpace(t.AuthorLabel),
			When:   formatWhen(t.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type editPublishedRequest struct {
	ID     int64  `json:"id"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

// APIAdminEditPublished updates body / author label on an already-published experience.
func (s *Server) APIAdminEditPublished(w http.ResponseWriter, r *http.Request) {
	email, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	var req editPublishedRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	if req.ID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid id."})
		return
	}
	body := strings.TrimSpace(req.Body)
	author := strings.TrimSpace(req.Author)
	if err := store.UpdateApprovedTestimonial(r.Context(), s.DB, req.ID, body, author, email); err != nil {
		log.Printf("admin edit published %d: %v", req.ID, err)
		msg := "Could not save changes. Edit must be at least 10 characters."
		if errors.Is(err, store.ErrApproveBodyTooLong) {
			msg = "Edited text is too long (8,000 characters max)."
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}
	writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Saved."})
}

type deletePublishedRequest struct {
	ID int64 `json:"id"`
}

// APIAdminDeletePublished hard-deletes any testimonial row by id (intended for live ones, but
// works regardless of status — admins know what they're clicking).
func (s *Server) APIAdminDeletePublished(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var req deletePublishedRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	if req.ID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid id."})
		return
	}
	if err := store.DeleteTestimonial(r.Context(), s.DB, req.ID); err != nil {
		log.Printf("admin delete published %d: %v", req.ID, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not delete (may already be removed)."})
		return
	}
	writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Deleted."})
}

type pendingAccountDTO struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
	When      string `json:"when"`
}

type pendingAccountsListDTO struct {
	Pending []pendingAccountDTO `json:"pending"`
}

// APIAdminPendingAccounts lists password registrations awaiting admin approval.
func (s *Server) APIAdminPendingAccounts(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	rows, err := store.ListPendingLeaderAccounts(r.Context(), s.DB)
	if err != nil {
		log.Printf("api admin pending accounts: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load pending accounts."})
		return
	}
	out := pendingAccountsListDTO{Pending: make([]pendingAccountDTO, 0, len(rows))}
	for _, p := range rows {
		full := strings.TrimSpace(strings.TrimSpace(p.FirstName) + " " + strings.TrimSpace(p.LastName))
		out.Pending = append(out.Pending, pendingAccountDTO{
			Email:     p.Email,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			FullName:  full,
			When:      formatWhen(p.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type accountDecisionRequest struct {
	Email  string `json:"email"`
	Action string `json:"action"` // "approve" or "deny"
}

// APIAdminDecidePendingAccount approves or denies a pending password registration.
func (s *Server) APIAdminDecidePendingAccount(w http.ResponseWriter, r *http.Request) {
	approver, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	var req accountDecisionRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email is required."})
		return
	}
	ctx := r.Context()
	switch strings.TrimSpace(strings.ToLower(req.Action)) {
	case "approve":
		if err := store.ApproveLeaderAccount(ctx, s.DB, email, approver); err != nil {
			log.Printf("api approve account %s: %v", email, err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not approve (already approved or removed)."})
			return
		}
		writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Approved. They can sign in now."})
	case "deny":
		if err := store.DeletePendingLeaderAccount(ctx, s.DB, email); err != nil {
			log.Printf("api deny account %s: %v", email, err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not deny (already approved or removed)."})
			return
		}
		writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Denied. They can register again later."})
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown action."})
	}
}

type approvedLeaderDTO struct {
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	FullName     string `json:"fullName"`
	ApprovedWhen string `json:"approvedWhen"`
}

type approvedLeadersListDTO struct {
	Approved []approvedLeaderDTO `json:"approved"`
}

// APIAdminApprovedLeaders lists approved leader accounts (roster).
func (s *Server) APIAdminApprovedLeaders(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	rows, err := store.ListApprovedLeaderAccounts(r.Context(), s.DB)
	if err != nil {
		log.Printf("api admin approved leaders: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load approved leaders."})
		return
	}
	out := approvedLeadersListDTO{Approved: make([]approvedLeaderDTO, 0, len(rows))}
	for _, p := range rows {
		full := strings.TrimSpace(strings.TrimSpace(p.FirstName) + " " + strings.TrimSpace(p.LastName))
		out.Approved = append(out.Approved, approvedLeaderDTO{
			Email:        p.Email,
			FirstName:    p.FirstName,
			LastName:     p.LastName,
			FullName:     full,
			ApprovedWhen: formatWhen(p.ApprovedAt),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type removeApprovedLeaderRequest struct {
	Email string `json:"email"`
}

// APIAdminRemoveApprovedLeader deletes an approved leader credential (never the last one).
func (s *Server) APIAdminRemoveApprovedLeader(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var req removeApprovedLeaderRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email is required."})
		return
	}
	ctx := r.Context()
	if err := store.DeleteApprovedLeaderAccount(ctx, s.DB, email); err != nil {
		if errors.Is(err, store.ErrCannotRemoveLastLeader) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot remove the last approved leader. Approve someone else first."})
			return
		}
		log.Printf("api remove approved leader %s: %v", email, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Could not remove that account."})
		return
	}
	writeJSON(w, http.StatusOK, okMessage{OK: true, Message: "Leader access removed."})
}
