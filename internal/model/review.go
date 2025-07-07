package model

import (
	"time"
)

// Review - Yemek ve chef değerlendirmeleri
type Review struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	
	// Hangi entitye ait review
	ChefID   *uint `json:"chef_id,omitempty" gorm:"index"`
	MealID   *uint `json:"meal_id,omitempty" gorm:"index"`
	OrderID  uint  `json:"order_id" gorm:"not null;index"`
	
	// Review yapan kullanıcı
	UserID   uint `json:"user_id" gorm:"not null;index"`
	User     User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	
	// Review içeriği
	Rating   int    `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment  string `json:"comment" gorm:"type:text"`
	Title    string `json:"title" gorm:"size:100"`
	
	// Yardımcı bilgiler
	IsVerified bool `json:"is_verified" gorm:"default:false"` // Doğrulanmış alım
	
	// Zaman damgaları
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	Chef  *Chef  `json:"chef,omitempty" gorm:"foreignKey:ChefID"`
	Meal  *Meal  `json:"meal,omitempty" gorm:"foreignKey:MealID"`
	Order *Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`
}

// ReviewRequest - Review oluşturma için
type ReviewRequest struct {
	ChefID  *uint  `json:"chef_id,omitempty"`
	MealID  *uint  `json:"meal_id,omitempty"`
	OrderID uint   `json:"order_id" binding:"required"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=500"`
	Title   string `json:"title" binding:"max=100"`
}

// ReviewResponse - Review gösterimi için
type ReviewResponse struct {
	ID         uint      `json:"id"`
	ChefID     *uint     `json:"chef_id,omitempty"`
	MealID     *uint     `json:"meal_id,omitempty"`
	OrderID    uint      `json:"order_id"`
	UserID     uint      `json:"user_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
	Title      string    `json:"title"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	// İlişkili veriler
	UserName     string `json:"user_name,omitempty"`
	ChefName     string `json:"chef_name,omitempty"`
	MealName     string `json:"meal_name,omitempty"`
	KitchenName  string `json:"kitchen_name,omitempty"`
}

// ReviewSummary - Genel review özeti
type ReviewSummary struct {
	TotalReviews int     `json:"total_reviews"`
	AverageRating float64 `json:"average_rating"`
	RatingDistribution map[int]int `json:"rating_distribution"` // 1-5 arası dağılım
}
