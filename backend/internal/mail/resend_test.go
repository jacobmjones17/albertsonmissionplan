package mail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResendClientSendsJSON(t *testing.T) {
	var got map[string]any
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"x"}`))
	}))
	defer srv.Close()

	c := NewResendClient("re_test", "Ward <noreply@example.org>", srv.URL)
	if err := c.SendAlternative(context.Background(), []string{" a@example.org ", ""}, "Hi", "plain", "<p>html</p>"); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer re_test" {
		t.Fatalf("auth header %q", auth)
	}
	to, _ := got["to"].([]any)
	if len(to) != 1 || to[0] != "a@example.org" || got["subject"] != "Hi" || got["text"] != "plain" || got["html"] != "<p>html</p>" {
		t.Fatalf("payload %+v", got)
	}
}

func TestResendClientReportsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"domain not verified"}`, http.StatusForbidden)
	}))
	defer srv.Close()
	c := NewResendClient("re_test", "noreply@example.org", srv.URL)
	if err := c.SendAlternative(context.Background(), []string{"a@example.org"}, "s", "p", "h"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFromEnvDisabledIsNilInterface(t *testing.T) {
	t.Setenv("MAIL_ENABLED", "")
	s, err := FromEnv()
	if err != nil || s != nil {
		t.Fatalf("want nil sender, got %v %v", s, err)
	}
}

func TestFromEnvResend(t *testing.T) {
	t.Setenv("MAIL_ENABLED", "true")
	t.Setenv("MAIL_PROVIDER", "resend")
	t.Setenv("RESEND_API_KEY", "re_x")
	t.Setenv("MAIL_FROM", "noreply@example.org")
	s, err := FromEnv()
	if err != nil || ProviderName(s) != "resend" {
		t.Fatalf("got %T %v", s, err)
	}
	t.Setenv("RESEND_API_KEY", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected missing key error")
	}
}
