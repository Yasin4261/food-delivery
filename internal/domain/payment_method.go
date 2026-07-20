package domain

import (
	"context"
	"time"
)

// SavedCard is a customer's stored card, persisted as opaque iyzico references
// (cardUserKey/cardToken) plus display-only metadata (mirrors payment_methods,
// migrations/000025). The raw PAN/CVC never touch this system — only the
// gateway tokens and a masked number are held.
type SavedCard struct {
	ID           int       `json:"id"`
	UserID       int       `json:"-"`
	CardUserKey  string    `json:"-"` // iyzico wallet key; never exposed
	CardToken    string    `json:"card_token"`
	MaskedNumber string    `json:"masked_number"`
	Association  string    `json:"association,omitempty"`
	Family       string    `json:"family,omitempty"`
	BankName     string    `json:"bank_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// StoredCard is the metadata a gateway reports for a card it registered during
// checkout (or lists from a wallet). It is the transport shape the payment
// service turns into a SavedCard row.
type StoredCard struct {
	CardUserKey  string
	CardToken    string
	MaskedNumber string
	Association  string
	Family       string
	BankName     string
}

// CheckoutOptions carries the saved-card intent into a hosted checkout: reuse
// an existing wallet (CardUserKey) so the customer's stored cards appear, and/or
// opt in to saving the card used this time (RegisterCard). The zero value is a
// plain one-off checkout that stores nothing.
type CheckoutOptions struct {
	CardUserKey  string
	RegisterCard bool
}

// PaymentMethodRepository is the port for saved-card persistence.
type PaymentMethodRepository interface {
	// Add stores a card, idempotent on (user_id, card_token). It back-fills the
	// row's ID and CreatedAt.
	Add(ctx context.Context, c *SavedCard) error
	// ListByUser returns a user's saved cards, newest first.
	ListByUser(ctx context.Context, userID int) ([]*SavedCard, error)
	// FindByToken returns one of the user's saved cards, or ErrCardNotFound.
	FindByToken(ctx context.Context, userID int, cardToken string) (*SavedCard, error)
	// CardUserKey returns the user's iyzico wallet key (any saved card's key),
	// or "" when the user has no stored cards yet.
	CardUserKey(ctx context.Context, userID int) (string, error)
	// Delete removes one of the user's saved cards. Owner-scoped: deleting a
	// token that is not the caller's is a no-op returning ErrCardNotFound.
	Delete(ctx context.Context, userID int, cardToken string) error
}
