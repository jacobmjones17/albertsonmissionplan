package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	sessionName     = "wardmission-session"
	sessionEmailKey = "email"
	sessionNameKey  = "name"
	oauthStateKey   = "oauth_state"
)

// Config holds OAuth and session settings.
type Config struct {
	OAuth   *oauth2.Config
	Store   *sessions.CookieStore
	Allowed map[string]struct{}
}

// NewGoogleOAuth builds a Google OAuth2 config.
func NewGoogleOAuth(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
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

// ParseAllowlist returns lowercase email -> struct{} for membership checks.
func ParseAllowlist(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		e := strings.TrimSpace(strings.ToLower(part))
		if e != "" {
			out[e] = struct{}{}
		}
	}
	return out
}

// IsLeader reports whether an email may use leader tools.
func (c *Config) IsLeader(email string) bool {
	if email == "" {
		return false
	}
	_, ok := c.Allowed[strings.ToLower(strings.TrimSpace(email))]
	return ok
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

// SetOAuthState stores CSRF state for the OAuth handshake.
func (c *Config) SetOAuthState(w http.ResponseWriter, r *http.Request, state string) error {
	sess, err := c.session(r)
	if err != nil {
		return err
	}
	sess.Values[oauthStateKey] = state
	return sess.Save(r, w)
}

// ValidOAuthState checks the state param and clears it when valid.
func (c *Config) ValidOAuthState(w http.ResponseWriter, r *http.Request, state string) (bool, error) {
	sess, err := c.session(r)
	if err != nil {
		return false, err
	}
	want, _ := sess.Values[oauthStateKey].(string)
	delete(sess.Values, oauthStateKey)
	if err := sess.Save(r, w); err != nil {
		return false, err
	}
	return want != "" && want == state, nil
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

// SessionDisplayName returns a display name from Google profile if stored.
func (c *Config) SessionDisplayName(r *http.Request) string {
	sess, err := c.session(r)
	if err != nil {
		return ""
	}
	v, _ := sess.Values[sessionNameKey].(string)
	return v
}

// RandomState returns a URL-safe random string for OAuth state.
func RandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// FetchGoogleEmail uses an OAuth-authenticated client to load the Google account email and name.
func FetchGoogleEmail(ctx context.Context, client *http.Client) (email, name string, err error) {
	if client == nil {
		return "", "", errors.New("nil http client")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", "", errors.New("userinfo failed: " + string(body))
	}
	var u struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", "", err
	}
	return u.Email, u.Name, nil
}
