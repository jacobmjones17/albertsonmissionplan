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

func (s *Server) trimmedPublicBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(s.PublicBaseURL), "/")
}

// notifyPendingLeaderSignup emails approved leaders (best-effort). Does not block the HTTP handler.
func (s *Server) notifyPendingLeaderSignup(email, firstName, lastName string) {
	if s.Mail == nil {
		return
	}
	base := s.trimmedPublicBaseURL()
	if base == "" {
		log.Printf("pending signup notify skipped: PUBLIC_BASE_URL empty")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admins, err := store.ListApprovedLeaderEmails(ctx, s.DB)
	if err != nil {
		log.Printf("list admins for pending signup notify: %v", err)
		return
	}
	if len(admins) == 0 {
		return
	}
	fn := strings.TrimSpace(firstName)
	ln := strings.TrimSpace(lastName)
	name := strings.TrimSpace(fn + " " + ln)
	subject, plain, html, err := mail.RenderPendingSignupEmail(name, email, fmt.Sprintf("%s/admin/approvals", base))
	if err != nil {
		log.Printf("render pending signup mail: %v", err)
		return
	}
	if err := s.Mail.SendAlternative(ctx, admins, subject, plain, html); err != nil {
		log.Printf("send pending signup mail: %v", err)
	}
}
