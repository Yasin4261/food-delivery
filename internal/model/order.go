package model

import (
	"time"
)

// Order model - Sipariş (Ev yemekleri için)
type Order struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	UserID   uint `json:"user_id" gorm:"not null;index"`
	User     User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	
	// Chef Bilgisi - Ana Chef (birden fazla chef olabilir ama ana chef gerekli)
	ChefID uint `json:"chef_id" gorm:"not null;index"`
	Chef   Chef `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
	
	// Sipariş Bilgileri
	OrderNumber string    `json:"order_number" gorm:"uniqueIndex;not null"` // ORD-20240101-001
	Total       float64   `json:"total" gorm:"not null"`
	Currency    string    `json:"currency" gorm:"size:3;default:'TRY'"`
	Status      string    `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, confirmed, preparing, ready, delivered, cancelled
	
	// Teslimat Bilgileri
	DeliveryType    string    `json:"delivery_type" gorm:"size:20;not null"` // pickup, delivery
	Address         string    `json:"address" gorm:"type:text;not null"` // Ana adres field'ı
	DeliveryAddress string    `json:"delivery_address" gorm:"type:text"`  // Ayrıntılı teslimat adresi
	DeliveryDate    *time.Time `json:"delivery_date"`
	DeliveryTime    *time.Time `json:"delivery_time"`
	
	// Ödeme Bilgileri
	PaymentMethod string `json:"payment_method" gorm:"size:20"` // cash, card, online
	PaymentStatus string `json:"payment_status" gorm:"size:20;default:'pending'"` // pending, paid, failed
	
	// Notlar
	CustomerNote string `json:"customer_note" gorm:"type:text"`
	ChefNote     string `json:"chef_note" gorm:"type:text"`
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Items   []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	Reviews []Review    `json:"reviews,omitempty" gorm:"foreignKey:OrderID"`
}

// OrderItem model - Sipariş öğesi (Yemek bazlı)
type OrderItem struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	OrderID uint `json:"order_id" gorm:"not null;index"`
	MealID  uint `json:"meal_id" gorm:"not null;index"`
	ChefID  uint `json:"chef_id" gorm:"not null;index"`
	
	// Sipariş Detayları
	Quantity int     `json:"quantity" gorm:"not null;default:1"`
	Price    float64 `json:"price" gorm:"not null"` // Sipariş anındaki fiyat
	Subtotal float64 `json:"subtotal" gorm:"not null"`
	
	// Özel Talimatlar
	SpecialInstructions string `json:"special_instructions" gorm:"type:text"`
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Meal *Meal `json:"meal,omitempty" gorm:"foreignKey:MealID"`
	Chef *Chef `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
}

// Order request/response models
type CreateOrderRequest struct {
	DeliveryType    string           `json:"delivery_type" binding:"required,oneof=pickup delivery"`
	DeliveryAddress string           `json:"delivery_address"`
	DeliveryDate    time.Time        `json:"delivery_date" binding:"required"`
	DeliveryTime    string           `json:"delivery_time" binding:"required"`
	PaymentMethod   string           `json:"payment_method" binding:"required,oneof=cash card online"`
	CustomerNote    string           `json:"customer_note"`
	Items           []OrderItemInput `json:"items" binding:"required,min=1"`
}

type OrderItemInput struct {
	MealID              uint   `json:"meal_id" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,gt=0"`
	SpecialInstructions string `json:"special_instructions"`
}

type UpdateOrderStatusRequest struct {
	Status   string `json:"status" binding:"required,oneof=pending confirmed preparing ready delivered cancelled"`
	ChefNote string `json:"chef_note"`
}

type OrderResponse struct {
	ID              uint                `json:"id"`
	OrderNumber     string              `json:"order_number"`
	UserID          uint                `json:"user_id"`
	Total           float64             `json:"total"`
	Currency        string              `json:"currency"`
	Status          string              `json:"status"`
	DeliveryType    string              `json:"delivery_type"`
	DeliveryAddress string              `json:"delivery_address"`
	DeliveryDate    time.Time           `json:"delivery_date"`
	DeliveryTime    string              `json:"delivery_time"`
	PaymentMethod   string              `json:"payment_method"`
	PaymentStatus   string              `json:"payment_status"`
	CustomerNote    string              `json:"customer_note"`
	ChefNote        string              `json:"chef_note"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	User            *UserInfo           `json:"user,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID                  uint      `json:"id"`
	MealID              uint      `json:"meal_id"`
	ChefID              uint      `json:"chef_id"`
	MealName            string    `json:"meal_name"`
	ChefName            string    `json:"chef_name"`
	KitchenName         string    `json:"kitchen_name"`
	Quantity            int       `json:"quantity"`
	Price               float64   `json:"price"`
	Subtotal            float64   `json:"subtotal"`
	SpecialInstructions string    `json:"special_instructions"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
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
	Phone     string `json:"phone"`
}
