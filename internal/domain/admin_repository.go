package domain

import "context"

// AdminRepository is the port for admin-only reads and moderation actions.
// All admin SQL (cross-entity listings, aggregation) lives in one adapter.
type AdminRepository interface {
	// ListUsers returns a page of ALL users (including inactive) matching f,
	// newest first, plus the total matching count. Password hashes are cleared
	// by the service.
	ListUsers(ctx context.Context, f AdminUserFilters, limit, offset int) ([]*User, int, error)
	// SetUserActive toggles a user's active flag (deactivation blocks login).
	// Returns ErrUserNotFound when no row matches.
	SetUserActive(ctx context.Context, userID int, active bool) error
	// ListChefs returns a page of ALL chefs (including inactive/hidden)
	// matching f, newest first, plus the total matching count.
	ListChefs(ctx context.Context, f AdminChefFilters, limit, offset int) ([]*Chef, int, error)
	// SetChefActive toggles a chef's active flag (deactivation hides them from
	// browse/search and blocks new orders). Returns ErrChefNotFound.
	SetChefActive(ctx context.Context, chefID int, active bool) error
	// ListOrders returns a page of ALL orders (with items) matching f, newest
	// first, plus the total matching count — the platform-wide order overview.
	ListOrders(ctx context.Context, f AdminOrderFilters, limit, offset int) ([]*Order, int, error)
	// UserDetail returns one account with its kitchen, recent orders and
	// reviews — the support console's drill-in. ErrUserNotFound if absent.
	UserDetail(ctx context.Context, userID int) (*AdminUserDetail, error)
	// OrderDetail returns one order with its items, sub-orders, customer and
	// payment attempts. ErrOrderNotFound if absent.
	OrderDetail(ctx context.Context, orderID int) (*AdminOrderDetail, error)
	// ChefDetail returns one kitchen with its owner, dishes and recent orders.
	// Resolves inactive chefs too (support must be able to inspect a
	// deactivated kitchen). ErrChefNotFound if absent.
	ChefDetail(ctx context.Context, chefID int) (*AdminChefDetail, error)
	// Stats returns the aggregated platform dashboard figures.
	Stats(ctx context.Context) (*PlatformStats, error)
}
