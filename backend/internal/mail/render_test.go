package mail

import (
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 2026-11-23 07:20 UTC = 12:20 AM MST in America/Denver.
var sampleTime = time.Date(2026, 11, 23, 7, 20, 0, 0, time.UTC)

func TestPendingSignupEmail(t *testing.T) {
	t.Setenv("MAIL_TIME_ZONE", "")
	subject, plain, html, err := RenderPendingSignupEmail("Jacob Test", "jacob@example.org", "https://ward.example.org/admin/approvals", sampleTime)
	if err != nil {
		t.Fatal(err)
	}
	if subject != "New leader request: Jacob Test" {
		t.Errorf("subject = %q", subject)
	}
	const when = "Monday, November 23, 2026 at 12:20 AM MST"
	for _, want := range []string{"Jacob Test", "jacob@example.org", when, "https://ward.example.org/admin/approvals"} {
		if !strings.Contains(plain, want) {
			t.Errorf("plain text missing %q:\n%s", want, plain)
		}
		if !strings.Contains(html, want) {
			t.Errorf("html missing %q", want)
		}
	}
	assertLightTheme(t, html)
}

func TestPendingExperienceEmail(t *testing.T) {
	t.Setenv("MAIL_TIME_ZONE", "")
	subject, plain, html, err := RenderPendingExperienceEmail("Sarah", "We invited <b>our</b> neighbor.", "https://ward.example.org/admin/moderate", sampleTime)
	if err != nil {
		t.Fatal(err)
	}
	if subject != "New missionary experience from Sarah" {
		t.Errorf("subject = %q", subject)
	}
	for _, want := range []string{"Sarah", "Monday, November 23, 2026 at 12:20 AM MST", "https://ward.example.org/admin/moderate"} {
		if !strings.Contains(plain, want) || !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(html, "<b>our</b>") || !strings.Contains(html, "&lt;b&gt;our&lt;/b&gt;") {
		t.Error("story text must be HTML-escaped")
	}
	assertLightTheme(t, html)

	subject, _, _, err = RenderPendingExperienceEmail("  ", "x", "https://x/admin/moderate", sampleTime)
	if err != nil || subject != "New missionary experience (anonymous)" {
		t.Errorf("anonymous subject = %q, %v", subject, err)
	}
}

func TestSubjectNameIsOneLine(t *testing.T) {
	subject, _, _, err := RenderPendingSignupEmail("Eve\r\nBcc: evil@example.org", "e@example.org", "https://x/admin/approvals", sampleTime)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(subject, "\r\n") {
		t.Fatalf("subject contains a line break: %q", subject)
	}
}

func TestWardTimeZoneOverride(t *testing.T) {
	t.Setenv("MAIL_TIME_ZONE", "America/Boise")
	if got := FormatWardTime(sampleTime); got != "Monday, November 23, 2026 at 12:20 AM MST" {
		t.Errorf("Boise: %q", got)
	}
	t.Setenv("MAIL_TIME_ZONE", "America/Los_Angeles")
	if got := FormatWardTime(sampleTime); got != "Sunday, November 22, 2026 at 11:20 PM PST" {
		t.Errorf("LA: %q", got)
	}
	t.Setenv("MAIL_TIME_ZONE", "Not/AZone")
	if got := FormatWardTime(sampleTime); !strings.HasSuffix(got, "MST") {
		t.Errorf("bad zone should fall back to Denver: %q", got)
	}
}

// assertLightTheme checks the restyle: parchment card, readable header type, gold CTA, and no
// background images (some dark-mode clients skip inverting those, which made the old header
// text navy-on-navy).
func assertLightTheme(t *testing.T, html string) {
	t.Helper()
	for _, want := range []string{
		"background-color:#fbf6ea", // parchment card
		"color:#1b1812",            // dark body text
		"color:#e8c98a",            // gold eyebrow on navy
		"color:#ffffff",            // white headline on navy
		"background-color:#e0b868", // gold CTA
		"color:#1b140a",            // dark CTA text
		`content="light only"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("html missing %q", want)
		}
	}
	for _, bad := range []string{"linear-gradient", "background-image", "<link"} {
		if strings.Contains(html, bad) {
			t.Errorf("html should not contain %q", bad)
		}
	}
}

func TestSMTPMessageIsMultipartAlternative(t *testing.T) {
	raw := buildAlternativeMessage("Ward <noreply@example.org>", []string{"a@example.org"},
		"New leader request: José Test", "plain body\nline 2", "<p>html body</p>", "BOUNDARY123")
	msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	subj, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if err != nil || subj != "New leader request: José Test" {
		t.Errorf("subject = %q, %v", subj, err)
	}
	mt, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil || mt != "multipart/alternative" {
		t.Fatalf("content-type = %q, %v", mt, err)
	}
	mr := multipart.NewReader(msg.Body, params["boundary"])
	var types []string
	for {
		p, err := mr.NextPart()
		if err != nil {
			break
		}
		types = append(types, strings.Split(p.Header.Get("Content-Type"), ";")[0])
	}
	if strings.Join(types, ",") != "text/plain,text/html" {
		t.Errorf("parts = %v, want text/plain then text/html", types)
	}
}

// TestWriteEmailSamples writes sample HTML/text files when EMAIL_SAMPLE_DIR is set (used to
// render screenshots for review). It is a no-op in normal test runs.
func TestWriteEmailSamples(t *testing.T) {
	dir := os.Getenv("EMAIL_SAMPLE_DIR")
	if dir == "" {
		t.Skip("set EMAIL_SAMPLE_DIR to write samples")
	}
	t.Setenv("MAIL_TIME_ZONE", "")
	write := func(name, subject, plain, html string) {
		for ext, body := range map[string]string{".html": html, ".txt": "Subject: " + subject + "\n\n" + plain} {
			if err := os.WriteFile(filepath.Join(dir, name+ext), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	s, p, h, err := RenderPendingSignupEmail("Jacob Test", "jacob.test@example.org", "https://albertsonwardplan.com/admin/approvals", sampleTime)
	if err != nil {
		t.Fatal(err)
	}
	write("leader-request", s, p, h)
	story := "Last week Brother Hansen and I visited a neighbor who had just moved in. We brought cookies and an invitation to the ward Christmas party.\n\nShe came, and afterward she told us it was the first time she felt at home since the move."
	s, p, h, err = RenderPendingExperienceEmail("Sarah", story, "https://albertsonwardplan.com/admin/moderate", sampleTime)
	if err != nil {
		t.Fatal(err)
	}
	write("missionary-experience", s, p, h)
	s, p, h, err = RenderPasswordResetEmail("jacob@example.org", "https://albertsonwardplan.com/admin/reset-password?token=sample", 24)
	if err != nil {
		t.Fatal(err)
	}
	write("password-reset", s, p, h)
}
