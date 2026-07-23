package repository

import (
	"context"
	"database/sql"
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
