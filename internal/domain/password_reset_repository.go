package domain

import "context"

// PasswordResetRepository is the port for password-reset token persistence.
type PasswordResetRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	// FindByHash returns the token with the given hash, or
	// ErrResetTokenNotFound.
	FindByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	// MarkUsed stamps used_at, consuming the token.
	MarkUsed(ctx context.Context, id int) error
}
