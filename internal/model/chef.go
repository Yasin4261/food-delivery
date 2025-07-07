package model

import (
	"time"
)

// Chef - Evde yemek yapan kişiler
type Chef struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	
	// Chef Bilgileri
	KitchenName    string `json:"kitchen_name" gorm:"size:100;not null"` // "Ayşe'nin Mutfağı"
	Description    string `json:"description" gorm:"type:text"`
	Speciality     string `json:"speciality" gorm:"size:100"`     // "Ev yemekleri", "Tatlılar", vb.
	Experience     int    `json:"experience"`                      // Yıl olarak
	
	// Adres Bilgileri
	Address     string  `json:"address" gorm:"type:text;not null"`
	District    string  `json:"district" gorm:"size:50"`
	City        string  `json:"city" gorm:"size:50"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	
	// Durum Bilgileri
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	IsVerified   bool      `json:"is_verified" gorm:"default:false"`
	Rating       float64   `json:"rating" gorm:"default:0"`
	TotalOrders  int       `json:"total_orders" gorm:"default:0"`
	
	// Zaman Damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Meals   []Meal   `json:"meals,omitempty" gorm:"foreignKey:ChefID"`
	Reviews []Review `json:"reviews,omitempty" gorm:"foreignKey:ChefID"`
}

// ChefProfile - Chef profil güncelleme için
type ChefProfile struct {
	KitchenName string `json:"kitchen_name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
	Speciality  string `json:"speciality" binding:"max=100"`
	Experience  int    `json:"experience" binding:"min=0,max=50"`
	Address     string `json:"address" binding:"required"`
	District    string `json:"district" binding:"required"`
	City        string `json:"city" binding:"required"`
}

// ChefStats - Chef istatistikleri
type ChefStats struct {
	TotalMeals      int     `json:"total_meals"`
	TotalOrders     int     `json:"total_orders"`
	AverageRating   float64 `json:"average_rating"`
	TotalRevenue    float64 `json:"total_revenue"`
	ActiveMeals     int     `json:"active_meals"`
	MonthlyOrders   int     `json:"monthly_orders"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`
}
