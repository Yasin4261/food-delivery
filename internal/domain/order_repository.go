package domain

import "context"

// OrderRepository is the port for order persistence. Lookups return
// ErrOrderNotFound when no row matches.
type OrderRepository interface {
	// Create persists an order together with its items atomically (in a single
	// transaction) and back-fills generated ids and timestamps.
	Create(ctx context.Context, order *Order) error
	// FindByID returns an order with all of its items.
	FindByID(ctx context.Context, id int) (*Order, error)
	// ListByUser returns a page of a customer's orders (newest first, each with
	// all its items) and the total order count.
	ListByUser(ctx context.Context, userID, limit, offset int) ([]*Order, int, error)
	// ListByChef returns a page of orders containing at least one of the chef's
	// items (newest first, items filtered to that chef) and the total count.
	ListByChef(ctx context.Context, chefID, limit, offset int) ([]*Order, int, error)
	// UpdateStatus persists the mutable status/payment/timestamp fields of an
	// order after a transition.
	UpdateStatus(ctx context.Context, order *Order) error
}
