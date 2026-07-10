package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// Mock simulates the hosted-checkout dance in development (no iyzico
// credentials needed): InitiateCheckout points the browser at the SPA's
// /mock-pay page, whose buttons post the token back to the real callback
// endpoint. A token suffixed ":fail" verifies as a failed payment.
type Mock struct {
	appBaseURL string
}

// NewMock builds the development gateway.
func NewMock(appBaseURL string) *Mock {
	return &Mock{appBaseURL: strings.TrimRight(appBaseURL, "/")}
}

// InitiateCheckout returns a fake hosted-payment page served by the SPA.
func (m *Mock) InitiateCheckout(_ context.Context, order *domain.Order, _ *domain.User, _ string) (*domain.CheckoutSession, error) {
	token := fmt.Sprintf("mock-%d-%d", order.ID, time.Now().UnixNano())
	return &domain.CheckoutSession{
		Token:          token,
		PaymentPageURL: m.appBaseURL + "/mock-pay?token=" + token,
	}, nil
}

// VerifyCheckout succeeds unless the token carries the ":fail" suffix chosen
// on the mock payment page.
func (m *Mock) VerifyCheckout(_ context.Context, token string) (*domain.PaymentResult, error) {
	if base, failed := strings.CutSuffix(token, ":fail"); failed {
		return &domain.PaymentResult{Token: base, Paid: false}, nil
	}
	return &domain.PaymentResult{Token: token, Paid: true, PaymentID: "mockpay-" + token}, nil
}

// Refund always succeeds.
func (m *Mock) Refund(context.Context, string) error { return nil }
