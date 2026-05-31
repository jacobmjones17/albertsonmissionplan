package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testConfig(idle, max time.Duration) *Config {
	return &Config{
		Store: CookieStoreFromSecret("test-secret-at-least-16-chars"),
		Policy: SessionPolicy{
			IdleTimeout: idle,
			MaxLifetime: max,
		},
	}
}

func TestSessionIdleExpiry(t *testing.T) {
	c := testConfig(30*time.Minute, 24*time.Hour)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := c.SetUserSession(w, r, "leader@example.com", "Pat"); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range w.Result().Cookies() {
		req.AddCookie(cookie)
	}

	sess, err := c.session(req)
	if err != nil {
		t.Fatal(err)
	}
	sess.Values[sessionLastActivityKey] = time.Now().Add(-31 * time.Minute).Unix()
	sess.Options.MaxAge = c.idleMaxAgeSeconds()
	w2 := httptest.NewRecorder()
	if err := sess.Save(req, w2); err != nil {
		t.Fatal(err)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range w2.Result().Cookies() {
		req2.AddCookie(cookie)
	}

	w3 := httptest.NewRecorder()
	if got := c.SessionEmail(w3, req2); got != "" {
		t.Fatalf("expected expired idle session, got email %q", got)
	}
}

func TestSessionTouchExtendsIdle(t *testing.T) {
	c := testConfig(time.Hour, 24*time.Hour)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := c.SetUserSession(w, r, "leader@example.com", ""); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range w.Result().Cookies() {
		req.AddCookie(cookie)
	}

	if got := c.SessionEmail(httptest.NewRecorder(), req); got != "leader@example.com" {
		t.Fatalf("email = %q", got)
	}
	time.Sleep(15 * time.Millisecond)
	w3 := httptest.NewRecorder()
	if got := c.SessionEmail(w3, req); got != "leader@example.com" {
		t.Fatalf("touch failed, email = %q", got)
	}
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range w3.Result().Cookies() {
		req3.AddCookie(cookie)
	}
	sess, _ := c.session(req3)
	last, ok := sessionUnix(sess.Values[sessionLastActivityKey])
	if !ok || time.Since(time.Unix(last, 0)) > 2*time.Second {
		t.Fatal("expected recent last_activity after touch save")
	}
}
