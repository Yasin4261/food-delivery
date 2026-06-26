package router

import (
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
)

// Router builds the application's HTTP route table.
type Router struct {
	mux           *http.ServeMux
	auth          *middleware.Auth
	healthHandler *handler.HealthHandler
	authHandler   *handler.AuthHandler
	chefHandler   *handler.ChefHandler
	menuHandler   *handler.MenuHandler
}

// NewRouter creates a Router with its handler and middleware dependencies.
func NewRouter(
	auth *middleware.Auth,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	chefHandler *handler.ChefHandler,
	menuHandler *handler.MenuHandler,
) *Router {
	return &Router{
		mux:           http.NewServeMux(),
		auth:          auth,
		healthHandler: healthHandler,
		authHandler:   authHandler,
		chefHandler:   chefHandler,
		menuHandler:   menuHandler,
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

	// Chefs: reads are public; opening a profile requires the chef role.
	// The literal /nearby pattern is matched ahead of /{id} by ServeMux.
	r.mux.HandleFunc("GET /api/v2/chefs", r.chefHandler.List)
	r.mux.HandleFunc("GET /api/v2/chefs/nearby", r.chefHandler.Nearby)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}", r.chefHandler.Get)
	r.handleRole("POST /api/v2/chefs", r.chefHandler.Create)

	// A chef's menus and dishes (public reads).
	r.mux.HandleFunc("GET /api/v2/chefs/{id}/menus", r.menuHandler.ListChefMenus)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}/menu-items", r.menuHandler.ListChefItems)

	// Menus: reads are public; mutations require the owning chef.
	r.mux.HandleFunc("GET /api/v2/menus/{id}", r.menuHandler.GetMenu)
	r.mux.HandleFunc("GET /api/v2/menus/{id}/items", r.menuHandler.ListMenuItems)
	r.handleRole("POST /api/v2/menus", r.menuHandler.CreateMenu)
	r.handleRole("PUT /api/v2/menus/{id}", r.menuHandler.UpdateMenu)
	r.handleRole("DELETE /api/v2/menus/{id}", r.menuHandler.DeleteMenu)

	// Dishes: mutations require the owning chef.
	r.handleRole("POST /api/v2/menu-items", r.menuHandler.CreateItem)
	r.handleRole("PUT /api/v2/menu-items/{id}", r.menuHandler.UpdateItem)
	r.handleRole("DELETE /api/v2/menu-items/{id}", r.menuHandler.DeleteItem)

	return r.mux
}

// handleRole registers a chef-only route: it requires a valid token whose role
// is chef before reaching the handler.
func (r *Router) handleRole(pattern string, h http.HandlerFunc) {
	r.mux.Handle(pattern, r.auth.RequireRole(domain.RoleChef)(http.HandlerFunc(h)))
}
