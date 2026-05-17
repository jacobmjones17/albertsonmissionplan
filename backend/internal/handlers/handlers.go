package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/mail"
)

// Server wires HTTP handlers to storage and auth.
type Server struct {
	DB            *sql.DB
	Auth          *auth.Config
	Mail          *mail.Client
	PublicBaseURL string // Canonical site URL with scheme, no trailing slash — used in outbound emails.
}

func formatWhen(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(time.Local).Format("January 2, 2006")
}

// Logout clears the session cookie. Accepts GET or POST so the SPA can always sign out even
// if its in-memory CSRF token is stale (the session cookie itself proves identity for clearing).
func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	_ = s.Auth.ClearSession(w, r)
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
