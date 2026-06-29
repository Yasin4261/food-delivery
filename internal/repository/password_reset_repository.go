package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// PasswordResetRepository is the PostgreSQL adapter for
// domain.PasswordResetRepository.
type PasswordResetRepository struct {
	db *sql.DB
}

// NewPasswordResetRepository builds a PasswordResetRepository.
func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

// Create inserts a reset token and back-fills its id and created_at.
func (r *PasswordResetRepository) Create(ctx context.Context, t *domain.PasswordResetToken) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, t.UserID, t.TokenHash, t.ExpiresAt).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}
	return nil
}

// FindByHash returns the token with the given hash, or ErrResetTokenNotFound.
func (r *PasswordResetRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	t := &domain.PasswordResetToken{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens WHERE token_hash = $1`, tokenHash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrResetTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find reset token: %w", err)
	}
	return t, nil
}

// MarkUsed stamps used_at, consuming the token.
func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark reset token used: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrResetTokenNotFound
	}
	return nil
}
