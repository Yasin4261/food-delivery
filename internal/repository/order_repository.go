package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"time"
)

// OrderRepository - sipariş veritabanı işlemleri
type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *model.Order) error {
	query := `
		INSERT INTO orders (user_id, total, status, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, order.UserID, order.Total, order.Status, 
		order.Address, now, now).Scan(&order.ID)
	
	if err != nil {
		return err
	}
	
	order.CreatedAt = now
	order.UpdatedAt = now
	return nil
}

func (r *OrderRepository) CreateOrderItem(item *model.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, item.OrderID, item.ProductID, item.Quantity, 
		item.Price, now, now).Scan(&item.ID)
	
	if err != nil {
		return err
	}
	
	item.CreatedAt = now
	item.UpdatedAt = now
	return nil
}

func (r *OrderRepository) GetByUserID(userID uint) ([]model.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.total, o.status, o.address, o.created_at, o.updated_at,
		       u.email, u.first_name, u.last_name
		FROM orders o
		JOIN users u ON o.user_id = u.id
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
			&order.ID, &order.UserID, &order.Total, &order.Status, &order.Address,
			&order.CreatedAt, &order.UpdatedAt, &order.UserEmail, 
			&order.UserFirstName, &order.UserLastName,
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
		SELECT o.id, o.user_id, o.total, o.status, o.address, o.created_at, o.updated_at,
		       u.email, u.first_name, u.last_name
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = $1`
	
	err := r.db.QueryRow(query, id).Scan(
		&order.ID, &order.UserID, &order.Total, &order.Status, &order.Address,
		&order.CreatedAt, &order.UpdatedAt, &order.UserEmail, 
		&order.UserFirstName, &order.UserLastName,
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
		SELECT o.id, o.user_id, o.total, o.status, o.address, o.created_at, o.updated_at,
		       u.email, u.first_name, u.last_name
		FROM orders o
		JOIN users u ON o.user_id = u.id
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
			&order.ID, &order.UserID, &order.Total, &order.Status, &order.Address,
			&order.CreatedAt, &order.UpdatedAt, &order.UserEmail, 
			&order.UserFirstName, &order.UserLastName,
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
		SELECT o.id, o.user_id, o.total, o.status, o.address, o.created_at, o.updated_at,
		       u.email, u.first_name, u.last_name
		FROM orders o
		JOIN users u ON o.user_id = u.id
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
			&order.ID, &order.UserID, &order.Total, &order.Status, &order.Address,
			&order.CreatedAt, &order.UpdatedAt, &order.UserEmail, 
			&order.UserFirstName, &order.UserLastName,
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
		SELECT o.id, o.user_id, o.total, o.status, o.address, o.created_at, o.updated_at,
		       u.email, u.first_name, u.last_name
		FROM orders o
		JOIN users u ON o.user_id = u.id
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
			&order.ID, &order.UserID, &order.Total, &order.Status, &order.Address,
			&order.CreatedAt, &order.UpdatedAt, &order.UserEmail, 
			&order.UserFirstName, &order.UserLastName,
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
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, 
		       oi.created_at, oi.updated_at, p.name as product_name
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
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
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price,
			&item.CreatedAt, &item.UpdatedAt, &item.ProductName,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, nil
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
