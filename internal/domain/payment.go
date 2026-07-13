package domain

import (
	"context"
	"time"
)

// PaymentSession tracks one hosted-checkout attempt for an order (mirrors
// payment_sessions, migrations/000014). An order may accumulate several
// sessions (failed attempts, retries); at most one ends up paid.
type PaymentSession struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Token     string    `json:"-"` // gateway checkout token; never exposed
	PaymentID *string   `json:"-"` // gateway payment id once paid
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Payment session statuses.
const (
	PaymentSessionInitiated = "initiated"
	PaymentSessionPaid      = "paid"
	PaymentSessionFailed    = "failed"
	PaymentSessionRefunded  = "refunded"
)

// CheckoutSession is what a gateway returns when a hosted checkout is opened:
// the customer's browser is sent to PaymentPageURL; Token identifies the
// attempt on the way back.
type CheckoutSession struct {
	Token          string
	PaymentPageURL string
}

// PaymentResult is the outcome of verifying a checkout token with the gateway
// (a server-to-server call — the browser-supplied token is never trusted on
// its own).
type PaymentResult struct {
	Token     string
	Paid      bool
	PaymentID string
}

// PaymentGateway is the port for card processing. Adapters live in
// internal/payment: iyzico (hosted Checkout Form) for real charges and a mock
// for development.
type PaymentGateway interface {
	// InitiateCheckout opens a hosted checkout for the order and returns where
	// to send the customer's browser.
	InitiateCheckout(ctx context.Context, order *Order, buyer *User, callbackURL string) (*CheckoutSession, error)
	// VerifyCheckout resolves a callback token to a payment outcome.
	VerifyCheckout(ctx context.Context, token string) (*PaymentResult, error)
	// Refund returns a captured payment (full amount).
	Refund(ctx context.Context, paymentID string) error
	// RefundPartial returns part of a captured payment — a declined sub-order's
	// subtotal in a multi-chef order that otherwise stays alive.
	RefundPartial(ctx context.Context, paymentID string, amount float64) error
}

// PaymentSessionRepository is the port for payment-session persistence.
type PaymentSessionRepository interface {
	Create(ctx context.Context, s *PaymentSession) error
	// FindByToken returns the session for a checkout token, or
	// ErrPaymentSessionNotFound.
	FindByToken(ctx context.Context, token string) (*PaymentSession, error)
	// FindPaidByOrder returns the paid session of an order (for refunds).
	FindPaidByOrder(ctx context.Context, orderID int) (*PaymentSession, error)
	// UpdateStatus sets the session status and, when non-nil, the gateway
	// payment id.
	UpdateStatus(ctx context.Context, id int, status string, paymentID *string) error
}

// PaymentRefunder refunds an order's captured card payment. Implemented by the
// payment service; consumed by the order service when a paid order is
// cancelled (full) or one of its sub-orders is declined (partial).
type PaymentRefunder interface {
	RefundOrderPayment(ctx context.Context, order *Order) error
	// RefundSubOrderPayment returns amount of the order's captured payment —
	// the declined chef's slice. The payment session stays paid: the remaining
	// sub-orders' money is still captured.
	RefundSubOrderPayment(ctx context.Context, order *Order, amount float64) error
}
