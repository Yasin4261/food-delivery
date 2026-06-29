package router

import (
	"net/http"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
)

// Router builds the application's HTTP route table.
type Router struct {
	mux             *http.ServeMux
	auth            *middleware.Auth
	healthHandler   *handler.HealthHandler
	authHandler     *handler.AuthHandler
	chefHandler     *handler.ChefHandler
	menuHandler     *handler.MenuHandler
	orderHandler    *handler.OrderHandler
	favoriteHandler *handler.FavoriteHandler
	reviewHandler   *handler.ReviewHandler
	earningsHandler *handler.EarningsHandler
	searchHandler   *handler.SearchHandler
	chatHandler     *handler.ChatHandler
}

// NewRouter creates a Router with its handler and middleware dependencies.
func NewRouter(
	auth *middleware.Auth,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	chefHandler *handler.ChefHandler,
	menuHandler *handler.MenuHandler,
	orderHandler *handler.OrderHandler,
	favoriteHandler *handler.FavoriteHandler,
	reviewHandler *handler.ReviewHandler,
	earningsHandler *handler.EarningsHandler,
	searchHandler *handler.SearchHandler,
	chatHandler *handler.ChatHandler,
) *Router {
	return &Router{
		mux:             http.NewServeMux(),
		auth:            auth,
		healthHandler:   healthHandler,
		authHandler:     authHandler,
		chefHandler:     chefHandler,
		menuHandler:     menuHandler,
		orderHandler:    orderHandler,
		favoriteHandler: favoriteHandler,
		reviewHandler:   reviewHandler,
		earningsHandler: earningsHandler,
		searchHandler:   searchHandler,
		chatHandler:     chatHandler,
	}
}

// Setup registers all routes and returns the configured handler.
func (r *Router) Setup() http.Handler {
	r.mux.HandleFunc("GET /health", r.healthHandler.HealthCheck)

	// Public auth routes. The credential/secret-bearing endpoints are rate
	// limited per IP to blunt brute-force / credential-stuffing attempts.
	authLimit := middleware.NewRateLimiter(10, time.Minute)
	limited := func(pattern string, h http.HandlerFunc) {
		r.mux.Handle(pattern, authLimit.Middleware(http.HandlerFunc(h)))
	}
	limited("POST /api/v2/auth/register", r.authHandler.Register)
	limited("POST /api/v2/auth/login", r.authHandler.Login)
	limited("POST /api/v2/auth/forgot-password", r.authHandler.ForgotPassword)
	limited("POST /api/v2/auth/reset-password", r.authHandler.ResetPassword)
	// Logout requires a valid token so it can revoke that exact token.
	r.handleAuth("POST /api/v2/auth/logout", r.authHandler.Logout)

	// Protected: requires a valid bearer token.
	r.mux.Handle("GET /api/v2/auth/me", r.auth.Require(http.HandlerFunc(r.authHandler.Me)))

	// Chefs: reads are public; opening a profile requires the chef role.
	// The literal /nearby pattern is matched ahead of /{id} by ServeMux.
	r.mux.HandleFunc("GET /api/v2/chefs", r.chefHandler.List)
	r.mux.HandleFunc("GET /api/v2/chefs/nearby", r.chefHandler.Nearby)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}", r.chefHandler.Get)
	r.handleRole("POST /api/v2/chefs", r.chefHandler.Create)
	r.handleRole("PATCH /api/v2/chefs/me/status", r.chefHandler.SetStatus)
	r.handleRole("GET /api/v2/chefs/me/earnings", r.earningsHandler.Get)

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

	// Orders: customers place and track their own orders (any authenticated
	// user); the chef-scoped views and status actions require the chef role.
	r.handleAuth("POST /api/v2/orders", r.orderHandler.Place)
	r.handleAuth("GET /api/v2/orders", r.orderHandler.List)
	r.handleAuth("GET /api/v2/orders/{id}", r.orderHandler.Get)
	r.handleAuth("POST /api/v2/orders/{id}/cancel", r.orderHandler.Cancel)
	r.handleRole("GET /api/v2/chef/orders", r.orderHandler.ChefList)
	r.handleRole("POST /api/v2/chef/orders/{id}/status", r.orderHandler.ChefAdvance)

	// Favorites: a customer favoriting chefs (any authenticated user).
	r.handleAuth("GET /api/v2/favorites", r.favoriteHandler.List)
	r.handleAuth("POST /api/v2/favorites/{chefId}", r.favoriteHandler.Add)
	r.handleAuth("DELETE /api/v2/favorites/{chefId}", r.favoriteHandler.Remove)

	// Reviews: customers rate chefs/dishes from their orders (auth); reads are
	// public.
	r.handleAuth("POST /api/v2/reviews", r.reviewHandler.Create)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}/reviews", r.reviewHandler.ListForChef)
	r.mux.HandleFunc("GET /api/v2/menu-items/{id}/reviews", r.reviewHandler.ListForMenuItem)

	// Search: dishes/chefs for any authenticated user; users for admins only.
	r.handleAuth("GET /api/v2/search", r.searchHandler.Search)

	// Chat: customer <-> chef messaging (auth; only participants may access a
	// conversation). The /ws route upgrades to a WebSocket for live delivery.
	r.handleAuth("POST /api/v2/chat/conversations", r.chatHandler.StartConversation)
	r.handleAuth("GET /api/v2/chat/conversations", r.chatHandler.ListConversations)
	r.handleAuth("POST /api/v2/chat/conversations/{id}/messages", r.chatHandler.PostMessage)
	r.handleAuth("GET /api/v2/chat/conversations/{id}/messages", r.chatHandler.ListMessages)
	r.handleAuth("GET /api/v2/chat/conversations/{id}/ws", r.chatHandler.WebSocket)

	return r.mux
}

// handleAuth registers a route that requires a valid bearer token.
func (r *Router) handleAuth(pattern string, h http.HandlerFunc) {
	r.mux.Handle(pattern, r.auth.Require(http.HandlerFunc(h)))
}

// handleRole registers a chef-only route: it requires a valid token whose role
// is chef before reaching the handler.
func (r *Router) handleRole(pattern string, h http.HandlerFunc) {
	r.mux.Handle(pattern, r.auth.RequireRole(domain.RoleChef)(http.HandlerFunc(h)))
}
