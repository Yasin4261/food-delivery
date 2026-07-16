package domain

import "context"

// PromoRepository is the port for promo-code persistence.
type PromoRepository interface {
	// Create persists a new code. A duplicate code returns ErrPromoExists.
	Create(ctx context.Context, p *PromoCode) error
	// FindByCode returns a code by its normalised name, or ErrPromoNotFound.
	FindByCode(ctx context.Context, code string) (*PromoCode, error)
	// Redeem atomically increments used_count when the usage limit still
	// allows it, returning ErrPromoUsedUp when the cap is already reached.
	Redeem(ctx context.Context, id int) error
	// List returns all codes (newest first) for the admin surface.
	List(ctx context.Context, limit, offset int) ([]*PromoCode, int, error)
	// SetActive toggles a code's active flag.
	SetActive(ctx context.Context, id int, active bool) error
}
