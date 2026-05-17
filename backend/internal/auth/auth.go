package auth

import (
	"crypto/sha256"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

const (
	sessionName     = "wardmission-session"
	sessionEmailKey = "email"
	sessionNameKey  = "name"
)

// Config holds session settings. There is no email allowlist — access is gated entirely by the
// admin-approval flow on the leader_credentials table (see internal/store).
type Config struct {
	Store *sessions.CookieStore
}

// CookieStoreFromSecret builds a session cookie store (derives 32-byte key).
func CookieStoreFromSecret(secret string) *sessions.CookieStore {
	sum := sha256.Sum256([]byte(secret))
	store := sessions.NewCookieStore(sum[:])
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteLaxMode
	store.Options.Secure = false
	store.Options.Path = "/"
	return store
}

// SetCookieSecure sets the Secure flag on session cookies (use behind HTTPS).
func SetCookieSecure(store *sessions.CookieStore, secure bool) {
	store.Options.Secure = secure
}

// session returns the named session. If an old cookie cannot be decoded (e.g. SESSION_SECRET
// changed), clears values so Save can replace the cookie instead of failing the whole flow.
func (c *Config) session(r *http.Request) (*sessions.Session, error) {
	sess, err := c.Store.Get(r, sessionName)
	if err != nil {
		if sess == nil {
			return nil, err
		}
		sess.Values = make(map[interface{}]interface{})
	}
	return sess, nil
}

// SetUserSession stores a signed-in leader identity.
func (c *Config) SetUserSession(w http.ResponseWriter, r *http.Request, email, display string) error {
	sess, err := c.session(r)
	if err != nil {
		return err
	}
	sess.Values[sessionEmailKey] = strings.ToLower(strings.TrimSpace(email))
	sess.Values[sessionNameKey] = strings.TrimSpace(display)
	return sess.Save(r, w)
}

// ClearSession removes the session cookie.
func (c *Config) ClearSession(w http.ResponseWriter, r *http.Request) error {
	sess, err := c.session(r)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}

// SessionEmail returns the signed-in leader email, if any.
func (c *Config) SessionEmail(r *http.Request) string {
	sess, err := c.session(r)
	if err != nil {
		return ""
	}
	v, _ := sess.Values[sessionEmailKey].(string)
	return v
}

// SessionDisplayName returns a display name stored at sign-in, if any.
func (c *Config) SessionDisplayName(r *http.Request) string {
	sess, err := c.session(r)
	if err != nil {
		return ""
	}
	v, _ := sess.Values[sessionNameKey].(string)
	return v
}
