package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AdminRepository is the PostgreSQL adapter for domain.AdminRepository. It
// reuses the column lists and scanners of the other adapters and keeps the
// cross-entity listings + aggregation SQL in one place.
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository builds an AdminRepository.
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// userFilterWhere matches the admin user filters with bound placeholders only
// (no interpolation): $1 free-text over email/username, $2 role, $3 active
// tri-state (NULL = both).
const userFilterWhere = `
	WHERE ($1 = '' OR email ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%')
	  AND ($2 = '' OR role = $2)
	  AND ($3::boolean IS NULL OR is_active = $3)`

// ListUsers returns a page of all users (including inactive) matching f,
// newest first.
func (r *AdminRepository) ListUsers(ctx context.Context, f domain.AdminUserFilters, limit, offset int) ([]*domain.User, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT`+userColumns+` FROM users`+userFilterWhere+`
		ORDER BY created_at DESC, id DESC LIMIT $4 OFFSET $5`,
		f.Query, f.Role, f.Active, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM users`+userFilterWhere, f.Query, f.Role, f.Active).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	return users, total, nil
}

// SetUserActive toggles a user's active flag.
func (r *AdminRepository) SetUserActive(ctx context.Context, userID int, active bool) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = $2, updated_at = now() WHERE id = $1`, userID, active)
	if err != nil {
		return fmt.Errorf("set user active: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// chefFilterWhere matches the admin chef filters with bound placeholders only:
// $1 free-text over the business name, $2 active tri-state (NULL = both).
const chefFilterWhere = `
	WHERE ($1 = '' OR business_name ILIKE '%' || $1 || '%')
	  AND ($2::boolean IS NULL OR is_active = $2)`

// ListChefs returns a page of all chefs (including inactive) matching f,
// newest first.
func (r *AdminRepository) ListChefs(ctx context.Context, f domain.AdminChefFilters, limit, offset int) ([]*domain.Chef, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT`+chefColumns+` FROM chefs`+chefFilterWhere+`
		ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4`,
		f.Query, f.Active, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list chefs: %w", err)
	}
	defer rows.Close()
	chefs, err := collectChefs(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM chefs`+chefFilterWhere, f.Query, f.Active).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chefs: %w", err)
	}
	return chefs, total, nil
}

// SetChefActive toggles a chef's active flag.
func (r *AdminRepository) SetChefActive(ctx context.Context, chefID int, active bool) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE chefs SET is_active = $2, updated_at = now() WHERE id = $1`, chefID, active)
	if err != nil {
		return fmt.Errorf("set chef active: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrChefNotFound
	}
	return nil
}

// orderFilterWhere matches the admin order filters with bound placeholders
// only: $1 status, $2 payment status, $3 customer id (0 = any).
const orderFilterWhere = `
	WHERE ($1 = '' OR status = $1)
	  AND ($2 = '' OR payment_status = $2)
	  AND ($3 = 0 OR user_id = $3)`

// ListOrders returns a page of all orders (with items + sub-orders) matching
// f, newest first — the platform-wide overview.
func (r *AdminRepository) ListOrders(ctx context.Context, f domain.AdminOrderFilters, limit, offset int) ([]*domain.Order, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT`+orderColumns+` FROM orders`+orderFilterWhere+`
		ORDER BY created_at DESC, id DESC LIMIT $4 OFFSET $5`,
		f.Status, f.PaymentStatus, f.UserID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list orders: %w", err)
	}
	orders, err := collectOrders(rows)
	if err != nil {
		return nil, 0, err
	}
	orderRepo := &OrderRepository{db: r.db}
	for _, o := range orders {
		if o.Items, err = orderRepo.loadItems(ctx, o.ID, 0); err != nil {
			return nil, 0, err
		}
		if o.SubOrders, err = orderRepo.loadSubOrders(ctx, o.ID); err != nil {
			return nil, 0, err
		}
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM orders`+orderFilterWhere,
		f.Status, f.PaymentStatus, f.UserID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}
	return orders, total, nil
}

// Stats aggregates the platform dashboard figures in a handful of queries.
func (r *AdminRepository) Stats(ctx context.Context) (*domain.PlatformStats, error) {
	s := &domain.PlatformStats{}

	// Scalar counters + GMV (delivered & paid) in one round trip.
	err := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT count(*) FROM users),
			(SELECT count(*) FROM chefs),
			(SELECT count(*) FROM chefs WHERE is_active),
			(SELECT count(*) FROM orders),
			(SELECT count(*) FROM orders WHERE status = 'delivered'),
			(SELECT count(*) FROM orders WHERE created_at >= date_trunc('day', now())),
			(SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'delivered' AND payment_status = 'paid')
	`).Scan(&s.TotalUsers, &s.TotalChefs, &s.ActiveChefs, &s.TotalOrders,
		&s.DeliveredOrders, &s.OrdersToday, &s.GMV)
	if err != nil {
		return nil, fmt.Errorf("platform stats: %w", err)
	}

	// Top chefs by delivered & paid revenue (from their own sub-order slices).
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.business_name,
		       COUNT(DISTINCT oi.order_id) AS orders,
		       COALESCE(SUM(oi.subtotal), 0) AS revenue
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN sub_orders so ON so.order_id = oi.order_id AND so.chef_id = oi.chef_id
		JOIN chefs c ON c.id = oi.chef_id
		WHERE so.status = 'delivered' AND o.payment_status = 'paid'
		GROUP BY c.id, c.business_name
		ORDER BY revenue DESC, orders DESC
		LIMIT 5`)
	if err != nil {
		return nil, fmt.Errorf("top chefs: %w", err)
	}
	defer rows.Close()

	s.TopChefs = make([]domain.TopChef, 0, 5)
	for rows.Next() {
		var t domain.TopChef
		if err := rows.Scan(&t.ChefID, &t.BusinessName, &t.Orders, &t.Revenue); err != nil {
			return nil, fmt.Errorf("scan top chef: %w", err)
		}
		s.TopChefs = append(s.TopChefs, t)
	}
	return s, rows.Err()
}

// --- detail views (#119): read-only support drill-in -----------------------

// UserDetail returns one account with its kitchen (if any), recent orders and
// the reviews it wrote. Nested lists are capped at domain.AdminDetailLimit.
func (r *AdminRepository) UserDetail(ctx context.Context, userID int) (*domain.AdminUserDetail, error) {
	users := &UserRepository{db: r.db}
	user, err := users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := &domain.AdminUserDetail{User: user, Orders: []*domain.Order{}, Reviews: []*domain.Review{}}

	// The kitchen this user owns, if they are a chef. Resolved without the
	// is_active filter so a deactivated kitchen is still inspectable.
	chef, err := r.findChefByUserID(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrChefNotFound) {
		return nil, err
	}
	out.Chef = chef

	orders := &OrderRepository{db: r.db}
	list, _, err := orders.ListByUser(ctx, userID, domain.AdminDetailLimit, 0)
	if err != nil {
		return nil, err
	}
	if list != nil {
		out.Orders = list
	}

	reviews := &ReviewRepository{db: r.db}
	written, err := reviews.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(written) > domain.AdminDetailLimit {
		written = written[:domain.AdminDetailLimit]
	}
	if written != nil {
		out.Reviews = written
	}
	return out, nil
}

// OrderDetail returns one order with items, sub-orders, its customer and every
// payment attempt made against it.
func (r *AdminRepository) OrderDetail(ctx context.Context, orderID int) (*domain.AdminOrderDetail, error) {
	orders := &OrderRepository{db: r.db}
	order, err := orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	out := &domain.AdminOrderDetail{Order: order, Payments: []*domain.PaymentSession{}}

	users := &UserRepository{db: r.db}
	if customer, err := users.FindByID(ctx, order.UserID); err == nil {
		out.Customer = customer
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Every attempt, newest first — a failed-then-retried payment is exactly
	// what a support ticket needs to see.
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+paymentSessionColumns+` FROM payment_sessions WHERE order_id = $1 ORDER BY id DESC`, orderID)
	if err != nil {
		return nil, fmt.Errorf("admin order payments: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		ps, err := scanPaymentSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan payment session: %w", err)
		}
		out.Payments = append(out.Payments, ps)
	}
	return out, rows.Err()
}

// ChefDetail returns one kitchen with its owner, dishes and recent orders.
func (r *AdminRepository) ChefDetail(ctx context.Context, chefID int) (*domain.AdminChefDetail, error) {
	chef, err := r.findChefByID(ctx, chefID)
	if err != nil {
		return nil, err
	}
	out := &domain.AdminChefDetail{Chef: chef, Items: []*domain.MenuItem{}, Orders: []*domain.Order{}}

	users := &UserRepository{db: r.db}
	if owner, err := users.FindByID(ctx, chef.UserID); err == nil {
		out.Owner = owner
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	items := &MenuItemRepository{db: r.db}
	dishes, _, err := items.ListByChef(ctx, chefID, domain.AdminDetailLimit, 0)
	if err != nil {
		return nil, err
	}
	if dishes != nil {
		out.Items = dishes
	}

	orders := &OrderRepository{db: r.db}
	list, _, err := orders.ListByChef(ctx, chefID, domain.AdminDetailLimit, 0)
	if err != nil {
		return nil, err
	}
	if list != nil {
		out.Orders = list
	}
	return out, nil
}

// findChefByID resolves a chef by id WITHOUT the is_active filter that
// ChefRepository.FindByID applies — support must be able to open a
// deactivated kitchen.
func (r *AdminRepository) findChefByID(ctx context.Context, chefID int) (*domain.Chef, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+chefColumns+` FROM chefs WHERE id = $1`, chefID)
	c, err := scanChef(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrChefNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin find chef: %w", err)
	}
	return c, nil
}

// findChefByUserID is findChefByID's owner-side twin, also unfiltered.
func (r *AdminRepository) findChefByUserID(ctx context.Context, userID int) (*domain.Chef, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+chefColumns+` FROM chefs WHERE user_id = $1`, userID)
	c, err := scanChef(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrChefNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("admin find chef by user: %w", err)
	}
	return c, nil
}
