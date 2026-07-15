package domain

import "context"

// AdminRepository is the port for admin-only reads and moderation actions.
// All admin SQL (cross-entity listings, aggregation) lives in one adapter.
type AdminRepository interface {
	// ListUsers returns a page of ALL users (including inactive), newest
	// first, plus the total. Password hashes are cleared by the service.
	ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error)
	// SetUserActive toggles a user's active flag (deactivation blocks login).
	// Returns ErrUserNotFound when no row matches.
	SetUserActive(ctx context.Context, userID int, active bool) error
	// ListChefs returns a page of ALL chefs (including inactive/hidden),
	// newest first, plus the total.
	ListChefs(ctx context.Context, limit, offset int) ([]*Chef, int, error)
	// SetChefActive toggles a chef's active flag (deactivation hides them from
	// browse/search and blocks new orders). Returns ErrChefNotFound.
	SetChefActive(ctx context.Context, chefID int, active bool) error
	// ListOrders returns a page of ALL orders (with items), newest first,
	// plus the total — the platform-wide order overview.
	ListOrders(ctx context.Context, limit, offset int) ([]*Order, int, error)
	// Stats returns the aggregated platform dashboard figures.
	Stats(ctx context.Context) (*PlatformStats, error)
}
