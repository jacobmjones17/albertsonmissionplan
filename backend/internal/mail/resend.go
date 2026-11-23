package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultResendEndpoint = "https://api.resend.com/emails"

// ResendClient sends mail through Resend's HTTPS API (https://resend.com/docs/api-reference/emails/send-email).
// Useful where outbound SMTP ports are blocked (Render free web services).
type ResendClient struct {
	apiKey     string
	fromHeader string
	endpoint   string
	http       *http.Client
}

func resendFromEnv() (*ResendClient, error) {
	key := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	from := strings.TrimSpace(os.Getenv("MAIL_FROM"))
	if key == "" || from == "" {
		return nil, fmt.Errorf("MAIL_PROVIDER=resend requires RESEND_API_KEY and MAIL_FROM")
	}
	return NewResendClient(key, fromHeaderFromEnv(from), ""), nil
}

// NewResendClient builds a client; endpoint "" uses Resend's production API.
func NewResendClient(apiKey, fromHeader, endpoint string) *ResendClient {
	if endpoint == "" {
		endpoint = defaultResendEndpoint
	}
	return &ResendClient{
		apiKey:     apiKey,
		fromHeader: fromHeader,
		endpoint:   endpoint,
		http:       &http.Client{Timeout: 20 * time.Second},
	}
}

// SendAlternative posts one email (plain + HTML) to all recipients.
func (c *ResendClient) SendAlternative(ctx context.Context, to []string, subject, plainBody, htmlBody string) error {
	if c == nil {
		return nil
	}
	var clean []string
	for _, addr := range to {
		if addr = strings.TrimSpace(addr); addr != "" {
			clean = append(clean, addr)
		}
	}
	if len(clean) == 0 {
		return nil
	}
	payload, err := json.Marshal(map[string]any{
		"from":    c.fromHeader,
		"to":      clean,
		"subject": subject,
		"text":    plainBody,
		"html":    htmlBody,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("resend: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
