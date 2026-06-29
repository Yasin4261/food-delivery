package domain

import "context"

// FavoriteRepository is the port for favorite persistence.
type FavoriteRepository interface {
	// Add favorites a chef for a user. It is idempotent: favoriting an
	// already-favorited chef is a no-op (relies on the unique constraint).
	Add(ctx context.Context, userID, chefID int) error
	// Remove unfavorites a chef. Removing a non-favorite is a no-op.
	Remove(ctx context.Context, userID, chefID int) error
	// ListChefs returns the active chefs a user has favorited, newest first.
	ListChefs(ctx context.Context, userID, limit, offset int) ([]*Chef, error)
}
