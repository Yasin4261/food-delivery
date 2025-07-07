package model

import (
	"time"
)

// Meal - Ev yemekleri 
type Meal struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	ChefID   uint   `json:"chef_id" gorm:"not null"`
	Chef     Chef   `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
	
	// Yemek Bilgileri
	Name         string  `json:"name" gorm:"size:100;not null"`
	Description  string  `json:"description" gorm:"type:text"`
	Category     string  `json:"category" gorm:"size:50"`        // "Ana Yemek", "Çorba", "Tatlı", vb.
	Cuisine      string  `json:"cuisine" gorm:"size:50"`        // "Türk", "İtalyan", "Çin", vb.
	
	// Fiyat ve Miktar
	Price        float64 `json:"price" gorm:"not null"`
	Currency     string  `json:"currency" gorm:"size:3;default:'TRY'"`
	Portion      string  `json:"portion" gorm:"size:50"`        // "1 kişilik", "Aile boyu", vb.
	ServingSize  int     `json:"serving_size" gorm:"default:1"` // Kaç kişilik
	
	// Stok ve Hazırlık
	AvailableQuantity int    `json:"available_quantity" gorm:"default:0"`
	PreparationTime   int    `json:"preparation_time"`    // Dakika olarak
	CookingTime       int    `json:"cooking_time"`        // Dakika olarak
	
	// Beslenme Bilgileri (opsiyonel)
	Calories      int    `json:"calories,omitempty"`
	Ingredients   string `json:"ingredients" gorm:"type:text"`    // Malzemeler
	Allergens     string `json:"allergens" gorm:"type:text"`      // Alerjen bilgileri
	IsVegetarian  bool   `json:"is_vegetarian" gorm:"default:false"`
	IsVegan       bool   `json:"is_vegan" gorm:"default:false"`
	IsGlutenFree  bool   `json:"is_gluten_free" gorm:"default:false"`
	
	// Durum Bilgileri
	IsActive      bool    `json:"is_active" gorm:"default:true"`
	IsAvailable   bool    `json:"is_available" gorm:"default:true"`
	Rating        float64 `json:"rating" gorm:"default:0"`
	TotalOrders   int     `json:"total_orders" gorm:"default:0"`
	
	// Medya
	Images        string `json:"images" gorm:"type:text"` // JSON array of image URLs
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:MealID"`
	Reviews    []Review    `json:"reviews,omitempty" gorm:"foreignKey:MealID"`
	CartItems  []CartItem  `json:"cart_items,omitempty" gorm:"foreignKey:MealID"`
}

// MealRequest - Yemek oluşturma/güncelleme için
type MealRequest struct {
	Name              string  `json:"name" binding:"required,min=3,max=100"`
	Description       string  `json:"description" binding:"max=500"`
	Category          string  `json:"category" binding:"required"`
	Cuisine           string  `json:"cuisine"`
	Price             float64 `json:"price" binding:"required,gt=0"`
	Portion           string  `json:"portion"`
	ServingSize       int     `json:"serving_size" binding:"min=1,max=20"`
	AvailableQuantity int     `json:"available_quantity" binding:"min=0"`
	PreparationTime   int     `json:"preparation_time" binding:"min=0"`
	CookingTime       int     `json:"cooking_time" binding:"min=0"`
	Calories          int     `json:"calories"`
	Ingredients       string  `json:"ingredients"`
	Allergens         string  `json:"allergens"`
	IsVegetarian      bool    `json:"is_vegetarian"`
	IsVegan           bool    `json:"is_vegan"`
	IsGlutenFree      bool    `json:"is_gluten_free"`
}

// MealFilter - Yemek filtreleme için
type MealFilter struct {
	Category     string  `json:"category"`
	Cuisine      string  `json:"cuisine"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	IsVegetarian bool    `json:"is_vegetarian"`
	IsVegan      bool    `json:"is_vegan"`
	IsGlutenFree bool    `json:"is_gluten_free"`
	ChefID       uint    `json:"chef_id"`
	District     string  `json:"district"`
	City         string  `json:"city"`
	Rating       float64 `json:"min_rating"`
	
	// Konum bazlı arama
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"` // km olarak
}

// MealWithChef - Chef bilgileri ile birlikte yemek
type MealWithChef struct {
	Meal
	ChefName     string  `json:"chef_name"`
	KitchenName  string  `json:"kitchen_name"`
	ChefRating   float64 `json:"chef_rating"`
	ChefDistrict string  `json:"chef_district"`
	ChefCity     string  `json:"chef_city"`
	Distance     float64 `json:"distance,omitempty"` // km olarak
}
