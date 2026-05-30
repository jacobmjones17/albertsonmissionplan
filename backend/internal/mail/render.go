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
	tmplPasswordResetText     *texttpl.Template
	tmplPendingExperienceText *texttpl.Template
	htmlTemplates             *htmltpl.Template
)

func init() {
	tmplPendingSignupText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/pending_signup.text.tmpl"))
	tmplPasswordResetText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/password_reset.text.tmpl"))
	tmplPendingExperienceText = texttpl.Must(texttpl.ParseFS(tmplFS, "templates/pending_experience.text.tmpl"))
	htmlTemplates = htmltpl.Must(htmltpl.ParseFS(tmplFS,
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
	EmailTitle string
	Headline   string
	Preheader  string
	FooterNote string
	CTAURL     htmltpl.URL
	CTALabel   string
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

	data := struct {
		htmlEmailData
		RequesterDisplayName string
		RequesterEmail       string
		ApprovalsURL         htmltpl.URL
	}{
		htmlEmailData: htmlEmailData{
			SiteLabel:  label,
			EmailTitle: "New leader account request",
			Headline:   "New leader account request",
			Preheader:  fmt.Sprintf("%s requested access to leader tools", requesterDisplayName),
			FooterNote: "You received this email because you are an approved ward leader.",
			CTAURL:     htmltpl.URL(approvalsURL),
			CTALabel:   "Review pending accounts",
		},
		RequesterDisplayName: requesterDisplayName,
		RequesterEmail:       requesterEmail,
		ApprovalsURL:         htmltpl.URL(approvalsURL),
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
func RenderPendingExperienceEmail(authorLabel, bodyPreview, moderateURL string) (subject string, plain string, html string, err error) {
	label := SiteLabel()
	subject = fmt.Sprintf("%s — New mission experience to review", label)

	var pb bytes.Buffer
	if err := tmplPendingExperienceText.Execute(&pb, struct {
		SiteLabel   string
		AuthorLabel string
		BodyPreview string
		ModerateURL string
	}{label, authorLabel, bodyPreview, moderateURL}); err != nil {
		return "", "", "", err
	}
	plain = pb.String()

	data := struct {
		htmlEmailData
		AuthorLabel string
		BodyPreview string
		ModerateURL htmltpl.URL
	}{
		htmlEmailData: htmlEmailData{
			SiteLabel:  label,
			EmailTitle: "New mission experience",
			Headline:   "New mission experience",
			Preheader:  fmt.Sprintf("From %s — waiting for your review", authorLabel),
			FooterNote: "You received this email because you are an approved ward leader.",
			CTAURL:     htmltpl.URL(moderateURL),
			CTALabel:   "Review moderation queue",
		},
		AuthorLabel: authorLabel,
		BodyPreview: bodyPreview,
		ModerateURL: htmltpl.URL(moderateURL),
	}

	var hb bytes.Buffer
	if err := htmlTemplates.ExecuteTemplate(&hb, "pending_experience.html.tmpl", data); err != nil {
		return "", "", "", err
	}
	html = hb.String()
	return subject, plain, html, nil
}
