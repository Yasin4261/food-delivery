package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// OrderRepository is the PostgreSQL adapter for domain.OrderRepository. Order
// creation spans two tables, so it runs in a transaction.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository builds an OrderRepository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

const orderColumns = `
	id, order_code, user_id,
	subtotal, delivery_fee, service_fee, tax, discount, total_price,
	status, payment_method, payment_status,
	delivery_address, delivery_city, delivery_latitude, delivery_longitude,
	estimated_delivery_time, actual_delivery_time,
	customer_notes, chef_notes, delivery_notes,
	created_at, updated_at, cancelled_at`

func scanOrder(s interface{ Scan(...any) error }) (*domain.Order, error) {
	o := &domain.Order{}
	err := s.Scan(
		&o.ID, &o.OrderCode, &o.UserID,
		&o.Subtotal, &o.DeliveryFee, &o.ServiceFee, &o.Tax, &o.Discount, &o.TotalPrice,
		&o.Status, &o.PaymentMethod, &o.PaymentStatus,
		&o.DeliveryAddress, &o.DeliveryCity, &o.DeliveryLatitude, &o.DeliveryLongitude,
		&o.EstimatedDeliveryTime, &o.ActualDeliveryTime,
		&o.CustomerNotes, &o.ChefNotes, &o.DeliveryNotes,
		&o.CreatedAt, &o.UpdatedAt, &o.CancelledAt,
	)
	return o, err
}

const orderItemColumns = `
	id, order_id, menu_item_id, chef_id, item_name, quantity, unit_price, subtotal,
	special_instructions, created_at`

func scanOrderItem(s interface{ Scan(...any) error }) (*domain.OrderItem, error) {
	it := &domain.OrderItem{}
	err := s.Scan(
		&it.ID, &it.OrderID, &it.MenuItemID, &it.ChefID, &it.ItemName, &it.Quantity,
		&it.UnitPrice, &it.Subtotal, &it.SpecialInstructions, &it.CreatedAt,
	)
	return it, err
}

// Create persists an order and its items in one transaction.
func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	orderQuery := `
		INSERT INTO orders (
			order_code, user_id,
			subtotal, delivery_fee, service_fee, tax, discount, total_price,
			status, payment_method, payment_status,
			delivery_address, delivery_city, delivery_latitude, delivery_longitude,
			customer_notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(
		ctx, orderQuery,
		o.OrderCode, o.UserID,
		o.Subtotal, o.DeliveryFee, o.ServiceFee, o.Tax, o.Discount, o.TotalPrice,
		o.Status, o.PaymentMethod, o.PaymentStatus,
		o.DeliveryAddress, o.DeliveryCity, o.DeliveryLatitude, o.DeliveryLongitude,
		o.CustomerNotes,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (
			order_id, menu_item_id, chef_id, item_name, quantity, unit_price, subtotal, special_instructions
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	for _, it := range o.Items {
		it.OrderID = o.ID
		err = tx.QueryRowContext(
			ctx, itemQuery,
			it.OrderID, it.MenuItemID, it.ChefID, it.ItemName, it.Quantity, it.UnitPrice, it.Subtotal, it.SpecialInstructions,
		).Scan(&it.ID, &it.CreatedAt)
		if err != nil {
			return fmt.Errorf("create order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit order: %w", err)
	}
	return nil
}

// FindByID returns an order with all of its items.
func (r *OrderRepository) FindByID(ctx context.Context, id int) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+orderColumns+` FROM orders WHERE id = $1`, id)
	o, err := scanOrder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find order: %w", err)
	}

	items, err := r.loadItems(ctx, o.ID, 0)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

// ListByUser returns a page of a customer's orders, newest first, each with all
// items, plus the total order count.
func (r *OrderRepository) ListByUser(ctx context.Context, userID, limit, offset int) ([]*domain.Order, int, error) {
	query := `SELECT` + orderColumns + `
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by user: %w", err)
	}
	orders, err := collectOrders(rows)
	if err != nil {
		return nil, 0, err
	}
	for _, o := range orders {
		if o.Items, err = r.loadItems(ctx, o.ID, 0); err != nil {
			return nil, 0, err
		}
	}
	total, err := r.countOrders(ctx, `SELECT count(*) FROM orders WHERE user_id = $1`, userID)
	return orders, total, err
}

// ListByChef returns a page of orders containing the chef's items, newest
// first (items filtered to that chef), plus the total count.
func (r *OrderRepository) ListByChef(ctx context.Context, chefID, limit, offset int) ([]*domain.Order, int, error) {
	query := `SELECT` + orderColumns + `
		FROM orders o
		WHERE EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = o.id AND oi.chef_id = $1)
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, chefID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders by chef: %w", err)
	}
	orders, err := collectOrders(rows)
	if err != nil {
		return nil, 0, err
	}
	for _, o := range orders {
		if o.Items, err = r.loadItems(ctx, o.ID, chefID); err != nil {
			return nil, 0, err
		}
	}
	total, err := r.countOrders(ctx,
		`SELECT count(*) FROM orders o WHERE EXISTS (SELECT 1 FROM order_items oi WHERE oi.order_id = o.id AND oi.chef_id = $1)`, chefID)
	return orders, total, err
}

func (r *OrderRepository) countOrders(ctx context.Context, query string, arg any) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, query, arg).Scan(&total); err != nil {
		return 0, fmt.Errorf("count orders: %w", err)
	}
	return total, nil
}

// UpdateStatus persists the mutable fields touched by a transition.
func (r *OrderRepository) UpdateStatus(ctx context.Context, o *domain.Order) error {
	query := `
		UPDATE orders
		SET status = $2, payment_status = $3, actual_delivery_time = $4, cancelled_at = $5, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		o.ID, o.Status, o.PaymentStatus, o.ActualDeliveryTime, o.CancelledAt,
	).Scan(&o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrOrderNotFound
	}
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

// loadItems returns an order's items. When chefID > 0 the result is filtered to
// that chef (chef-scoped views).
func (r *OrderRepository) loadItems(ctx context.Context, orderID, chefID int) ([]*domain.OrderItem, error) {
	query := `SELECT` + orderItemColumns + ` FROM order_items WHERE order_id = $1`
	args := []any{orderID}
	if chefID > 0 {
		query += ` AND chef_id = $2`
		args = append(args, chefID)
	}
	query += ` ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load order items: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.OrderItem, 0)
	for rows.Next() {
		it, err := scanOrderItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func collectOrders(rows *sql.Rows) ([]*domain.Order, error) {
	defer rows.Close()
	orders := make([]*domain.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
