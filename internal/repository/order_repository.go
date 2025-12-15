package repository

import (
	"database/sql"
	"fmt"
	
	"github.com/Yasin4261/food-delivery/internal/domain"
)

// OrderRepository implements domain.OrderRepository interface
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order with order items in a transaction
func (r *OrderRepository) Create(order *domain.Order, items []domain.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Insert order
	orderQuery := `
		INSERT INTO orders (
			order_code, user_id, subtotal, delivery_fee, service_fee, tax, discount, total_price,
			status, payment_method, payment_status, delivery_address, delivery_city,
			delivery_latitude, delivery_longitude, estimated_delivery_time, actual_delivery_time,
			customer_notes, chef_notes, delivery_notes, created_at, updated_at, cancelled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		RETURNING id, created_at, updated_at
	`
	
	err = tx.QueryRow(
		orderQuery,
		order.OrderCode,
		order.UserID,
		order.Subtotal,
		order.DeliveryFee,
		order.ServiceFee,
		order.Tax,
		order.Discount,
		order.TotalPrice,
		order.Status,
		order.PaymentMethod,
		order.PaymentStatus,
		order.DeliveryAddress,
		order.DeliveryCity,
		order.DeliveryLatitude,
		order.DeliveryLongitude,
		order.EstimatedDeliveryTime,
		order.ActualDeliveryTime,
		order.CustomerNotes,
		order.ChefNotes,
		order.DeliveryNotes,
		order.CreatedAt,
		order.UpdatedAt,
		order.CancelledAt,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	
	// Insert order items
	itemQuery := `
		INSERT INTO order_items (
			order_id, menu_item_id, chef_id, item_name, quantity, unit_price, subtotal, special_instructions, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	
	for i := range items {
		items[i].OrderID = order.ID
		err = tx.QueryRow(
			itemQuery,
			items[i].OrderID,
			items[i].MenuItemID,
			items[i].ChefID,
			items[i].ItemName,
			items[i].Quantity,
			items[i].UnitPrice,
			items[i].Subtotal,
			items[i].SpecialInstructions,
			items[i].CreatedAt,
		).Scan(&items[i].ID)
		
		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}
	
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	order.Items = items
	return nil
}

// FindByID finds an order by ID
func (r *OrderRepository) FindByID(id int) (*domain.Order, error) {
	query := `
		SELECT id, order_code, user_id, subtotal, delivery_fee, service_fee, tax, discount, total_price,
		       status, payment_method, payment_status, delivery_address, delivery_city,
		       delivery_latitude, delivery_longitude, estimated_delivery_time, actual_delivery_time,
		       customer_notes, chef_notes, delivery_notes, created_at, updated_at, cancelled_at
		FROM orders
		WHERE id = $1
	`
	
	order := &domain.Order{}
	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.OrderCode,
		&order.UserID,
		&order.Subtotal,
		&order.DeliveryFee,
		&order.ServiceFee,
		&order.Tax,
		&order.Discount,
		&order.TotalPrice,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.DeliveryAddress,
		&order.DeliveryCity,
		&order.DeliveryLatitude,
		&order.DeliveryLongitude,
		&order.EstimatedDeliveryTime,
		&order.ActualDeliveryTime,
		&order.CustomerNotes,
		&order.ChefNotes,
		&order.DeliveryNotes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.CancelledAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}
	
	// Load order items
	items, err := r.FindOrderItems(order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	order.Items = items
	
	return order, nil
}

// FindByOrderCode finds an order by order code
func (r *OrderRepository) FindByOrderCode(orderCode string) (*domain.Order, error) {
	query := `
		SELECT id, order_code, user_id, subtotal, delivery_fee, service_fee, tax, discount, total_price,
		       status, payment_method, payment_status, delivery_address, delivery_city,
		       delivery_latitude, delivery_longitude, estimated_delivery_time, actual_delivery_time,
		       customer_notes, chef_notes, delivery_notes, created_at, updated_at, cancelled_at
		FROM orders
		WHERE order_code = $1
	`
	
	order := &domain.Order{}
	err := r.db.QueryRow(query, orderCode).Scan(
		&order.ID,
		&order.OrderCode,
		&order.UserID,
		&order.Subtotal,
		&order.DeliveryFee,
		&order.ServiceFee,
		&order.Tax,
		&order.Discount,
		&order.TotalPrice,
		&order.Status,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.DeliveryAddress,
		&order.DeliveryCity,
		&order.DeliveryLatitude,
		&order.DeliveryLongitude,
		&order.EstimatedDeliveryTime,
		&order.ActualDeliveryTime,
		&order.CustomerNotes,
		&order.ChefNotes,
		&order.DeliveryNotes,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.CancelledAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find order: %w", err)
	}
	
	// Load order items
	items, err := r.FindOrderItems(order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	order.Items = items
	
	return order, nil
}

// Update updates order information
func (r *OrderRepository) Update(order *domain.Order) error {
	query := `
		UPDATE orders
		SET subtotal = $1, delivery_fee = $2, service_fee = $3, tax = $4, discount = $5, total_price = $6,
		    status = $7, payment_method = $8, payment_status = $9, delivery_address = $10, delivery_city = $11,
		    delivery_latitude = $12, delivery_longitude = $13, estimated_delivery_time = $14, actual_delivery_time = $15,
		    customer_notes = $16, chef_notes = $17, delivery_notes = $18, updated_at = $19, cancelled_at = $20
		WHERE id = $21
	`
	
	result, err := r.db.Exec(
		query,
		order.Subtotal,
		order.DeliveryFee,
		order.ServiceFee,
		order.Tax,
		order.Discount,
		order.TotalPrice,
		order.Status,
		order.PaymentMethod,
		order.PaymentStatus,
		order.DeliveryAddress,
		order.DeliveryCity,
		order.DeliveryLatitude,
		order.DeliveryLongitude,
		order.EstimatedDeliveryTime,
		order.ActualDeliveryTime,
		order.CustomerNotes,
		order.ChefNotes,
		order.DeliveryNotes,
		order.UpdatedAt,
		order.CancelledAt,
		order.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}
	
	return nil
}

// UpdateStatus updates order status
func (r *OrderRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	
	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}
	
	return nil
}

// FindByUserID finds orders by user ID
func (r *OrderRepository) FindByUserID(userID int, offset, limit int) ([]*domain.Order, error) {
	query := `
		SELECT id, order_code, user_id, subtotal, delivery_fee, service_fee, tax, discount, total_price,
		       status, payment_method, payment_status, delivery_address, delivery_city,
		       delivery_latitude, delivery_longitude, estimated_delivery_time, actual_delivery_time,
		       customer_notes, chef_notes, delivery_notes, created_at, updated_at, cancelled_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by user: %w", err)
	}
	defer rows.Close()
	
	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.ID,
			&order.OrderCode,
			&order.UserID,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.ServiceFee,
			&order.Tax,
			&order.Discount,
			&order.TotalPrice,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.DeliveryAddress,
			&order.DeliveryCity,
			&order.DeliveryLatitude,
			&order.DeliveryLongitude,
			&order.EstimatedDeliveryTime,
			&order.ActualDeliveryTime,
			&order.CustomerNotes,
			&order.ChefNotes,
			&order.DeliveryNotes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

// FindByChefID finds orders containing items from a chef
func (r *OrderRepository) FindByChefID(chefID int, offset, limit int) ([]*domain.Order, error) {
	query := `
		SELECT DISTINCT o.id, o.order_code, o.user_id, o.subtotal, o.delivery_fee, o.service_fee, 
		       o.tax, o.discount, o.total_price, o.status, o.payment_method, o.payment_status,
		       o.delivery_address, o.delivery_city, o.delivery_latitude, o.delivery_longitude,
		       o.estimated_delivery_time, o.actual_delivery_time, o.customer_notes, o.chef_notes,
		       o.delivery_notes, o.created_at, o.updated_at, o.cancelled_at
		FROM orders o
		INNER JOIN order_items oi ON o.id = oi.order_id
		WHERE oi.chef_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, chefID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by chef: %w", err)
	}
	defer rows.Close()
	
	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.ID,
			&order.OrderCode,
			&order.UserID,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.ServiceFee,
			&order.Tax,
			&order.Discount,
			&order.TotalPrice,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.DeliveryAddress,
			&order.DeliveryCity,
			&order.DeliveryLatitude,
			&order.DeliveryLongitude,
			&order.EstimatedDeliveryTime,
			&order.ActualDeliveryTime,
			&order.CustomerNotes,
			&order.ChefNotes,
			&order.DeliveryNotes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

// FindByStatus finds orders by status
func (r *OrderRepository) FindByStatus(status string, offset, limit int) ([]*domain.Order, error) {
	query := `
		SELECT id, order_code, user_id, subtotal, delivery_fee, service_fee, tax, discount, total_price,
		       status, payment_method, payment_status, delivery_address, delivery_city,
		       delivery_latitude, delivery_longitude, estimated_delivery_time, actual_delivery_time,
		       customer_notes, chef_notes, delivery_notes, created_at, updated_at, cancelled_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by status: %w", err)
	}
	defer rows.Close()
	
	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.ID,
			&order.OrderCode,
			&order.UserID,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.ServiceFee,
			&order.Tax,
			&order.Discount,
			&order.TotalPrice,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.DeliveryAddress,
			&order.DeliveryCity,
			&order.DeliveryLatitude,
			&order.DeliveryLongitude,
			&order.EstimatedDeliveryTime,
			&order.ActualDeliveryTime,
			&order.CustomerNotes,
			&order.ChefNotes,
			&order.DeliveryNotes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

// FindActiveOrders finds all non-completed/non-cancelled orders
func (r *OrderRepository) FindActiveOrders(offset, limit int) ([]*domain.Order, error) {
	query := `
		SELECT id, order_code, user_id, subtotal, delivery_fee, service_fee, tax, discount, total_price,
		       status, payment_method, payment_status, delivery_address, delivery_city,
		       delivery_latitude, delivery_longitude, estimated_delivery_time, actual_delivery_time,
		       customer_notes, chef_notes, delivery_notes, created_at, updated_at, cancelled_at
		FROM orders
		WHERE status NOT IN ('delivered', 'cancelled')
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find active orders: %w", err)
	}
	defer rows.Close()
	
	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.ID,
			&order.OrderCode,
			&order.UserID,
			&order.Subtotal,
			&order.DeliveryFee,
			&order.ServiceFee,
			&order.Tax,
			&order.Discount,
			&order.TotalPrice,
			&order.Status,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.DeliveryAddress,
			&order.DeliveryCity,
			&order.DeliveryLatitude,
			&order.DeliveryLongitude,
			&order.EstimatedDeliveryTime,
			&order.ActualDeliveryTime,
			&order.CustomerNotes,
			&order.ChefNotes,
			&order.DeliveryNotes,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.CancelledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	
	return orders, rows.Err()
}

// FindOrderItems finds items for an order
func (r *OrderRepository) FindOrderItems(orderID int) ([]domain.OrderItem, error) {
	query := `
		SELECT id, order_id, menu_item_id, chef_id, item_name, quantity, unit_price, subtotal, special_instructions, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`
	
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find order items: %w", err)
	}
	defer rows.Close()
	
	var items []domain.OrderItem
	for rows.Next() {
		item := domain.OrderItem{}
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.ChefID,
			&item.ItemName,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
			&item.SpecialInstructions,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

// CountByStatus counts orders by status
func (r *OrderRepository) CountByStatus(status string) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE status = $1`
	
	var count int
	err := r.db.QueryRow(query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders by status: %w", err)
	}
	
	return count, nil
}

// GetTotalRevenue calculates total revenue for a chef
func (r *OrderRepository) GetTotalRevenue(chefID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(oi.subtotal), 0)
		FROM order_items oi
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE oi.chef_id = $1 AND o.status = 'delivered'
	`
	
	var revenue float64
	err := r.db.QueryRow(query, chefID).Scan(&revenue)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}
	
	return revenue, nil
}
