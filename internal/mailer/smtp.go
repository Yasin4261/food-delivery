// Package mailer holds the driven adapters for the domain.Mailer port: an SMTP
// sender for production and a logging stub for development.
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// SMTP delivers email through an SMTP server. It implements domain.Mailer.
type SMTP struct {
	addr     string // host:port
	host     string
	from     string
	auth     smtp.Auth
	sendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// NewSMTP builds an SMTP mailer. When username is empty the connection is made
// without authentication (useful for local relays).
func NewSMTP(host, port, username, password, from string) *SMTP {
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return &SMTP{
		addr:     host + ":" + port,
		host:     host,
		from:     from,
		auth:     auth,
		sendMail: smtp.SendMail,
	}
}

// Send delivers the message.
func (m *SMTP) Send(_ context.Context, msg domain.Email) error {
	if err := m.sendMail(m.addr, m.auth, m.from, []string{msg.To}, buildMessage(m.from, msg)); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

// buildMessage renders a minimal RFC 5322 message with CRLF line endings.
func buildMessage(from string, msg domain.Email) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	// Normalise the body to CRLF.
	b.WriteString(strings.ReplaceAll(msg.Body, "\n", "\r\n"))
	return []byte(b.String())
}
