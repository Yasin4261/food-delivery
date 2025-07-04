package model

import (
	"time"
)

// Admin dashboard statistics
type DashboardStats struct {
	TotalUsers     int     `json:"total_users"`
	TotalOrders    int     `json:"total_orders"`
	TotalProducts  int     `json:"total_products"`
	TotalRevenue   float64 `json:"total_revenue"`
	PendingOrders  int     `json:"pending_orders"`
	LastUpdated    time.Time `json:"last_updated"`
}

// Common response structures
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Health check response
type HealthCheckResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// API Info response
type APIInfoResponse struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Timestamp   time.Time `json:"timestamp"`
	Endpoints   []string  `json:"endpoints,omitempty"`
}

// Generic list response
type ListResponse struct {
	Data       interface{} `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// JWT Claims structure
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
