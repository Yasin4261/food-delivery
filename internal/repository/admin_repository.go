package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

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

// SetUserActive toggles a user's active flag, atomic with its audit entry.
func (r *AdminRepository) SetUserActive(ctx context.Context, e *domain.AuditEntry, userID int, active bool) error {
	return r.auditBoolToggle(ctx, e, "is_active",
		`SELECT is_active FROM users WHERE id = $1 FOR UPDATE`,
		`UPDATE users SET is_active = $2, updated_at = now() WHERE id = $1`,
		userID, active, domain.ErrUserNotFound)
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

// SetChefActive toggles a chef's active flag, atomic with its audit entry.
func (r *AdminRepository) SetChefActive(ctx context.Context, e *domain.AuditEntry, chefID int, active bool) error {
	return r.auditBoolToggle(ctx, e, "is_active",
		`SELECT is_active FROM chefs WHERE id = $1 FOR UPDATE`,
		`UPDATE chefs SET is_active = $2, updated_at = now() WHERE id = $1`,
		chefID, active, domain.ErrChefNotFound)
}

// SetChefOnline drives a chef's live presence on their behalf (admin support).
func (r *AdminRepository) SetChefOnline(ctx context.Context, e *domain.AuditEntry, chefID int, online bool) error {
	return r.auditBoolToggle(ctx, e, "is_online",
		`SELECT is_online FROM chefs WHERE id = $1 FOR UPDATE`,
		`UPDATE chefs SET is_online = $2, updated_at = now() WHERE id = $1`,
		chefID, online, domain.ErrChefNotFound)
}

// SetChefAcceptingOrders drives a chef's availability on their behalf.
func (r *AdminRepository) SetChefAcceptingOrders(ctx context.Context, e *domain.AuditEntry, chefID int, accepting bool) error {
	return r.auditBoolToggle(ctx, e, "is_accepting_orders",
		`SELECT is_accepting_orders FROM chefs WHERE id = $1 FOR UPDATE`,
		`UPDATE chefs SET is_accepting_orders = $2, updated_at = now() WHERE id = $1`,
		chefID, accepting, domain.ErrChefNotFound)
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

// --- audit infrastructure (#121) -------------------------------------------

// auditJSON marshals a small state map for the before/after columns. The maps
// are built from known non-secret fields, so marshalling never fails in
// practice; a failure yields JSON null rather than a panic.
func auditJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return b
}

// insertAudit writes one audit row into the caller's transaction, so the entry
// and the mutation it describes commit or roll back together.
func insertAudit(ctx context.Context, tx *sql.Tx, e *domain.AuditEntry) error {
	err := tx.QueryRowContext(ctx, `
		INSERT INTO admin_audit_log
			(actor_user_id, action, target_type, target_id, reason, before_json, after_json)
		VALUES ($1, $2, $3, $4, NULLIF($5,''), $6, $7)
		RETURNING id, created_at`,
		e.ActorUserID, e.Action, e.TargetType, e.TargetID, e.Reason, nullableJSON(e.Before), nullableJSON(e.After)).
		Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

// nullableJSON turns an empty RawMessage into a SQL NULL.
func nullableJSON(m json.RawMessage) any {
	if len(m) == 0 {
		return nil
	}
	return []byte(m)
}

// auditBoolToggle updates a single boolean column and records an audit row in
// one transaction, capturing before/after. selectSQL/updateSQL/field are
// internal constants (never user input). A missing row rolls back and returns
// notFound.
func (r *AdminRepository) auditBoolToggle(ctx context.Context, e *domain.AuditEntry, field, selectSQL, updateSQL string, id int, newVal bool, notFound error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var before bool
	err = tx.QueryRowContext(ctx, selectSQL, id).Scan(&before)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	if err != nil {
		return fmt.Errorf("read before: %w", err)
	}
	if _, err := tx.ExecContext(ctx, updateSQL, id, newVal); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	e.Before = auditJSON(map[string]bool{field: before})
	e.After = auditJSON(map[string]bool{field: newVal})
	if err := insertAudit(ctx, tx, e); err != nil {
		return err
	}
	return tx.Commit()
}

// --- promo management (#122), each atomic with an audit entry ---------------

// promoSnapshot is the non-secret promo state captured in before/after.
func promoSnapshot(p *domain.PromoCode) map[string]any {
	return map[string]any{
		"code": p.Code, "discount_type": p.DiscountType, "discount_value": p.DiscountValue,
		"min_order": p.MinOrder, "usage_limit": p.UsageLimit, "is_active": p.IsActive,
	}
}

// ListPromos returns a page of all promo codes (newest first) for the admin
// surface, so promo reads and writes share one adapter (and one test store).
func (r *AdminRepository) ListPromos(ctx context.Context, limit, offset int) ([]*domain.PromoCode, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+promoColumns+` FROM promo_codes ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list promos: %w", err)
	}
	defer rows.Close()
	out := make([]*domain.PromoCode, 0)
	for rows.Next() {
		p, err := scanPromo(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan promo: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM promo_codes`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count promos: %w", err)
	}
	return out, total, nil
}

// FindPromo returns one code by id.
func (r *AdminRepository) FindPromo(ctx context.Context, id int) (*domain.PromoCode, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+promoColumns+` FROM promo_codes WHERE id = $1`, id)
	p, err := scanPromo(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPromoNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find promo: %w", err)
	}
	return p, nil
}

// CreatePromo inserts a code and its audit row in one transaction.
func (r *AdminRepository) CreatePromo(ctx context.Context, e *domain.AuditEntry, p *domain.PromoCode) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO promo_codes (code, discount_type, discount_value, min_order, valid_from, valid_until, usage_limit, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, used_count, created_at`,
		p.Code, p.DiscountType, p.DiscountValue, p.MinOrder, p.ValidFrom, p.ValidUntil, p.UsageLimit, p.IsActive).
		Scan(&p.ID, &p.UsedCount, &p.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrPromoExists
		}
		return fmt.Errorf("create promo: %w", err)
	}
	e.TargetID = p.ID
	e.After = auditJSON(promoSnapshot(p))
	if err := insertAudit(ctx, tx, e); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdatePromo edits a code's definition (not its usage counter) + audit.
func (r *AdminRepository) UpdatePromo(ctx context.Context, e *domain.AuditEntry, p *domain.PromoCode) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	before, err := scanPromo(tx.QueryRowContext(ctx, `SELECT `+promoColumns+` FROM promo_codes WHERE id = $1 FOR UPDATE`, p.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrPromoNotFound
	}
	if err != nil {
		return fmt.Errorf("read promo: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE promo_codes
		SET code = $2, discount_type = $3, discount_value = $4, min_order = $5,
		    valid_from = $6, valid_until = $7, usage_limit = $8, is_active = $9
		WHERE id = $1`,
		p.ID, p.Code, p.DiscountType, p.DiscountValue, p.MinOrder, p.ValidFrom, p.ValidUntil, p.UsageLimit, p.IsActive)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrPromoExists
		}
		return fmt.Errorf("update promo: %w", err)
	}
	e.Before = auditJSON(promoSnapshot(before))
	e.After = auditJSON(promoSnapshot(p))
	if err := insertAudit(ctx, tx, e); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePromo removes a code + audit.
func (r *AdminRepository) DeletePromo(ctx context.Context, e *domain.AuditEntry, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	before, err := scanPromo(tx.QueryRowContext(ctx, `SELECT `+promoColumns+` FROM promo_codes WHERE id = $1 FOR UPDATE`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrPromoNotFound
	}
	if err != nil {
		return fmt.Errorf("read promo: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM promo_codes WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete promo: %w", err)
	}
	e.Before = auditJSON(promoSnapshot(before))
	if err := insertAudit(ctx, tx, e); err != nil {
		return err
	}
	return tx.Commit()
}

// SetPromoActive toggles a code's active flag + audit.
func (r *AdminRepository) SetPromoActive(ctx context.Context, e *domain.AuditEntry, id int, active bool) error {
	return r.auditBoolToggle(ctx, e, "is_active",
		`SELECT is_active FROM promo_codes WHERE id = $1 FOR UPDATE`,
		`UPDATE promo_codes SET is_active = $2 WHERE id = $1`,
		id, active, domain.ErrPromoNotFound)
}

// --- audit log (read-only) --------------------------------------------------

const auditColumns = `id, actor_user_id, action, target_type, target_id, reason, before_json, after_json, created_at`

// auditFilterWhere narrows the audit listing with bound placeholders only.
const auditFilterWhere = `
	WHERE ($1 = '' OR action = $1)
	  AND ($2 = '' OR target_type = $2)
	  AND ($3 = 0 OR target_id = $3)
	  AND ($4 = 0 OR actor_user_id = $4)`

// ListAudit returns a page of the audit log matching f, newest first.
func (r *AdminRepository) ListAudit(ctx context.Context, f domain.AuditFilters, limit, offset int) ([]*domain.AuditEntry, int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+auditColumns+` FROM admin_audit_log`+auditFilterWhere+`
		ORDER BY id DESC LIMIT $5 OFFSET $6`,
		f.Action, f.TargetType, f.TargetID, f.ActorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.AuditEntry, 0)
	for rows.Next() {
		e := &domain.AuditEntry{}
		var reason sql.NullString
		var before, after []byte
		if err := rows.Scan(&e.ID, &e.ActorUserID, &e.Action, &e.TargetType, &e.TargetID,
			&reason, &before, &after, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan audit: %w", err)
		}
		e.Reason = reason.String
		e.Before = json.RawMessage(before)
		e.After = json.RawMessage(after)
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM admin_audit_log`+auditFilterWhere,
		f.Action, f.TargetType, f.TargetID, f.ActorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit: %w", err)
	}
	return out, total, nil
}
