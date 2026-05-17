package mail

import (
	"bytes"
	"embed"
	"fmt"
	htmltpl "html/template"
	"os"
	"strings"
	texttpl "text/template"
)

//go:embed templates/*.tmpl
var tmplFS embed.FS

var (
	tmplPendingSignupText     *texttpl.Template
	tmplPendingSignupHTML     *htmltpl.Template
	tmplPasswordResetText     *texttpl.Template
	tmplPasswordResetHTML     *htmltpl.Template
	tmplPendingExperienceText *texttpl.Template
	tmplPendingExperienceHTML *htmltpl.Template
)

func init() {
	tmplPendingSignupText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/pending_signup.text.tmpl"))
	tmplPendingSignupHTML = htmltpl.Must(htmltpl.ParseFS(tmplFS, "templates/pending_signup.html.tmpl"))
	tmplPasswordResetText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/password_reset.text.tmpl"))
	tmplPasswordResetHTML = htmltpl.Must(htmltpl.ParseFS(tmplFS, "templates/password_reset.html.tmpl"))
	tmplPendingExperienceText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/pending_experience.text.tmpl"))
	tmplPendingExperienceHTML = htmltpl.Must(htmltpl.ParseFS(tmplFS, "templates/pending_experience.html.tmpl"))
}

// SiteLabel is used in outbound email subjects and headers (MAIL_SITE_LABEL, default "Albertson Ward").
func SiteLabel() string {
	s := strings.TrimSpace(os.Getenv("MAIL_SITE_LABEL"))
	if s == "" {
		return "Albertson Ward"
	}
	return s
}

// RenderPendingSignupEmail builds subject + multipart bodies from embedded templates.
func RenderPendingSignupEmail(requesterDisplayName, requesterEmail, approvalsURL string) (subject string, plain string, html string, err error) {
	label := SiteLabel()
	subject = fmt.Sprintf("%s — New leader account request", label)

	var pb bytes.Buffer
	if err := tmplPendingSignupText.Execute(&pb, struct {
		SiteLabel            string
		RequesterDisplayName string
		RequesterEmail       string
		ApprovalsURL         string
	}{label, requesterDisplayName, requesterEmail, approvalsURL}); err != nil {
		return "", "", "", err
	}
	plain = pb.String()

	var hb bytes.Buffer
	if err := tmplPendingSignupHTML.Execute(&hb, struct {
		SiteLabel            string
		RequesterDisplayName string
		RequesterEmail       string
		ApprovalsURL         htmltpl.URL
	}{label, requesterDisplayName, requesterEmail, htmltpl.URL(approvalsURL)}); err != nil {
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

	var hb bytes.Buffer
	if err := tmplPasswordResetHTML.Execute(&hb, struct {
		SiteLabel   string
		LeaderEmail string
		ResetLink   htmltpl.URL
		ExpiryHours int
	}{label, leaderEmail, htmltpl.URL(resetLink), expiryHours}); err != nil {
		return "", "", "", err
	}
	html = hb.String()
	return subject, plain, html, nil
}

// RenderPendingExperienceEmail builds subject + multipart bodies when a mission experience awaits moderation.
func RenderPendingExperienceEmail(authorLabel, bodyPreview, moderateURL string, submissionID int64) (subject string, plain string, html string, err error) {
	label := SiteLabel()
	subject = fmt.Sprintf("%s — New mission experience to review", label)

	var pb bytes.Buffer
	if err := tmplPendingExperienceText.Execute(&pb, struct {
		SiteLabel    string
		AuthorLabel  string
		BodyPreview  string
		ModerateURL  string
		SubmissionID int64
	}{label, authorLabel, bodyPreview, moderateURL, submissionID}); err != nil {
		return "", "", "", err
	}
	plain = pb.String()

	var hb bytes.Buffer
	if err := tmplPendingExperienceHTML.Execute(&hb, struct {
		SiteLabel    string
		AuthorLabel  string
		BodyPreview  string
		ModerateURL  htmltpl.URL
		SubmissionID int64
	}{label, authorLabel, bodyPreview, htmltpl.URL(moderateURL), submissionID}); err != nil {
		return "", "", "", err
	}
	html = hb.String()
	return subject, plain, html, nil
}
