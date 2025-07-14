package model

import (
	"time"
)

// Order model - Ana Sipariş (Multi-vendor destekli)
type Order struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	UserID   uint `json:"user_id" gorm:"not null;index"`
	User     User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	
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
	
	// Konum Bilgileri (Şef filtrelemesi için)
	DeliveryLatitude  *float64 `json:"delivery_latitude" gorm:"index"`
	DeliveryLongitude *float64 `json:"delivery_longitude" gorm:"index"`
	DeliveryRadius    float64  `json:"delivery_radius" gorm:"default:10"` // km cinsinden
	
	// Ödeme Bilgileri
	PaymentMethod string `json:"payment_method" gorm:"size:20"` // cash, card, online
	PaymentStatus string `json:"payment_status" gorm:"size:20;default:'pending'"` // pending, paid, failed
	
	// Notlar
	CustomerNote string `json:"customer_note" gorm:"type:text"`
	
	// Multi-vendor support
	ChefCount int `json:"chef_count" gorm:"default:0"` // Kaç farklı şeften sipariş verildi
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	SubOrders []SubOrder  `json:"sub_orders,omitempty" gorm:"foreignKey:OrderID"` // Her şef için ayrı sub-order
	Items     []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	Reviews   []Review    `json:"reviews,omitempty" gorm:"foreignKey:OrderID"`
}

// SubOrder model - Her şef için ayrı sipariş
type SubOrder struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	OrderID uint `json:"order_id" gorm:"not null;index"`
	ChefID  uint `json:"chef_id" gorm:"not null;index"`
	
	// Şef bazlı sipariş bilgileri
	ChefOrderNumber string  `json:"chef_order_number" gorm:"uniqueIndex;not null"` // ORD-20240101-001-CHEF1
	Subtotal        float64 `json:"subtotal" gorm:"not null"`
	DeliveryFee     float64 `json:"delivery_fee" gorm:"default:0"`
	ServiceFee      float64 `json:"service_fee" gorm:"default:0"`
	Total           float64 `json:"total" gorm:"not null"`
	
	// Şef bazlı durum takibi
	Status         string `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, confirmed, preparing, ready, delivered, cancelled
	EstimatedTime  int    `json:"estimated_time" gorm:"default:30"` // dakika cinsinden
	ChefNote       string `json:"chef_note" gorm:"type:text"`
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Order *Order      `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Chef  *Chef       `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
	Items []OrderItem `json:"items,omitempty" gorm:"foreignKey:SubOrderID"`
}

// OrderItem model - Sipariş öğesi (Yemek bazlı)
type OrderItem struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	OrderID    uint `json:"order_id" gorm:"not null;index"`
	SubOrderID uint `json:"sub_order_id" gorm:"not null;index"` // Hangi şefin siparişi
	MealID     uint `json:"meal_id" gorm:"not null;index"`
	ChefID     uint `json:"chef_id" gorm:"not null;index"`
	
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
	Order    *Order    `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	SubOrder *SubOrder `json:"sub_order,omitempty" gorm:"foreignKey:SubOrderID"`
	Meal     *Meal     `json:"meal,omitempty" gorm:"foreignKey:MealID"`
	Chef     *Chef     `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
}

// Order request/response models
type CreateOrderRequest struct {
	DeliveryType      string             `json:"delivery_type" binding:"required,oneof=pickup delivery"`
	DeliveryAddress   string             `json:"delivery_address" binding:"required"`
	DeliveryLatitude  *float64           `json:"delivery_latitude"`
	DeliveryLongitude *float64           `json:"delivery_longitude"`
	DeliveryDate      time.Time          `json:"delivery_date" binding:"required"`
	DeliveryTime      string             `json:"delivery_time" binding:"required"`
	PaymentMethod     string             `json:"payment_method" binding:"required,oneof=cash card online"`
	CustomerNote      string             `json:"customer_note"`
	Items             []OrderItemInput   `json:"items" binding:"required,min=1"`
}

type OrderItemInput struct {
	MealID              uint   `json:"meal_id" binding:"required"`
	ChefID              uint   `json:"chef_id" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,gt=0"`
	SpecialInstructions string `json:"special_instructions"`
}

type UpdateOrderStatusRequest struct {
	Status   string `json:"status" binding:"required,oneof=pending confirmed preparing ready delivered cancelled"`
	ChefNote string `json:"chef_note"`
}

type UpdateSubOrderStatusRequest struct {
	Status        string `json:"status" binding:"required,oneof=pending confirmed preparing ready delivered cancelled"`
	EstimatedTime int    `json:"estimated_time"`
	ChefNote      string `json:"chef_note"`
}

// Response models
type OrderResponse struct {
	ID                uint                `json:"id"`
	OrderNumber       string              `json:"order_number"`
	UserID            uint                `json:"user_id"`
	Total             float64             `json:"total"`
	Currency          string              `json:"currency"`
	Status            string              `json:"status"`
	DeliveryType      string              `json:"delivery_type"`
	DeliveryAddress   string              `json:"delivery_address"`
	DeliveryLatitude  *float64            `json:"delivery_latitude,omitempty"`
	DeliveryLongitude *float64            `json:"delivery_longitude,omitempty"`
	DeliveryDate      *time.Time          `json:"delivery_date"`
	DeliveryTime      *time.Time          `json:"delivery_time"`
	PaymentMethod     string              `json:"payment_method"`
	PaymentStatus     string              `json:"payment_status"`
	CustomerNote      string              `json:"customer_note"`
	ChefCount         int                 `json:"chef_count"`
	SubOrders         []SubOrderResponse  `json:"sub_orders,omitempty"`
	Items             []OrderItemResponse `json:"items,omitempty"`
	User              *UserInfo           `json:"user,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

type SubOrderResponse struct {
	ID              uint                `json:"id"`
	OrderID         uint                `json:"order_id"`
	ChefID          uint                `json:"chef_id"`
	ChefOrderNumber string              `json:"chef_order_number"`
	Subtotal        float64             `json:"subtotal"`
	DeliveryFee     float64             `json:"delivery_fee"`
	ServiceFee      float64             `json:"service_fee"`
	Total           float64             `json:"total"`
	Status          string              `json:"status"`
	EstimatedTime   int                 `json:"estimated_time"`
	ChefNote        string              `json:"chef_note,omitempty"`
	Chef            *ChefInfo           `json:"chef,omitempty"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID                  uint      `json:"id"`
	OrderID             uint      `json:"order_id"`
	SubOrderID          uint      `json:"sub_order_id"`
	MealID              uint      `json:"meal_id"`
	ChefID              uint      `json:"chef_id"`
	MealName            string    `json:"meal_name,omitempty"`
	MealImage           string    `json:"meal_image,omitempty"`
	ChefName            string    `json:"chef_name,omitempty"`
	KitchenName         string    `json:"kitchen_name,omitempty"`
	Quantity            int       `json:"quantity"`
	Price               float64   `json:"price"`
	Subtotal            float64   `json:"subtotal"`
	SpecialInstructions string    `json:"special_instructions,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// Nearby meals response for location-based recommendations
type NearbyMealsResponse struct {
	Meals   []MealWithDistance `json:"meals"`
	Chefs   []ChefWithDistance `json:"chefs"`
	Categories []CategoryCount `json:"categories"`
	Total   int                `json:"total"`
}

type MealWithDistance struct {
	*Meal
	Distance    float64 `json:"distance_km"`
	DeliveryFee float64 `json:"delivery_fee"`
	ChefInfo    *ChefInfo `json:"chef_info"`
}

type ChefWithDistance struct {
	*Chef
	Distance    float64 `json:"distance_km"`
	MealCount   int     `json:"meal_count"`
	AvgDeliveryTime int `json:"avg_delivery_time"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Chefs    []uint `json:"chef_ids"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type ChefInfo struct {
	ID           uint   `json:"id"`
	BusinessName string `json:"business_name"`
	Location     string `json:"location"`
	Rating       float64 `json:"rating"`
	TotalOrders  int    `json:"total_orders"`
	IsVerified   bool   `json:"is_verified"`
}
