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
	requestedAt := time.Now()
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
	subject, plain, html, err := mail.RenderPendingSignupEmail(name, email, fmt.Sprintf("%s/admin/approvals", base), requestedAt)
	if err != nil {
		log.Printf("render pending signup mail: %v", err)
		return
	}
	s.sendToEachLeader(ctx, "pending signup", admins, subject, plain, html)
}

// sendToEachLeader sends one message per recipient so leaders don't see each other's
// addresses and one bad address can't stop the rest. Failures are logged, never returned:
// notifications are best-effort and must not affect the request that triggered them.
func (s *Server) sendToEachLeader(ctx context.Context, what string, recipients []string, subject, plain, html string) {
	sent := 0
	for _, to := range recipients {
		if err := s.Mail.SendAlternative(ctx, []string{to}, subject, plain, html); err != nil {
			log.Printf("send %s mail to %s: %v", what, to, err)
			continue
		}
		sent++
	}
	log.Printf("%s mail: sent to %d of %d leaders", what, sent, len(recipients))
}
