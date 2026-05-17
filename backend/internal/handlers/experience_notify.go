package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/albertson/albertsonmissionplan/internal/mail"
	"github.com/albertson/albertsonmissionplan/internal/store"
)

const experienceEmailPreviewRunes = 500

func truncateForEmailBody(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return strings.TrimSpace(string(r[:maxRunes])) + "…"
}

// notifyPendingExperienceSubmission emails approved leaders when a member submits an experience (best-effort).
func (s *Server) notifyPendingExperienceSubmission(body, authorLabel string, submissionID int64) {
	if s.Mail == nil {
		return
	}
	base := s.trimmedPublicBaseURL()
	if base == "" {
		log.Printf("pending experience notify skipped: PUBLIC_BASE_URL empty")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admins, err := store.ListApprovedLeaderEmails(ctx, s.DB)
	if err != nil {
		log.Printf("list admins for experience notify: %v", err)
		return
	}
	if len(admins) == 0 {
		return
	}
	author := strings.TrimSpace(authorLabel)
	if author == "" {
		author = "Anonymous"
	}
	preview := truncateForEmailBody(body, experienceEmailPreviewRunes)
	moderateURL := fmt.Sprintf("%s/admin/moderate", base)
	subject, plain, html, err := mail.RenderPendingExperienceEmail(author, preview, moderateURL, submissionID)
	if err != nil {
		log.Printf("render pending experience mail: %v", err)
		return
	}
	if err := s.Mail.SendAlternative(ctx, admins, subject, plain, html); err != nil {
		log.Printf("send pending experience mail: %v", err)
	}
}
