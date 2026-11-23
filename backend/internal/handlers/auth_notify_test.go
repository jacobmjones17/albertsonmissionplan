package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/auth"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

type sentMail struct {
	to      []string
	subject string
	plain   string
}

type fakeSender struct {
	mu   sync.Mutex
	sent []sentMail
	fail bool
	done chan struct{}
}

func (f *fakeSender) SendAlternative(_ context.Context, to []string, subject, plain, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sentMail{to: append([]string(nil), to...), subject: subject, plain: plain})
	if f.done != nil {
		select {
		case f.done <- struct{}{}:
		default:
		}
	}
	if f.fail {
		return errors.New("smtp unavailable")
	}
	return nil
}

func newTestServer(t *testing.T, sender *fakeSender) *Server {
	t.Helper()
	t.Setenv("DATABASE_PATH", t.TempDir()+"/test.db")
	t.Setenv("DATABASE_URL", "")
	db, err := store.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ac := &auth.Config{Store: auth.CookieStoreFromSecret("test-secret-0123456789abcdef"), Policy: auth.SessionPolicyFromEnv()}
	s := &Server{DB: db, Auth: ac, PublicBaseURL: "https://ward.example.org/"}
	if sender != nil {
		s.Mail = sender
	}
	return s
}

func register(t *testing.T, s *Server, email, first, last string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": "pw-" + first, "firstName": first, "lastName": last})
	rec := httptest.NewRecorder()
	s.APIAuthRegister(rec, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("register %s: status %d body %s", email, rec.Code, rec.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return out
}

// waitForSends waits until n messages were attempted (notifications run in a goroutine).
func waitForSends(t *testing.T, f *fakeSender, n int) []sentMail {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		f.mu.Lock()
		got := len(f.sent)
		f.mu.Unlock()
		if got >= n {
			break
		}
		select {
		case <-f.done:
		case <-deadline:
			t.Fatalf("timed out waiting for %d sends, got %d", n, got)
		}
	}
	time.Sleep(50 * time.Millisecond) // catch unexpected extra sends
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sentMail(nil), f.sent...)
}

func TestLeaderRequestEmailsEveryApprovedLeader(t *testing.T) {
	f := &fakeSender{done: make(chan struct{}, 16)}
	s := newTestServer(t, f)
	ctx := context.Background()

	// First account bootstraps as approved (no email: nobody to notify yet).
	if out := register(t, s, "first@example.org", "Ada", "One"); out["status"] != "approved" {
		t.Fatalf("first registration: %v", out)
	}
	// A second approved leader and one still-pending account.
	hash, _ := auth.HashPassword("x")
	if err := store.InsertLeaderCredential(ctx, s.DB, "second@example.org", "Bea", "Two", hash, "first@example.org"); err != nil {
		t.Fatal(err)
	}
	if err := store.InsertLeaderCredential(ctx, s.DB, "pending@example.org", "Cy", "Three", hash, ""); err != nil {
		t.Fatal(err)
	}

	if out := register(t, s, "newbie@example.org", "Dee", "Four"); out["status"] != "pending" {
		t.Fatalf("new request: %v", out)
	}
	sent := waitForSends(t, f, 2)
	if len(sent) != 2 {
		t.Fatalf("want 2 emails (one per approved leader), got %d: %+v", len(sent), sent)
	}
	got := map[string]bool{}
	for _, m := range sent {
		if len(m.to) != 1 {
			t.Fatalf("each email should have exactly one recipient, got %v", m.to)
		}
		got[m.to[0]] = true
		if !strings.Contains(m.plain, "https://ward.example.org/admin/approvals") {
			t.Fatalf("email missing approvals link: %q", m.plain)
		}
		if !strings.Contains(m.plain, "Dee Four") || !strings.Contains(m.plain, "newbie@example.org") {
			t.Fatalf("email missing requester details: %q", m.plain)
		}
	}
	if !got["first@example.org"] || !got["second@example.org"] || got["pending@example.org"] {
		t.Fatalf("wrong recipients: %v", got)
	}
}

func TestLeaderRequestSucceedsWhenEmailFails(t *testing.T) {
	f := &fakeSender{fail: true, done: make(chan struct{}, 16)}
	s := newTestServer(t, f)
	register(t, s, "first@example.org", "Ada", "One")
	if out := register(t, s, "newbie@example.org", "Dee", "Four"); out["status"] != "pending" {
		t.Fatalf("registration should still succeed when email fails: %v", out)
	}
	waitForSends(t, f, 1)
}

func TestLeaderRequestWithMailDisabled(t *testing.T) {
	s := newTestServer(t, nil)
	if s.Mail != nil {
		t.Fatal("Mail should be a nil interface")
	}
	register(t, s, "first@example.org", "Ada", "One")
	if out := register(t, s, "newbie@example.org", "Dee", "Four"); out["status"] != "pending" {
		t.Fatalf("got %v", out)
	}
}
