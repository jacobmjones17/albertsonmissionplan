package mail

import (
	"bytes"
	"embed"
	"fmt"
	htmltpl "html/template"
	"os"
	"strings"
	texttpl "text/template"
	"time"
	_ "time/tzdata" // MAIL_TIME_ZONE works even where the OS has no zoneinfo
)

// DefaultTimeZone is used for "received at" times in leader emails when MAIL_TIME_ZONE is unset.
const DefaultTimeZone = "America/Denver"

// wardLocation returns MAIL_TIME_ZONE (IANA name) or America/Denver.
func wardLocation() *time.Location {
	name := strings.TrimSpace(os.Getenv("MAIL_TIME_ZONE"))
	if name == "" {
		name = DefaultTimeZone
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		loc, _ = time.LoadLocation(DefaultTimeZone)
	}
	return loc
}

// FormatWardTime renders t in the ward's zone with the zone abbreviation, e.g.
// "Monday, November 23, 2026 at 12:20 AM MST".
func FormatWardTime(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	return t.In(wardLocation()).Format("Monday, January 2, 2006 at 3:04 PM MST")
}

// subjectName keeps user-entered names safe for a Subject header: one line, trimmed, capped.
func subjectName(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > 60 {
		s = strings.TrimSpace(string(r[:60])) + "…"
	}
	return s
}

//go:embed templates/*.tmpl
var tmplFS embed.FS

var (
	tmplPendingSignupText     *texttpl.Template
	tmplPasswordResetText     *texttpl.Template
	tmplPendingExperienceText *texttpl.Template
	htmlTemplates             *htmltpl.Template
)

func init() {
	tmplPendingSignupText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/pending_signup.text.tmpl"))
	tmplPasswordResetText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/password_reset.text.tmpl"))
	tmplPendingExperienceText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/pending_experience.text.tmpl"))
	htmlTemplates = htmltpl.Must(htmltpl.New("email").Funcs(htmltpl.FuncMap{
		// detail builds the {Label, Value} pair used by the "email_detail_row" partial.
		"detail": func(label, value string) map[string]string {
			return map[string]string{"Label": label, "Value": value}
		},
	}).ParseFS(tmplFS,
		"templates/email_layout.html.tmpl",
		"templates/pending_signup.html.tmpl",
		"templates/password_reset.html.tmpl",
		"templates/pending_experience.html.tmpl",
	))
}

// SiteLabel is used in outbound email subjects and headers (MAIL_SITE_LABEL, default "Albertson Ward").
func SiteLabel() string {
	s := strings.TrimSpace(os.Getenv("MAIL_SITE_LABEL"))
	if s == "" {
		return "Albertson Ward"
	}
	return s
}

type htmlEmailData struct {
	SiteLabel  string
	Eyebrow    string
	EmailTitle string
	Headline   string
	Preheader  string
	FooterNote string
	CTAURL     htmltpl.URL
	CTALabel   string
}

// RenderPendingSignupEmail builds subject + multipart bodies from embedded templates.
func RenderPendingSignupEmail(requesterDisplayName, requesterEmail, approvalsURL string, requestedAt time.Time) (subject string, plain string, html string, err error) {
	label := SiteLabel()
	name := subjectName(requesterDisplayName)
	if name == "" {
		name = subjectName(requesterEmail)
	}
	subject = "New leader request: " + name
	when := FormatWardTime(requestedAt)

	var pb bytes.Buffer
	if err := tmplPendingSignupText.Execute(&pb, struct {
		SiteLabel            string
		RequesterDisplayName string
		RequesterEmail       string
		ApprovalsURL         string
		RequestedAt          string
	}{label, requesterDisplayName, requesterEmail, approvalsURL, when}); err != nil {
		return "", "", "", err
	}
	plain = pb.String()

	data := struct {
		htmlEmailData
		RequesterDisplayName string
		RequesterEmail       string
		ApprovalsURL         htmltpl.URL
		RequestedAt          string
	}{
		htmlEmailData: htmlEmailData{
			SiteLabel:  label,
			Eyebrow:    label + " · Leader Tools",
			EmailTitle: subject,
			Headline:   "New leader request",
			Preheader:  fmt.Sprintf("%s asked for access to leader tools on %s", name, when),
			FooterNote: "You received this email because you are an approved ward leader.",
			CTAURL:     htmltpl.URL(approvalsURL),
			CTALabel:   "Review leader requests",
		},
		RequesterDisplayName: requesterDisplayName,
		RequesterEmail:       requesterEmail,
		ApprovalsURL:         htmltpl.URL(approvalsURL),
		RequestedAt:          when,
	}

	var hb bytes.Buffer
	if err := htmlTemplates.ExecuteTemplate(&hb, "pending_signup.html.tmpl", data); err != nil {
		return "", "", "", err
	}
	html = hb.String()
	return subject, plain, html, nil
}

// RenderPasswordResetEmail builds subject + multipart bodies from embedded templates.
func RenderPasswordResetEmail(leaderEmail, resetLink string, expiryHours int) (subject string, plain string, html string, err error) {
	label := SiteLabel()
	subject = fmt.Sprintf("%s — Reset your leader password", label)

	var pb bytes.Buffer
	if err := tmplPasswordResetText.Execute(&pb, struct {
		SiteLabel   string
		LeaderEmail string
		ResetLink   string
		ExpiryHours int
	}{label, leaderEmail, resetLink, expiryHours}); err != nil {
		return "", "", "", err
	}
	plain = pb.String()

	data := struct {
		htmlEmailData
		LeaderEmail string
		ResetLink   htmltpl.URL
		ExpiryHours int
	}{
		htmlEmailData: htmlEmailData{
			SiteLabel:  label,
			Eyebrow:    label + " · Leader Tools",
			EmailTitle: "Reset your leader password",
			Headline:   "Reset your password",
			Preheader:  "Use the link inside to choose a new leader password",
			FooterNote: "If you did not ask for this reset, you can ignore this email.",
			CTAURL:     htmltpl.URL(resetLink),
			CTALabel:   "Choose a new password",
		},
		LeaderEmail: leaderEmail,
		ResetLink:   htmltpl.URL(resetLink),
		ExpiryHours: expiryHours,
	}

	var hb bytes.Buffer
	if err := htmlTemplates.ExecuteTemplate(&hb, "password_reset.html.tmpl", data); err != nil {
		return "", "", "", err
	}
	html = hb.String()
	return subject, plain, html, nil
}

// RenderPendingExperienceEmail builds subject + multipart bodies when a mission experience awaits moderation.
func RenderPendingExperienceEmail(authorLabel, bodyPreview, moderateURL string, submittedAt time.Time) (subject string, plain string, html string, err error) {
	label := SiteLabel()
	author := subjectName(authorLabel)
	if author == "" {
		author = "Anonymous"
	}
	if strings.EqualFold(author, "anonymous") {
		subject = "New missionary experience (anonymous)"
	} else {
		subject = "New missionary experience from " + author
	}
	when := FormatWardTime(submittedAt)

	var pb bytes.Buffer
	if err := tmplPendingExperienceText.Execute(&pb, struct {
		SiteLabel   string
		AuthorLabel string
		BodyPreview string
		ModerateURL string
		SubmittedAt string
	}{label, author, bodyPreview, moderateURL, when}); err != nil {
		return "", "", "", err
	}
	plain = pb.String()

	data := struct {
		htmlEmailData
		AuthorLabel string
		BodyPreview string
		ModerateURL htmltpl.URL
		SubmittedAt string
	}{
		htmlEmailData: htmlEmailData{
			SiteLabel:  label,
			Eyebrow:    label + " · Leader Tools",
			EmailTitle: subject,
			Headline:   "New missionary experience",
			Preheader:  fmt.Sprintf("From %s, %s. Waiting for your review.", author, when),
			FooterNote: "You received this email because you are an approved ward leader.",
			CTAURL:     htmltpl.URL(moderateURL),
			CTALabel:   "Review experiences",
		},
		AuthorLabel: author,
		BodyPreview: bodyPreview,
		ModerateURL: htmltpl.URL(moderateURL),
		SubmittedAt: when,
	}

	var hb bytes.Buffer
	if err := htmlTemplates.ExecuteTemplate(&hb, "pending_experience.html.tmpl", data); err != nil {
		return "", "", "", err
	}
	html = hb.String()
	return subject, plain, html, nil
}
