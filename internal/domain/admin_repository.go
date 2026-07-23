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
	// The change and the audit entry are written in one transaction. Returns
	// ErrUserNotFound when no row matches.
	SetUserActive(ctx context.Context, e *AuditEntry, userID int, active bool) error
	// ListChefs returns a page of ALL chefs (including inactive/hidden)
	// matching f, newest first, plus the total matching count.
	ListChefs(ctx context.Context, f AdminChefFilters, limit, offset int) ([]*Chef, int, error)
	// SetChefActive toggles a chef's active flag (deactivation hides them from
	// browse/search and blocks new orders). Atomic with its audit entry.
	// Returns ErrChefNotFound.
	SetChefActive(ctx context.Context, e *AuditEntry, chefID int, active bool) error
	// SetChefOnline / SetChefAcceptingOrders let an admin drive a chef's
	// presence / availability on the chef's behalf (support). Atomic with audit.
	SetChefOnline(ctx context.Context, e *AuditEntry, chefID int, online bool) error
	SetChefAcceptingOrders(ctx context.Context, e *AuditEntry, chefID int, accepting bool) error
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

	// --- promo management, each atomic with an audit entry (#122) ---

	// CreatePromo inserts a code (backfilling p.ID) + audit. ErrPromoExists on
	// a duplicate code.
	CreatePromo(ctx context.Context, e *AuditEntry, p *PromoCode) error
	// UpdatePromo edits a code's definition + audit. ErrPromoNotFound if absent.
	UpdatePromo(ctx context.Context, e *AuditEntry, p *PromoCode) error
	// DeletePromo removes a code + audit. ErrPromoNotFound if absent.
	DeletePromo(ctx context.Context, e *AuditEntry, id int) error
	// SetPromoActive toggles a code's active flag + audit.
	SetPromoActive(ctx context.Context, e *AuditEntry, id int, active bool) error
	// ListPromos returns a page of all codes (newest first) for the admin surface.
	ListPromos(ctx context.Context, limit, offset int) ([]*PromoCode, int, error)
	// FindPromo returns one code by id, or ErrPromoNotFound.
	FindPromo(ctx context.Context, id int) (*PromoCode, error)

	// --- audit log (read-only) ---

	// ListAudit returns a page of the audit log matching f, newest first.
	ListAudit(ctx context.Context, f AuditFilters, limit, offset int) ([]*AuditEntry, int, error)
}
