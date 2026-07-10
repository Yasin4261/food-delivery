package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// PaymentSessionRepository is the PostgreSQL adapter for
// domain.PaymentSessionRepository.
type PaymentSessionRepository struct {
	db *sql.DB
}

// NewPaymentSessionRepository builds a PaymentSessionRepository.
func NewPaymentSessionRepository(db *sql.DB) *PaymentSessionRepository {
	return &PaymentSessionRepository{db: db}
}

const paymentSessionColumns = `id, order_id, token, payment_id, status, created_at, updated_at`

func scanPaymentSession(s interface{ Scan(...any) error }) (*domain.PaymentSession, error) {
	ps := &domain.PaymentSession{}
	err := s.Scan(&ps.ID, &ps.OrderID, &ps.Token, &ps.PaymentID, &ps.Status, &ps.CreatedAt, &ps.UpdatedAt)
	return ps, err
}

// Create inserts a session and back-fills id and timestamps.
func (r *PaymentSessionRepository) Create(ctx context.Context, ps *domain.PaymentSession) error {
	if ps.Status == "" {
		ps.Status = domain.PaymentSessionInitiated
	}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO payment_sessions (order_id, token, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`, ps.OrderID, ps.Token, ps.Status).
		Scan(&ps.ID, &ps.CreatedAt, &ps.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create payment session: %w", err)
	}
	return nil
}

// FindByToken returns the session for a checkout token.
func (r *PaymentSessionRepository) FindByToken(ctx context.Context, token string) (*domain.PaymentSession, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+paymentSessionColumns+` FROM payment_sessions WHERE token = $1`, token)
	ps, err := scanPaymentSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPaymentSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find payment session: %w", err)
	}
	return ps, nil
}

// FindPaidByOrder returns the paid session of an order (for refunds).
func (r *PaymentSessionRepository) FindPaidByOrder(ctx context.Context, orderID int) (*domain.PaymentSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+paymentSessionColumns+` FROM payment_sessions
		WHERE order_id = $1 AND status = $2
		ORDER BY id DESC LIMIT 1`, orderID, domain.PaymentSessionPaid)
	ps, err := scanPaymentSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPaymentSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find paid session: %w", err)
	}
	return ps, nil
}

// UpdateStatus sets the status and, when non-nil, the gateway payment id.
func (r *PaymentSessionRepository) UpdateStatus(ctx context.Context, id int, status string, paymentID *string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE payment_sessions
		SET status = $2, payment_id = COALESCE($3, payment_id), updated_at = now()
		WHERE id = $1`, id, status, paymentID)
	if err != nil {
		return fmt.Errorf("update payment session: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrPaymentSessionNotFound
	}
	return nil
}
