package domain

import "context"

// Email is a plain-text transactional message.
type Email struct {
	To      string
	Subject string
	Body    string
}

// Mailer is the port for sending transactional email. It is a driven adapter:
// implementations (SMTP, a logging stub for development, a fake for tests) live
// in internal/mailer. The core composes the message; the adapter delivers it.
type Mailer interface {
	Send(ctx context.Context, msg Email) error
}
