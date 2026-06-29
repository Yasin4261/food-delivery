package mailer

import (
	"context"
	"log/slog"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// Logging is a development/no-op mailer: it logs the message (so the reset link
// is visible in dev) instead of sending it. It implements domain.Mailer.
type Logging struct {
	logger *slog.Logger
}

// NewLogging builds a logging mailer.
func NewLogging(logger *slog.Logger) *Logging {
	if logger == nil {
		logger = slog.Default()
	}
	return &Logging{logger: logger}
}

// Send logs the message at info level and reports success.
func (m *Logging) Send(ctx context.Context, msg domain.Email) error {
	m.logger.InfoContext(ctx, "email (dev mailer; not delivered)",
		slog.String("to", msg.To),
		slog.String("subject", msg.Subject),
		slog.String("body", msg.Body),
	)
	return nil
}
