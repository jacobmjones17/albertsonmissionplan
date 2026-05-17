package mail

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// Client sends plain-text email via SMTP (STARTTLS on port 587 works with SendGrid,
// Mailgun, Amazon SES SMTP relay, Gmail app passwords, etc.).
type Client struct {
	addr       string
	host       string
	user       string
	password   string
	fromAddr   string
	fromHeader string
}

// FromEnv builds a Client when MAIL_ENABLED=true and required SMTP vars are set.
// Returns nil when mail is disabled (not an error).
func FromEnv() (*Client, error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("MAIL_ENABLED")), "true") {
		return nil, nil
	}
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	portStr := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
	from := strings.TrimSpace(os.Getenv("MAIL_FROM"))
	name := strings.TrimSpace(os.Getenv("MAIL_FROM_NAME"))
	if host == "" || portStr == "" || user == "" || pass == "" || from == "" {
		return nil, fmt.Errorf("MAIL_ENABLED=true requires SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, MAIL_FROM")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid SMTP_PORT %q", portStr)
	}
	fromHeader := from
	if name != "" {
		line := strings.ReplaceAll(strings.ReplaceAll(name, "\r", " "), "\n", " ")
		fromHeader = fmt.Sprintf("%s <%s>", line, from)
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	return &Client{
		addr: addr, host: host, user: user, password: pass,
		fromAddr: from, fromHeader: fromHeader,
	}, nil
}

// SendPlain sends a UTF-8 plain-text body to all recipients in one SMTP transaction.
func (c *Client) SendPlain(_ context.Context, to []string, subject, body string) error {
	if c == nil || len(to) == 0 {
		return nil
	}
	var clean []string
	for _, addr := range to {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			clean = append(clean, addr)
		}
	}
	if len(clean) == 0 {
		return nil
	}
	auth := smtp.PlainAuth("", c.user, c.password, c.host)
	msg := strings.Builder{}
	fmt.Fprintf(&msg, "From: %s\r\n", c.fromHeader)
	fmt.Fprintf(&msg, "To: %s\r\n", strings.Join(clean, ", "))
	fmt.Fprintf(&msg, "Subject: %s\r\n", encodeSubject(subject))
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "\r\n%s\r\n", body)
	return smtp.SendMail(c.addr, auth, c.fromAddr, clean, []byte(msg.String()))
}

func randomMIMEBoundary() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "=_ward_" + hex.EncodeToString(b[:])
}

func crlfSMTPBody(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}

// SendAlternative sends multipart/alternative plain text + HTML in one SMTP transaction.
func (c *Client) SendAlternative(_ context.Context, to []string, subject, plainBody, htmlBody string) error {
	if c == nil || len(to) == 0 {
		return nil
	}
	var clean []string
	for _, addr := range to {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			clean = append(clean, addr)
		}
	}
	if len(clean) == 0 {
		return nil
	}
	auth := smtp.PlainAuth("", c.user, c.password, c.host)
	boundary := randomMIMEBoundary()
	plainPart := crlfSMTPBody(plainBody)
	htmlPart := crlfSMTPBody(htmlBody)

	var msg strings.Builder
	fmt.Fprintf(&msg, "From: %s\r\n", c.fromHeader)
	fmt.Fprintf(&msg, "To: %s\r\n", strings.Join(clean, ", "))
	fmt.Fprintf(&msg, "Subject: %s\r\n", encodeSubject(subject))
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=%q\r\n", boundary)
	fmt.Fprintf(&msg, "\r\n--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: 8bit\r\n\r\n")
	fmt.Fprintf(&msg, "%s\r\n", plainPart)
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&msg, "Content-Transfer-Encoding: 8bit\r\n\r\n")
	fmt.Fprintf(&msg, "%s\r\n", htmlPart)
	fmt.Fprintf(&msg, "--%s--\r\n", boundary)
	return smtp.SendMail(c.addr, auth, c.fromAddr, clean, []byte(msg.String()))
}

func encodeSubject(s string) string {
	if s == "" {
		return ""
	}
	for _, r := range s {
		if r < 32 || r > 126 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
		}
	}
	return s
}
