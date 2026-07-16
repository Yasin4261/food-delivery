package domain

import "context"

// EmailVerificationRepository is the port for email-verification token
// persistence. The Postgres adapter lives in internal/repository.
type EmailVerificationRepository interface {
	Create(ctx context.Context, token *EmailVerificationToken) error
	// FindByHash returns the token with the given hash, or
	// ErrVerificationTokenNotFound.
	FindByHash(ctx context.Context, tokenHash string) (*EmailVerificationToken, error)
	// MarkUsed stamps used_at, consuming the token.
	MarkUsed(ctx context.Context, id int) error
}
