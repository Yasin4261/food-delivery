package payment

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// Mock simulates the hosted-checkout dance in development (no iyzico
// credentials needed): InitiateCheckout points the browser at the SPA's
// /mock-pay page, whose buttons post the token back to the real callback
// endpoint. A token suffixed ":fail" verifies as a failed payment.
//
// It also simulates iyzico card storage (#67): when a checkout opts to register
// the card, VerifyCheckout returns a synthetic StoredCard so the saved-card
// flow can be exercised end-to-end without a real gateway.
type Mock struct {
	appBaseURL string

	mu      sync.Mutex
	pending map[string]mockCheckout // token -> save intent for this attempt
}

// mockCheckout remembers, for a token, whether the customer opted to save the
// card and which wallet key to attribute it to.
type mockCheckout struct {
	registerCard bool
	cardUserKey  string
}

// NewMock builds the development gateway.
func NewMock(appBaseURL string) *Mock {
	return &Mock{
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
		pending:    map[string]mockCheckout{},
	}
}

// InitiateCheckout returns a fake hosted-payment page served by the SPA. It
// remembers the save intent so VerifyCheckout can mint a synthetic saved card.
func (m *Mock) InitiateCheckout(_ context.Context, order *domain.Order, buyer *domain.User, _ string, opts domain.CheckoutOptions) (*domain.CheckoutSession, error) {
	token := fmt.Sprintf("mock-%d-%d", order.ID, time.Now().UnixNano())
	if opts.RegisterCard {
		key := opts.CardUserKey
		if key == "" {
			// A stable per-user wallet key, mirroring iyzico reusing one
			// cardUserKey across a customer's saved cards.
			key = fmt.Sprintf("mockcuk-%d", buyer.ID)
		}
		m.mu.Lock()
		m.pending[token] = mockCheckout{registerCard: true, cardUserKey: key}
		m.mu.Unlock()
	}
	return &domain.CheckoutSession{
		Token:          token,
		PaymentPageURL: m.appBaseURL + "/mock-pay?token=" + token,
	}, nil
}

// VerifyCheckout succeeds unless the token carries the ":fail" suffix chosen on
// the mock payment page. When the attempt opted to register a card, a synthetic
// StoredCard is returned so the caller can persist it.
func (m *Mock) VerifyCheckout(_ context.Context, token string) (*domain.PaymentResult, error) {
	base, failed := strings.CutSuffix(token, ":fail")
	if failed {
		return &domain.PaymentResult{Token: base, Paid: false}, nil
	}

	res := &domain.PaymentResult{Token: token, Paid: true, PaymentID: "mockpay-" + token}
	m.mu.Lock()
	if pc, ok := m.pending[token]; ok && pc.registerCard {
		delete(m.pending, token)
		res.RegisteredCard = &domain.StoredCard{
			CardUserKey:  pc.cardUserKey,
			CardToken:    "mockcard-" + token,
			MaskedNumber: "552608******0006",
			Association:  "MASTER_CARD",
			Family:       "Bonus",
			BankName:     "Mock Bank",
		}
	}
	m.mu.Unlock()
	return res, nil
}

// Refund always succeeds.
func (m *Mock) Refund(context.Context, string) error { return nil }

// RefundPartial always succeeds.
func (m *Mock) RefundPartial(context.Context, string, float64) error { return nil }

// DeleteStoredCard always succeeds (nothing is stored gateway-side in the mock).
func (m *Mock) DeleteStoredCard(context.Context, string, string) error { return nil }
