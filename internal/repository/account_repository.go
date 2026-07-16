package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AccountRepository is the PostgreSQL adapter for domain.AccountRepository.
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository builds an AccountRepository.
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Anonymise scrubs the user's PII and deactivates the account (and chef
// storefront, if any) in one transaction. Counterparty records — orders,
// order_items, sub_orders and reviews — are retained so the other party's
// history and earnings stay intact; only the personal data is cleared.
//
// The scrubbed email/username are made unique by the user id so the account's
// UNIQUE constraints still hold and a fresh sign-up can reuse the real address.
func (r *AccountRepository) Anonymise(ctx context.Context, userID int) error {
	// A random suffix makes the scrubbed email/username collision-proof: a
	// deterministic 'deleted-<id>' could be pre-registered by an attacker to
	// trip the UNIQUE constraint and permanently block the deletion (#113).
	suffix, err := randomSuffix()
	if err != nil {
		return err
	}
	email := fmt.Sprintf("deleted-%d-%s@removed.invalid", userID, suffix)
	username := fmt.Sprintf("deleted-%d-%s", userID, suffix)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("anonymise: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// The account itself: clear PII, blank the password hash (login already
	// blocked by is_active, this removes the credential too) and deactivate.
	res, err := tx.ExecContext(ctx, `
		UPDATE users SET
			email = $2,
			username = $3,
			phone_number = NULL,
			address = NULL, city = NULL, state = NULL, zip_code = NULL,
			latitude = NULL, longitude = NULL,
			password_hash = '',
			is_active = false,
			email_notifications = false,
			updated_at = now()
		WHERE id = $1`, userID, email, username)
	if err != nil {
		return fmt.Errorf("anonymise user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrUserNotFound
	}

	// The chef storefront, if any: scrub the kitchen identity and take it
	// offline so it disappears from browse/search and can take no orders.
	if _, err := tx.ExecContext(ctx, `
		UPDATE chefs SET
			business_name = 'Closed kitchen',
			bio = NULL, specialty = NULL, image_url = NULL,
			kitchen_address = '', kitchen_city = NULL,
			kitchen_latitude = NULL, kitchen_longitude = NULL,
			food_license_number = NULL, health_certificate_url = NULL,
			is_active = false, is_accepting_orders = false, is_online = false,
			updated_at = now()
		WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("anonymise chef: %w", err)
	}

	// Personal-only rows with no counterparty: delete outright.
	if _, err := tx.ExecContext(ctx, `DELETE FROM addresses WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("anonymise addresses: %w", err)
	}

	// Chat: keep the threads (the other party's context) but tombstone the
	// message bodies this user wrote.
	if _, err := tx.ExecContext(ctx,
		`UPDATE chat_messages SET body = '[deleted]' WHERE sender_id = $1`, userID); err != nil {
		return fmt.Errorf("anonymise chat: %w", err)
	}

	// Any outstanding reset / verification tokens are now meaningless.
	if _, err := tx.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("anonymise reset tokens: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("anonymise verification tokens: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("anonymise: commit: %w", err)
	}
	return nil
}

// randomSuffix returns 8 random bytes, hex-encoded, for collision-proof
// anonymised identifiers.
func randomSuffix() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("anonymise: random suffix: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
