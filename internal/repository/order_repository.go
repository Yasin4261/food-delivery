package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"time"
)

// OrderRepository - sipariş veritabanı işlemleri (Ev yemekleri platformu için)
type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *model.Order) error {
	query := `
		INSERT INTO orders (user_id, order_number, total, currency, status, delivery_type, address, 
			delivery_address, delivery_date, delivery_time, delivery_latitude, delivery_longitude, 
			delivery_radius, payment_method, payment_status, customer_note, chef_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, 
		order.UserID, order.OrderNumber, order.Total, order.Currency, order.Status,
		order.DeliveryType, order.Address, order.DeliveryAddress, order.DeliveryDate, order.DeliveryTime,
		order.DeliveryLatitude, order.DeliveryLongitude, order.DeliveryRadius,
		order.PaymentMethod, order.PaymentStatus, order.CustomerNote, order.ChefCount,
		now, now).Scan(&order.ID)
	
	if err != nil {
		return err
	}
	
	order.CreatedAt = now
	order.UpdatedAt = now
	return nil
}

func (r *OrderRepository) CreateOrderItem(item *model.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, sub_order_id, meal_id, chef_id, quantity, price, subtotal, special_instructions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, 
		item.OrderID, item.SubOrderID, item.MealID, item.ChefID, item.Quantity, 
		item.Price, item.Subtotal, item.SpecialInstructions, now, now).Scan(&item.ID)
	
	if err != nil {
		return err
	}
	
	item.CreatedAt = now
	item.UpdatedAt = now
	return nil
}

func (r *OrderRepository) GetByUserID(userID uint) ([]model.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.order_number, o.total, o.currency, o.status, o.delivery_type,
		       o.address, o.delivery_address, o.delivery_date, o.delivery_time, 
		       o.delivery_latitude, o.delivery_longitude, o.delivery_radius,
		       o.payment_method, o.payment_status, o.customer_note, o.chef_count,
		       o.created_at, o.updated_at
		FROM orders o
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.OrderNumber, &order.Total, &order.Currency, &order.Status,
			&order.DeliveryType, &order.Address, &order.DeliveryAddress, &order.DeliveryDate, &order.DeliveryTime,
			&order.DeliveryLatitude, &order.DeliveryLongitude, &order.DeliveryRadius,
			&order.PaymentMethod, &order.PaymentStatus, &order.CustomerNote, &order.ChefCount,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	return orders, nil
}

func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	order := &model.Order{}
	query := `
		SELECT o.id, o.user_id, o.order_number, o.total, o.currency, o.status, o.delivery_type,
		       o.address, o.delivery_address, o.delivery_date, o.delivery_time, 
		       o.delivery_latitude, o.delivery_longitude, o.delivery_radius,
		       o.payment_method, o.payment_status, o.customer_note, o.chef_count,
		       o.created_at, o.updated_at
		FROM orders o
		WHERE o.id = $1`
	
	err := r.db.QueryRow(query, id).Scan(
		&order.ID, &order.UserID, &order.OrderNumber, &order.Total, &order.Currency, &order.Status,
		&order.DeliveryType, &order.Address, &order.DeliveryAddress, &order.DeliveryDate, &order.DeliveryTime,
		&order.DeliveryLatitude, &order.DeliveryLongitude, &order.DeliveryRadius,
		&order.PaymentMethod, &order.PaymentStatus, &order.CustomerNote, &order.ChefCount,
		&order.CreatedAt, &order.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Sipariş bulunamadı
		}
		return nil, err
	}
	
	return order, nil
}

func (r *OrderRepository) GetAll() ([]model.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.order_number, o.total, o.currency, o.status, o.delivery_type,
		       o.address, o.delivery_address, o.delivery_date, o.delivery_time, 
		       o.delivery_latitude, o.delivery_longitude, o.delivery_radius,
		       o.payment_method, o.payment_status, o.customer_note, o.chef_count,
		       o.created_at, o.updated_at
		FROM orders o
		ORDER BY o.created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.OrderNumber, &order.Total, &order.Currency, &order.Status,
			&order.DeliveryType, &order.Address, &order.DeliveryAddress, &order.DeliveryDate, &order.DeliveryTime,
			&order.DeliveryLatitude, &order.DeliveryLongitude, &order.DeliveryRadius,
			&order.PaymentMethod, &order.PaymentStatus, &order.CustomerNote, &order.ChefCount,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	return orders, nil
}

func (r *OrderRepository) GetByStatus(status string) ([]model.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.order_number, o.total, o.currency, o.status, o.delivery_type,
		       o.address, o.delivery_address, o.delivery_date, o.delivery_time, 
		       o.delivery_latitude, o.delivery_longitude, o.delivery_radius,
		       o.payment_method, o.payment_status, o.customer_note, o.chef_count,
		       o.created_at, o.updated_at
		FROM orders o
		WHERE o.status = $1
		ORDER BY o.created_at DESC`
	
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.OrderNumber, &order.Total, &order.Currency, &order.Status,
			&order.DeliveryType, &order.Address, &order.DeliveryAddress, &order.DeliveryDate, &order.DeliveryTime,
			&order.DeliveryLatitude, &order.DeliveryLongitude, &order.DeliveryRadius,
			&order.PaymentMethod, &order.PaymentStatus, &order.CustomerNote, &order.ChefCount,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	
	return orders, nil
}

func (r *OrderRepository) GetRecent(limit int) ([]model.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.order_number, o.total, o.currency, o.status, o.delivery_type,
		       o.address, o.delivery_address, o.delivery_date, o.delivery_time, 
		       o.delivery_latitude, o.delivery_longitude, o.delivery_radius,
		       o.payment_method, o.payment_status, o.customer_note, o.chef_count,
		       o.created_at, o.updated_at
		FROM orders o
		ORDER BY o.created_at DESC
		LIMIT $1`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.OrderNumber, &order.Total, &order.Currency, &order.Status,
			&order.DeliveryType, &order.Address, &order.DeliveryAddress, &order.DeliveryDate, &order.DeliveryTime,
			&order.DeliveryLatitude, &order.DeliveryLongitude, &order.DeliveryRadius,
			&order.PaymentMethod, &order.PaymentStatus, &order.CustomerNote, &order.ChefCount,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(id uint, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	now := time.Now()
	_, err := r.db.Exec(query, status, now, id)
	return err
}

func (r *OrderRepository) GetOrderItems(orderID uint) ([]model.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.sub_order_id, oi.meal_id, oi.chef_id, oi.quantity, oi.price, 
		       oi.subtotal, oi.special_instructions, oi.created_at, oi.updated_at
		FROM order_items oi
		WHERE oi.order_id = $1
		ORDER BY oi.created_at ASC`
	
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.SubOrderID, &item.MealID, &item.ChefID, &item.Quantity, &item.Price,
			&item.Subtotal, &item.SpecialInstructions, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, nil
}

// SubOrder repository methods
func (r *OrderRepository) CreateSubOrder(subOrder *model.SubOrder) error {
	query := `
		INSERT INTO sub_orders (order_id, chef_id, chef_order_number, subtotal, delivery_fee, service_fee, total, 
			status, estimated_time, chef_note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, 
		subOrder.OrderID, subOrder.ChefID, subOrder.ChefOrderNumber, subOrder.Subtotal, 
		subOrder.DeliveryFee, subOrder.ServiceFee, subOrder.Total, subOrder.Status,
		subOrder.EstimatedTime, subOrder.ChefNote, now, now).Scan(&subOrder.ID)
	
	if err != nil {
		return err
	}
	
	subOrder.CreatedAt = now
	subOrder.UpdatedAt = now
	return nil
}

func (r *OrderRepository) GetSubOrdersByOrderID(orderID uint) ([]model.SubOrder, error) {
	query := `
		SELECT so.id, so.order_id, so.chef_id, so.chef_order_number, so.subtotal, so.delivery_fee, 
		       so.service_fee, so.total, so.status, so.estimated_time, so.chef_note, 
		       so.created_at, so.updated_at
		FROM sub_orders so
		WHERE so.order_id = $1
		ORDER BY so.created_at ASC`
	
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subOrders []model.SubOrder
	for rows.Next() {
		var subOrder model.SubOrder
		err := rows.Scan(
			&subOrder.ID, &subOrder.OrderID, &subOrder.ChefID, &subOrder.ChefOrderNumber, 
			&subOrder.Subtotal, &subOrder.DeliveryFee, &subOrder.ServiceFee, &subOrder.Total,
			&subOrder.Status, &subOrder.EstimatedTime, &subOrder.ChefNote,
			&subOrder.CreatedAt, &subOrder.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subOrders = append(subOrders, subOrder)
	}
	
	return subOrders, nil
}

func (r *OrderRepository) GetSubOrdersByChefID(chefID uint) ([]model.SubOrder, error) {
	query := `
		SELECT so.id, so.order_id, so.chef_id, so.chef_order_number, so.subtotal, so.delivery_fee, 
		       so.service_fee, so.total, so.status, so.estimated_time, so.chef_note, 
		       so.created_at, so.updated_at
		FROM sub_orders so
		WHERE so.chef_id = $1
		ORDER BY so.created_at DESC`
	
	rows, err := r.db.Query(query, chefID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subOrders []model.SubOrder
	for rows.Next() {
		var subOrder model.SubOrder
		err := rows.Scan(
			&subOrder.ID, &subOrder.OrderID, &subOrder.ChefID, &subOrder.ChefOrderNumber, 
			&subOrder.Subtotal, &subOrder.DeliveryFee, &subOrder.ServiceFee, &subOrder.Total,
			&subOrder.Status, &subOrder.EstimatedTime, &subOrder.ChefNote,
			&subOrder.CreatedAt, &subOrder.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subOrders = append(subOrders, subOrder)
	}
	
	return subOrders, nil
}

func (r *OrderRepository) UpdateSubOrderStatus(subOrderID uint, status string, chefNote string) error {
	query := `UPDATE sub_orders SET status = $1, chef_note = $2, updated_at = $3 WHERE id = $4`
	now := time.Now()
	_, err := r.db.Exec(query, status, chefNote, now, subOrderID)
	return err
}

func (r *OrderRepository) UpdateSubOrder(subOrder *model.SubOrder) error {
	query := `UPDATE sub_orders SET 
		subtotal = $1, 
		delivery_fee = $2, 
		service_fee = $3, 
		total = $4, 
		status = $5, 
		estimated_time = $6, 
		chef_note = $7, 
		updated_at = $8 
		WHERE id = $9`
	now := time.Now()
	_, err := r.db.Exec(query, 
		subOrder.Subtotal, 
		subOrder.DeliveryFee, 
		subOrder.ServiceFee, 
		subOrder.Total, 
		subOrder.Status, 
		subOrder.EstimatedTime, 
		subOrder.ChefNote, 
		now, 
		subOrder.ID)
	return err
}

func (r *OrderRepository) Delete(id uint) error {
	// Önce order items'ları sil
	_, err := r.db.Exec(`DELETE FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		return err
	}
	
	// Sonra order'ı sil
	_, err = r.db.Exec(`DELETE FROM orders WHERE id = $1`, id)
	return err
}
