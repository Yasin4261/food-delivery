package router

import (
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/handler"
)

// Router builds the application's HTTP route table.
type Router struct {
	mux           *http.ServeMux
	healthHandler *handler.HealthHandler
}

// NewRouter creates a Router with its handler dependencies.
func NewRouter(healthHandler *handler.HealthHandler) *Router {
	return &Router{
		mux:           http.NewServeMux(),
		healthHandler: healthHandler,
	}
}

// Setup registers all routes and returns the configured handler.
func (r *Router) Setup() http.Handler {
	r.mux.HandleFunc("GET /health", r.healthHandler.HealthCheck)
	return r.mux
}
