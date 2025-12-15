package router

import (
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
)

// Router holds all route handlers and middleware
type Router struct {
	mux            *http.ServeMux
	authMiddleware *middleware.AuthMiddleware
	userHandler    *handler.UserHandler
	orderHandler   *handler.OrderHandler
	chefHandler    *handler.ChefHandler
	healthHandler  *handler.HealthHandler
}

// NewRouter creates a new router instance
func NewRouter(
	authMiddleware *middleware.AuthMiddleware,
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
	chefHandler *handler.ChefHandler,
	healthHandler *handler.HealthHandler,
) *Router {
	return &Router{
		mux:            http.NewServeMux(),
		authMiddleware: authMiddleware,
		userHandler:    userHandler,
		orderHandler:   orderHandler,
		chefHandler:    chefHandler,
		healthHandler:  healthHandler,
	}
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes() *http.ServeMux {
	// Health check endpoint
	r.mux.HandleFunc("GET /health", r.healthHandler.HealthCheck)

	// API v2 routes
	r.setupV2Routes()

	return r.mux
}

// setupV2Routes configures version 2 API routes
func (r *Router) setupV2Routes() {
	// Public auth routes (no authentication required)
	r.mux.HandleFunc("POST /api/v2/auth/register", r.userHandler.Register)
	r.mux.HandleFunc("POST /api/v2/auth/login", r.userHandler.Login)

	// Protected user routes (authentication required)
	r.mux.Handle("GET /api/v2/auth/profile",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.GetProfile)))
	r.mux.Handle("PUT /api/v2/auth/profile",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.UpdateProfile)))
	r.mux.Handle("PUT /api/v2/auth/location",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.UpdateLocation)))
	r.mux.Handle("POST /api/v2/auth/change-password",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.ChangePassword)))

	// User management routes
	r.mux.Handle("GET /api/v2/users",
		r.authMiddleware.Authenticate(
			r.authMiddleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(r.userHandler.ListUsers)),
		))
	r.mux.Handle("GET /api/v2/users/nearby",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.userHandler.FindNearbyUsers)))

	// Chef routes
	r.mux.Handle("POST /api/v2/chefs",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.chefHandler.CreateChef)))
	r.mux.HandleFunc("GET /api/v2/chefs/{id}", r.chefHandler.GetChef)
	r.mux.HandleFunc("GET /api/v2/chefs", r.chefHandler.ListChefs)
	r.mux.HandleFunc("GET /api/v2/chefs/nearby", r.chefHandler.FindNearbyChefs)

	// Order routes (authentication required)
	r.mux.Handle("POST /api/v2/orders",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.orderHandler.CreateOrder)))
	r.mux.Handle("GET /api/v2/orders/{id}",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.orderHandler.GetOrder)))
	r.mux.Handle("GET /api/v2/orders",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.orderHandler.ListOrders)))
}
