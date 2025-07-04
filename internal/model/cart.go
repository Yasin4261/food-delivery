package model

import (
	"time"
)

// Cart model - Sepet
type Cart struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CartItem model - Sepet öğesi
type CartItem struct {
	ID        uint      `json:"id" db:"id"`
	CartID    uint      `json:"cart_id" db:"cart_id"`
	ProductID uint      `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Join işlemleri için
	ProductName  string  `json:"product_name,omitempty" db:"product_name"`
	ProductPrice float64 `json:"product_price,omitempty" db:"product_price"`
	ProductImage string  `json:"product_image,omitempty" db:"product_image"`
}

// Cart request/response models
type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gte=0"`
}

type CartResponse struct {
	ID        uint               `json:"id"`
	UserID    uint               `json:"user_id"`
	Items     []CartItemResponse `json:"items"`
	Total     float64            `json:"total"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type CartItemResponse struct {
	ID           uint      `json:"id"`
	ProductID    uint      `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductPrice float64   `json:"product_price"`
	ProductImage string    `json:"product_image"`
	Quantity     int       `json:"quantity"`
	Subtotal     float64   `json:"subtotal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
