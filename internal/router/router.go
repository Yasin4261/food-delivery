package router

import (
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
)

// Router builds the application's HTTP route table.
type Router struct {
	mux           *http.ServeMux
	auth          *middleware.Auth
	healthHandler *handler.HealthHandler
	authHandler   *handler.AuthHandler
}

// NewRouter creates a Router with its handler and middleware dependencies.
func NewRouter(
	auth *middleware.Auth,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
) *Router {
	return &Router{
		mux:           http.NewServeMux(),
		auth:          auth,
		healthHandler: healthHandler,
		authHandler:   authHandler,
	}
}

// Setup registers all routes and returns the configured handler.
func (r *Router) Setup() http.Handler {
	r.mux.HandleFunc("GET /health", r.healthHandler.HealthCheck)

	// Public auth routes.
	r.mux.HandleFunc("POST /api/v2/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("POST /api/v2/auth/login", r.authHandler.Login)
	r.mux.HandleFunc("POST /api/v2/auth/logout", r.authHandler.Logout)

	// Protected: requires a valid bearer token.
	r.mux.Handle("GET /api/v2/auth/me", r.auth.Require(http.HandlerFunc(r.authHandler.Me)))

	return r.mux
}
