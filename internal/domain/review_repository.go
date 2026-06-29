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
	// ListByChef returns reviews of a chef, newest first.
	ListByChef(ctx context.Context, chefID, limit, offset int) ([]*Review, error)
	// ListByMenuItem returns reviews of a dish, newest first.
	ListByMenuItem(ctx context.Context, menuItemID, limit, offset int) ([]*Review, error)
}
