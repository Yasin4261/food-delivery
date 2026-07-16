package domain

import (
	"strings"
	"time"
)

// Promo discount types.
const (
	PromoPercent = "percent"
	PromoFixed   = "fixed"
)

// PromoCode is a platform-funded checkout discount (mirrors promo_codes,
// migrations/000022). The discount is percentage or fixed off the food
// subtotal; the chef's earnings are unaffected.
type PromoCode struct {
	ID   int    `json:"id"`
	Code string `json:"code"`

	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	MinOrder      float64 `json:"min_order"`

	ValidFrom  *time.Time `json:"valid_from,omitempty"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`

	UsageLimit int  `json:"usage_limit"` // 0 = unlimited
	UsedCount  int  `json:"used_count"`
	IsActive   bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
}

// NormaliseCode upper-cases and trims a code so lookups are case-insensitive.
func NormaliseCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// Validate checks a code definition (used on admin create).
func (p *PromoCode) Validate() error {
	p.Code = NormaliseCode(p.Code)
	if p.Code == "" {
		return ErrPromoCodeRequired
	}
	if p.DiscountType != PromoPercent && p.DiscountType != PromoFixed {
		return ErrPromoInvalid
	}
	if p.DiscountValue <= 0 {
		return ErrPromoInvalid
	}
	if p.DiscountType == PromoPercent && p.DiscountValue > 100 {
		return ErrPromoInvalid
	}
	if p.MinOrder < 0 || p.UsageLimit < 0 {
		return ErrPromoInvalid
	}
	if p.ValidFrom != nil && p.ValidUntil != nil && p.ValidUntil.Before(*p.ValidFrom) {
		return ErrPromoInvalid
	}
	return nil
}

// Redeemable reports whether the code may be applied to an order with the given
// food subtotal at time t. It returns the specific reason error when not.
func (p *PromoCode) Redeemable(subtotal float64, t time.Time) error {
	if !p.IsActive {
		return ErrPromoNotRedeemable
	}
	if p.ValidFrom != nil && t.Before(*p.ValidFrom) {
		return ErrPromoNotRedeemable
	}
	if p.ValidUntil != nil && t.After(*p.ValidUntil) {
		return ErrPromoExpired
	}
	if p.UsageLimit > 0 && p.UsedCount >= p.UsageLimit {
		return ErrPromoUsedUp
	}
	if subtotal < p.MinOrder {
		return ErrPromoMinOrder
	}
	return nil
}

// DiscountFor returns the money discount for the given food subtotal, never
// more than the subtotal itself (a code can't make an order negative).
func (p *PromoCode) DiscountFor(subtotal float64) float64 {
	var d float64
	switch p.DiscountType {
	case PromoPercent:
		d = subtotal * p.DiscountValue / 100
	case PromoFixed:
		d = p.DiscountValue
	}
	if d > subtotal {
		d = subtotal
	}
	return RoundMoney(d)
}
