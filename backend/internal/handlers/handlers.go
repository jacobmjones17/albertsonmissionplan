package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"golang.org/x/oauth2"
)

// Server wires HTTP handlers to storage and auth.
type Server struct {
	DB   *sql.DB
	Auth *auth.Config
}

func formatWhen(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(time.Local).Format("January 2, 2006")
}

// --- Google OAuth ---

func (s *Server) GoogleStart(w http.ResponseWriter, r *http.Request) {
	if s.Auth.OAuth == nil {
		http.Error(w, "Google sign-in is not configured (set OAuth env vars or unset DEV_SKIP_OAUTH).", http.StatusServiceUnavailable)
		return
	}
	st, err := auth.RandomState()
	if err != nil {
		log.Printf("auth google start: random state: %v", err)
		http.Error(w, "Auth error", http.StatusInternalServerError)
		return
	}
	if err := s.Auth.SetOAuthState(w, r, st); err != nil {
		log.Printf("auth google start: session: %v", err)
		http.Error(w, "Auth error", http.StatusInternalServerError)
		return
	}
	authURL := s.Auth.OAuth.AuthCodeURL(st, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if s.Auth.OAuth == nil {
		http.Error(w, "Google sign-in is not configured.", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		http.Redirect(w, r, "/admin?msg="+url.QueryEscape(errMsg), http.StatusSeeOther)
		return
	}
	state := r.URL.Query().Get("state")
	ok, err := s.Auth.ValidOAuthState(w, r, state)
	if err != nil || !ok {
		http.Redirect(w, r, "/admin?msg="+url.QueryEscape("Invalid sign-in state. Try again."), http.StatusSeeOther)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/admin?msg="+url.QueryEscape("Missing authorization code."), http.StatusSeeOther)
		return
	}
	tok, err := s.Auth.OAuth.Exchange(ctx, code)
	if err != nil {
		log.Printf("oauth exchange: %v", err)
		http.Redirect(w, r, "/admin?msg="+url.QueryEscape("Could not complete sign-in."), http.StatusSeeOther)
		return
	}
	client := s.Auth.OAuth.Client(ctx, tok)
	email, name, err := auth.FetchGoogleEmail(ctx, client)
	if err != nil || email == "" {
		log.Printf("google userinfo: %v", err)
		http.Redirect(w, r, "/admin?msg="+url.QueryEscape("Could not read your Google profile."), http.StatusSeeOther)
		return
	}
	if !s.Auth.IsLeader(email) {
		_ = s.Auth.ClearSession(w, r)
		q := url.Values{}
		q.Set("msg", "That Google account is not on the leader list. Add the email below to ALLOWED_LEADER_EMAILS in .env and restart the server.")
		q.Set("email", email)
		http.Redirect(w, r, "/admin?"+q.Encode(), http.StatusSeeOther)
		return
	}
	if err := s.Auth.SetUserSession(w, r, email, name); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	_ = s.Auth.ClearSession(w, r)
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
