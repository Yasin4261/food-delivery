package model

import (
	"time"
)

// Order model - Sipariş
type Order struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Total     float64   `json:"total" db:"total"`
	Status    string    `json:"status" db:"status"` // pending, confirmed, shipped, delivered, cancelled
	Address   string    `json:"address" db:"address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Join işlemleri için
	UserEmail     string `json:"user_email,omitempty" db:"user_email"`
	UserFirstName string `json:"user_first_name,omitempty" db:"user_first_name"`
	UserLastName  string `json:"user_last_name,omitempty" db:"user_last_name"`
}

// OrderItem model - Sipariş öğesi
type OrderItem struct {
	ID        uint      `json:"id" db:"id"`
	OrderID   uint      `json:"order_id" db:"order_id"`
	ProductID uint      `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"` // Sipariş anındaki fiyat
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Join işlemleri için
	ProductName string `json:"product_name,omitempty" db:"product_name"`
}

// Order request/response models
type CreateOrderRequest struct {
	Address string           `json:"address" binding:"required"`
	Items   []OrderItemInput `json:"items" binding:"required,min=1"`
}

type OrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type OrderResponse struct {
	ID        uint                `json:"id"`
	UserID    uint                `json:"user_id"`
	Total     float64             `json:"total"`
	Status    string              `json:"status"`
	Address   string              `json:"address"`
	Items     []OrderItemResponse `json:"items,omitempty"`
	User      *UserInfo           `json:"user,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID          uint      `json:"id"`
	ProductID   uint      `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
