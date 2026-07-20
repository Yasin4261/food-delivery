package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// PaymentMethodRepository is the PostgreSQL adapter for
// domain.PaymentMethodRepository (customer saved cards).
type PaymentMethodRepository struct {
	db *sql.DB
}

// NewPaymentMethodRepository builds a PaymentMethodRepository.
func NewPaymentMethodRepository(db *sql.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{db: db}
}

const paymentMethodColumns = `id, card_token, masked_number, association, family, bank_name, created_at`

func scanSavedCard(s interface{ Scan(...any) error }) (*domain.SavedCard, error) {
	c := &domain.SavedCard{}
	var association, family, bankName sql.NullString
	err := s.Scan(&c.ID, &c.CardToken, &c.MaskedNumber, &association, &family, &bankName, &c.CreatedAt)
	c.Association = association.String
	c.Family = family.String
	c.BankName = bankName.String
	return c, err
}

// Add stores a card idempotently on (user_id, card_token). On a repeat save the
// existing row's id/created_at are returned unchanged.
func (r *PaymentMethodRepository) Add(ctx context.Context, c *domain.SavedCard) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO payment_methods
			(user_id, card_user_key, card_token, masked_number, association, family, bank_name)
		VALUES ($1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), NULLIF($7,''))
		ON CONFLICT (user_id, card_token)
		DO UPDATE SET card_user_key = EXCLUDED.card_user_key
		RETURNING id, created_at`,
		c.UserID, c.CardUserKey, c.CardToken, c.MaskedNumber, c.Association, c.Family, c.BankName).
		Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return fmt.Errorf("add payment method: %w", err)
	}
	return nil
}

// ListByUser returns a user's saved cards, newest first.
func (r *PaymentMethodRepository) ListByUser(ctx context.Context, userID int) ([]*domain.SavedCard, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+paymentMethodColumns+` FROM payment_methods WHERE user_id = $1 ORDER BY id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}
	defer rows.Close()

	var cards []*domain.SavedCard
	for rows.Next() {
		c, err := scanSavedCard(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment method: %w", err)
		}
		c.UserID = userID
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// FindByToken returns one of the user's saved cards, or ErrCardNotFound.
func (r *PaymentMethodRepository) FindByToken(ctx context.Context, userID int, cardToken string) (*domain.SavedCard, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+paymentMethodColumns+` FROM payment_methods WHERE user_id = $1 AND card_token = $2`,
		userID, cardToken)
	c, err := scanSavedCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find payment method: %w", err)
	}
	c.UserID = userID
	return c, nil
}

// CardUserKey returns the user's iyzico wallet key, or "" when they have none.
func (r *PaymentMethodRepository) CardUserKey(ctx context.Context, userID int) (string, error) {
	var key string
	err := r.db.QueryRowContext(ctx,
		`SELECT card_user_key FROM payment_methods WHERE user_id = $1 ORDER BY id DESC LIMIT 1`, userID).
		Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find card user key: %w", err)
	}
	return key, nil
}

// Delete removes one of the user's saved cards (owner-scoped by user_id).
func (r *PaymentMethodRepository) Delete(ctx context.Context, userID int, cardToken string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM payment_methods WHERE user_id = $1 AND card_token = $2`, userID, cardToken)
	if err != nil {
		return fmt.Errorf("delete payment method: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrCardNotFound
	}
	return nil
}
