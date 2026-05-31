package auth

import (
	"crypto/sha256"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

const (
	sessionName            = "wardmission-session"
	sessionEmailKey        = "email"
	sessionNameKey         = "name"
	sessionIssuedAtKey     = "issued_at"
	sessionLastActivityKey = "last_activity"
)

// Config holds session settings. There is no email allowlist — access is gated entirely by the
// admin-approval flow on the leader_credentials table (see internal/store).
type Config struct {
	Store  *sessions.CookieStore
	Policy SessionPolicy
}

// CookieStoreFromSecret builds a session cookie store (derives 32-byte key).
func CookieStoreFromSecret(secret string) *sessions.CookieStore {
	sum := sha256.Sum256([]byte(secret))
	store := sessions.NewCookieStore(sum[:])
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteLaxMode
	store.Options.Secure = false
	store.Options.Path = "/"
	// MaxAge is set per Save from SessionPolicy (gorilla's default is 30 days).
	store.Options.MaxAge = 0
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

func sessionUnix(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	default:
		return 0, false
	}
}

func (c *Config) idleMaxAgeSeconds() int {
	sec := int(c.Policy.IdleTimeout.Seconds())
	if sec < 60 {
		return 60
	}
	return sec
}

func (c *Config) sessionExpired(issuedAt, lastActivity int64, now time.Time) bool {
	if issuedAt <= 0 || lastActivity <= 0 {
		return true
	}
	if now.Sub(time.Unix(issuedAt, 0)) > c.Policy.MaxLifetime {
		return true
	}
	if now.Sub(time.Unix(lastActivity, 0)) > c.Policy.IdleTimeout {
		return true
	}
	return false
}

func (c *Config) clearSessionCookie(w http.ResponseWriter, r *http.Request, sess *sessions.Session) error {
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}

// resolveSession returns leader identity from the signed session cookie. When touch is true and
// w is non-nil, a valid session refreshes last-activity and extends the cookie idle window.
func (c *Config) resolveSession(w http.ResponseWriter, r *http.Request, touch bool) (email, display string, ok bool) {
	sess, err := c.session(r)
	if err != nil {
		return "", "", false
	}
	email, _ = sess.Values[sessionEmailKey].(string)
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", "", false
	}
	display, _ = sess.Values[sessionNameKey].(string)
	display = strings.TrimSpace(display)

	issuedAt, _ := sessionUnix(sess.Values[sessionIssuedAtKey])
	lastActivity, _ := sessionUnix(sess.Values[sessionLastActivityKey])
	now := time.Now()

	if c.sessionExpired(issuedAt, lastActivity, now) {
		if w != nil {
			_ = c.clearSessionCookie(w, r, sess)
		}
		return "", "", false
	}

	if touch && w != nil {
		sess.Values[sessionLastActivityKey] = now.Unix()
		sess.Options.MaxAge = c.idleMaxAgeSeconds()
		if err := sess.Save(r, w); err != nil {
			return "", "", false
		}
	}
	return email, display, true
}

// SetUserSession stores a signed-in leader identity.
func (c *Config) SetUserSession(w http.ResponseWriter, r *http.Request, email, display string) error {
	sess, err := c.session(r)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	sess.Values[sessionEmailKey] = strings.ToLower(strings.TrimSpace(email))
	sess.Values[sessionNameKey] = strings.TrimSpace(display)
	sess.Values[sessionIssuedAtKey] = now
	sess.Values[sessionLastActivityKey] = now
	sess.Options.MaxAge = c.idleMaxAgeSeconds()
	return sess.Save(r, w)
}

// ClearSession removes the session cookie.
func (c *Config) ClearSession(w http.ResponseWriter, r *http.Request) error {
	sess, err := c.session(r)
	if err != nil {
		return err
	}
	return c.clearSessionCookie(w, r, sess)
}

// SessionIdentity returns email and display name when the session is valid. Refreshes idle
// timeout when w is non-nil (call once per request instead of SessionEmail + SessionDisplayName).
func (c *Config) SessionIdentity(w http.ResponseWriter, r *http.Request) (email, display string) {
	email, display, ok := c.resolveSession(w, r, w != nil)
	if !ok {
		return "", ""
	}
	return email, display
}

// SessionEmail returns the signed-in leader email, if any. Refreshes idle timeout when w is set.
func (c *Config) SessionEmail(w http.ResponseWriter, r *http.Request) string {
	email, _, ok := c.resolveSession(w, r, w != nil)
	if !ok {
		return ""
	}
	return email
}

// SessionDisplayName returns a display name stored at sign-in, if any.
func (c *Config) SessionDisplayName(w http.ResponseWriter, r *http.Request) string {
	_, display, ok := c.resolveSession(w, r, w != nil)
	if !ok {
		return ""
	}
	return display
}
