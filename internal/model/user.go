package model

import (
	"time"
)

// User model - Kullanıcı bilgileri (Customer ve Chef için temel model)
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // JSON'da gösterilmez
	FirstName string    `json:"first_name" gorm:"size:50;not null"`
	LastName  string    `json:"last_name" gorm:"size:50;not null"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Role      string    `json:"role" gorm:"size:20;not null;default:'customer'"` // customer, chef, admin
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// İlişkiler
	ChefProfile *Chef    `json:"chef_profile,omitempty" gorm:"foreignKey:UserID"`
	Orders      []Order  `json:"orders,omitempty" gorm:"foreignKey:UserID"`
	Reviews     []Review `json:"reviews,omitempty" gorm:"foreignKey:UserID"`
	Cart        *Cart    `json:"cart,omitempty" gorm:"foreignKey:UserID"`
}

// Auth request/response models
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Role      string `json:"role" binding:"required,oneof=customer chef"` // customer veya chef
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

// UserProfileResponse - Kullanıcı profil bilgilerini döndürmek için
type UserProfileResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Chef bilgileri (eğer chef ise)
	ChefProfile *Chef `json:"chef_profile,omitempty"`
}
