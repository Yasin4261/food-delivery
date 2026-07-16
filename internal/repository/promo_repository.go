package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// PromoRepository is the PostgreSQL adapter for domain.PromoRepository.
type PromoRepository struct {
	db *sql.DB
}

// NewPromoRepository builds a PromoRepository.
func NewPromoRepository(db *sql.DB) *PromoRepository {
	return &PromoRepository{db: db}
}

const promoColumns = `id, code, discount_type, discount_value, min_order,
	valid_from, valid_until, usage_limit, used_count, is_active, created_at`

func scanPromo(s interface{ Scan(...any) error }) (*domain.PromoCode, error) {
	p := &domain.PromoCode{}
	err := s.Scan(
		&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.MinOrder,
		&p.ValidFrom, &p.ValidUntil, &p.UsageLimit, &p.UsedCount, &p.IsActive, &p.CreatedAt,
	)
	return p, err
}

// Create persists a new code.
func (r *PromoRepository) Create(ctx context.Context, p *domain.PromoCode) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO promo_codes (code, discount_type, discount_value, min_order, valid_from, valid_until, usage_limit, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, used_count, created_at`,
		p.Code, p.DiscountType, p.DiscountValue, p.MinOrder, p.ValidFrom, p.ValidUntil, p.UsageLimit, p.IsActive,
	).Scan(&p.ID, &p.UsedCount, &p.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return domain.ErrPromoExists
		}
		return fmt.Errorf("create promo: %w", err)
	}
	return nil
}

// FindByCode returns a code by its normalised name.
func (r *PromoRepository) FindByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+promoColumns+` FROM promo_codes WHERE code = $1`, domain.NormaliseCode(code))
	p, err := scanPromo(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPromoNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find promo: %w", err)
	}
	return p, nil
}

// Redeem atomically bumps used_count while the limit still allows it. The
// guarded UPDATE makes the cap race-free under concurrent orders.
func (r *PromoRepository) Redeem(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE promo_codes SET used_count = used_count + 1
		WHERE id = $1 AND (usage_limit = 0 OR used_count < usage_limit)`, id)
	if err != nil {
		return fmt.Errorf("redeem promo: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrPromoUsedUp
	}
	return nil
}

// List returns all codes, newest first, plus the total.
func (r *PromoRepository) List(ctx context.Context, limit, offset int) ([]*domain.PromoCode, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+promoColumns+` FROM promo_codes ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list promos: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.PromoCode, 0)
	for rows.Next() {
		p, err := scanPromo(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan promo: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM promo_codes`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count promos: %w", err)
	}
	return out, total, nil
}

// SetActive toggles a code's active flag.
func (r *PromoRepository) SetActive(ctx context.Context, id int, active bool) error {
	res, err := r.db.ExecContext(ctx, `UPDATE promo_codes SET is_active = $2 WHERE id = $1`, id, active)
	if err != nil {
		return fmt.Errorf("set promo active: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrPromoNotFound
	}
	return nil
}
