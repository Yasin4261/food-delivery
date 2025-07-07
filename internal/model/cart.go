package model

import (
	"time"
)

// Cart model - Sepet (Ev yemekleri için)
type Cart struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Items []CartItem `json:"items,omitempty" gorm:"foreignKey:CartID"`
}

// CartItem model - Sepet öğesi (Yemek bazlı)
type CartItem struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	CartID   uint `json:"cart_id" gorm:"not null;index"`
	MealID   uint `json:"meal_id" gorm:"not null;index"`
	ChefID   uint `json:"chef_id" gorm:"not null;index"`
	Quantity int  `json:"quantity" gorm:"not null;default:1"`
	
	// Özel Talimatlar
	SpecialInstructions string `json:"special_instructions" gorm:"type:text"`
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Meal *Meal `json:"meal,omitempty" gorm:"foreignKey:MealID"`
	Chef *Chef `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
}

// Cart request/response models
type AddToCartRequest struct {
	MealID              uint   `json:"meal_id" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,gt=0"`
	SpecialInstructions string `json:"special_instructions"`
}

type UpdateCartItemRequest struct {
	Quantity            int    `json:"quantity" binding:"required,gte=0"`
	SpecialInstructions string `json:"special_instructions"`
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
	ID                  uint      `json:"id"`
	MealID              uint      `json:"meal_id"`
	ChefID              uint      `json:"chef_id"`
	MealName            string    `json:"meal_name"`
	MealPrice           float64   `json:"meal_price"`
	MealImage           string    `json:"meal_image"`
	ChefName            string    `json:"chef_name"`
	KitchenName         string    `json:"kitchen_name"`
	Quantity            int       `json:"quantity"`
	Subtotal            float64   `json:"subtotal"`
	SpecialInstructions string    `json:"special_instructions"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CartItemWithProduct - Eski test'lerde kullanılan tip için uyumluluk
type CartItemWithProduct struct {
	CartItemResponse
	Product *Meal `json:"product,omitempty"` // Meal'i Product olarak da döndür
}
