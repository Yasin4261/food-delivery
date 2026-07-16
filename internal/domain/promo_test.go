package domain

import (
	"testing"
	"time"
)

func TestPromoCode_Validate(t *testing.T) {
	valid := func() *PromoCode {
		return &PromoCode{Code: "save10", DiscountType: PromoPercent, DiscountValue: 10}
	}
	cases := map[string]struct {
		mutate func(*PromoCode)
		want   error
	}{
		"ok":              {func(*PromoCode) {}, nil},
		"empty code":      {func(p *PromoCode) { p.Code = "  " }, ErrPromoCodeRequired},
		"bad type":        {func(p *PromoCode) { p.DiscountType = "half" }, ErrPromoInvalid},
		"zero value":      {func(p *PromoCode) { p.DiscountValue = 0 }, ErrPromoInvalid},
		"percent over100": {func(p *PromoCode) { p.DiscountValue = 150 }, ErrPromoInvalid},
		"negative min":    {func(p *PromoCode) { p.MinOrder = -1 }, ErrPromoInvalid},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := valid()
			tc.mutate(p)
			if err := p.Validate(); err != tc.want {
				t.Errorf("Validate = %v, want %v", err, tc.want)
			}
		})
	}

	// Validate normalises the code.
	p := &PromoCode{Code: " save10 ", DiscountType: PromoFixed, DiscountValue: 5}
	_ = p.Validate()
	if p.Code != "SAVE10" {
		t.Errorf("code = %q, want normalised SAVE10", p.Code)
	}
}

func TestPromoCode_Redeemable(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	base := &PromoCode{IsActive: true, DiscountType: PromoPercent, DiscountValue: 10, MinOrder: 50, UsageLimit: 2, UsedCount: 0}

	if err := base.Redeemable(60, now); err != nil {
		t.Errorf("valid code = %v, want nil", err)
	}
	if err := base.Redeemable(40, now); err != ErrPromoMinOrder {
		t.Errorf("below min = %v, want ErrPromoMinOrder", err)
	}

	inactive := *base
	inactive.IsActive = false
	if err := inactive.Redeemable(60, now); err != ErrPromoNotRedeemable {
		t.Errorf("inactive = %v, want ErrPromoNotRedeemable", err)
	}

	expired := *base
	expired.ValidUntil = &past
	if err := expired.Redeemable(60, now); err != ErrPromoExpired {
		t.Errorf("expired = %v, want ErrPromoExpired", err)
	}

	notYet := *base
	notYet.ValidFrom = &future
	if err := notYet.Redeemable(60, now); err != ErrPromoNotRedeemable {
		t.Errorf("not yet valid = %v, want ErrPromoNotRedeemable", err)
	}

	usedUp := *base
	usedUp.UsedCount = 2
	if err := usedUp.Redeemable(60, now); err != ErrPromoUsedUp {
		t.Errorf("used up = %v, want ErrPromoUsedUp", err)
	}
}

func TestPromoCode_DiscountFor(t *testing.T) {
	pct := &PromoCode{DiscountType: PromoPercent, DiscountValue: 15}
	if got := pct.DiscountFor(200); got != 30 {
		t.Errorf("percent = %v, want 30", got)
	}
	fixed := &PromoCode{DiscountType: PromoFixed, DiscountValue: 20}
	if got := fixed.DiscountFor(200); got != 20 {
		t.Errorf("fixed = %v, want 20", got)
	}
	// Fixed can't exceed the subtotal.
	if got := fixed.DiscountFor(12); got != 12 {
		t.Errorf("fixed over subtotal = %v, want capped at 12", got)
	}
}
