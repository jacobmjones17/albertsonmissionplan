package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

type authCredentialRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func displayNameFromEmail(email string) string {
	e := strings.TrimSpace(strings.ToLower(email))
	i := strings.IndexByte(e, '@')
	if i <= 0 || i >= len(e)-1 {
		return ""
	}
	return strings.ReplaceAll(e[:i], ".", " ")
}

// displayName builds a "First Last" string, trimming and skipping empties; falls back to a
// best-effort name derived from the local part of the email so the navbar / admin page never
// shows an empty greeting for legacy accounts (e.g. seeded via the CLI before names existed).
func displayName(first, last, email string) string {
	parts := []string{strings.TrimSpace(first), strings.TrimSpace(last)}
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if out == "" {
			out = p
		} else {
			out += " " + p
		}
	}
	if out != "" {
		return out
	}
	return displayNameFromEmail(email)
}

func (s *Server) APIAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req authCredentialRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please enter a valid email address."})
		return
	}
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	if firstName == "" || lastName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please enter your first and last name."})
		return
	}
	const nameMax = 80
	if len(firstName) > nameMax || len(lastName) > nameMax {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Names are limited to 80 characters each."})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	switch {
	case errors.Is(err, auth.ErrPasswordEmpty):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please enter a password."})
		return
	case errors.Is(err, auth.ErrPasswordTooLong):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is too long (max 72 characters — bcrypt's hard limit)."})
		return
	case err != nil:
		log.Printf("register hash password: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not create account."})
		return
	}

	// First-ever registration auto-approves so the system can bootstrap with no admin yet.
	hasApproved, err := store.HasApprovedLeaderAccount(ctx, s.DB)
	if err != nil {
		log.Printf("register check approved: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not create account."})
		return
	}
	autoApprovedBy := ""
	if !hasApproved {
		autoApprovedBy = "system-bootstrap"
	}
	if err := store.InsertLeaderCredential(ctx, s.DB, email, firstName, lastName, hash, autoApprovedBy); err != nil {
		if errors.Is(err, store.ErrLeaderAlreadyRegistered) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "That email already has an account. Sign in instead."})
			return
		}
		log.Printf("register insert credential: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not create account."})
		return
	}
	if autoApprovedBy != "" {
		name := displayName(firstName, lastName, email)
		if err := s.Auth.SetUserSession(w, r, email, name); err != nil {
			log.Printf("register session: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Session error."})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"status":  "approved",
			"message": "Account created and signed in (first leader auto-approved).",
		})
		return
	}
	go s.notifyPendingLeaderSignup(email, firstName, lastName)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"status":  "pending",
		"message": "Account request submitted. An existing leader must approve it before you can sign in.",
	})
}

func (s *Server) APIAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req authCredentialRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body."})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password."})
		return
	}

	hash, err := store.LeaderCredentialHash(ctx, s.DB, email)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password."})
		return
	}
	if errors.Is(err, store.ErrLeaderAccountPending) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Your account is still awaiting admin approval. Once an existing leader approves it, you can sign in."})
		return
	}
	if err != nil {
		log.Printf("login fetch credential: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not sign in."})
		return
	}
	if !auth.CheckPassword(req.Password, hash) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password."})
		return
	}
	firstName, lastName, nameErr := store.LeaderName(ctx, s.DB, email)
	if nameErr != nil && !errors.Is(nameErr, sql.ErrNoRows) {
		log.Printf("login fetch name: %v", nameErr)
	}
	name := displayName(firstName, lastName, email)
	if err := s.Auth.SetUserSession(w, r, email, name); err != nil {
		log.Printf("login session: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Session error."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
