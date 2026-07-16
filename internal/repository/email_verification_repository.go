package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// EmailVerificationRepository is the PostgreSQL adapter for
// domain.EmailVerificationRepository.
type EmailVerificationRepository struct {
	db *sql.DB
}

// NewEmailVerificationRepository builds an EmailVerificationRepository.
func NewEmailVerificationRepository(db *sql.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

// Create inserts a verification token and back-fills its id and created_at.
func (r *EmailVerificationRepository) Create(ctx context.Context, t *domain.EmailVerificationToken) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, t.UserID, t.TokenHash, t.ExpiresAt).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("create verification token: %w", err)
	}
	return nil
}

// FindByHash returns the token with the given hash, or
// ErrVerificationTokenNotFound.
func (r *EmailVerificationRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.EmailVerificationToken, error) {
	t := &domain.EmailVerificationToken{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM email_verification_tokens WHERE token_hash = $1`, tokenHash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrVerificationTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find verification token: %w", err)
	}
	return t, nil
}

// MarkUsed stamps used_at, consuming the token.
func (r *EmailVerificationRepository) MarkUsed(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE email_verification_tokens SET used_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark verification token used: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrVerificationTokenNotFound
	}
	return nil
}
