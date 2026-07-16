package domain

import "context"

// ReviewRepository is the port for review persistence. Create also recomputes
// the target's aggregate rating (chefs.rating / menu_items.rating) atomically,
// since that derived value is owned by the review data.
type ReviewRepository interface {
	// Create persists a review and recomputes the reviewed chef's or dish's
	// aggregate rating in the same transaction. A duplicate (same user, order
	// and target) returns ErrReviewExists.
	Create(ctx context.Context, review *Review) error
	// ListByChef returns a page of a chef's reviews (newest first) and the total.
	ListByChef(ctx context.Context, chefID, limit, offset int) ([]*Review, int, error)
	// ListByMenuItem returns a page of a dish's reviews (newest first) and the total.
	ListByMenuItem(ctx context.Context, menuItemID, limit, offset int) ([]*Review, int, error)
	// ListByUserOrder returns the caller's own reviews for one order (their
	// rating history for that order) — scoped to the user by construction.
	ListByUserOrder(ctx context.Context, userID, orderID int) ([]*Review, error)
	// ListByUser returns every review the user has written (newest first), for
	// the data export (#107).
	ListByUser(ctx context.Context, userID int) ([]*Review, error)
}
