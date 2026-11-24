package handlers

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/mail"
)

// Server wires HTTP handlers to storage and auth.
type Server struct {
	DB            *sql.DB
	Auth          *auth.Config
	Mail          mail.Sender // nil when MAIL_ENABLED is not true
	PublicBaseURL string      // Canonical site URL with scheme, no trailing slash — used in outbound emails.
	// IndexHTML is the built SPA shell (index.html). Used to serve per-event share
	// pages with Open Graph tags so link previews show the event's title and flyer.
	IndexHTML []byte
}

// DefaultPublicBaseURL is used for calendar, share, and Open Graph links when
// PUBLIC_BASE_URL is unset or still points at the Render hostname.
const DefaultPublicBaseURL = "https://albertsonwardplan.com"

// CanonicalPublicBaseURL returns the site's public origin (no trailing slash).
// An empty or onrender.com value falls back to DefaultPublicBaseURL so shared
// links and calendar details use https://albertsonwardplan.com.
func CanonicalPublicBaseURL(configured string) string {
	base := strings.TrimRight(strings.TrimSpace(configured), "/")
	if base == "" {
		return DefaultPublicBaseURL
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return DefaultPublicBaseURL
	}
	host := strings.ToLower(u.Hostname())
	if host == "onrender.com" || strings.HasSuffix(host, ".onrender.com") {
		return DefaultPublicBaseURL
	}
	return base
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
