package router

import (
	"net/http"

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
	addressHandler  *handler.AddressHandler
	uploadHandler   *handler.UploadHandler
	adminHandler    *handler.AdminHandler
	reviewHandler   *handler.ReviewHandler
	earningsHandler *handler.EarningsHandler
	searchHandler   *handler.SearchHandler
	chatHandler     *handler.ChatHandler
	versionHandler  *handler.VersionHandler
	paymentHandler  *handler.PaymentHandler
	authLimiter     middleware.Limiter
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
	addressHandler *handler.AddressHandler,
	uploadHandler *handler.UploadHandler,
	adminHandler *handler.AdminHandler,
	reviewHandler *handler.ReviewHandler,
	earningsHandler *handler.EarningsHandler,
	searchHandler *handler.SearchHandler,
	chatHandler *handler.ChatHandler,
	versionHandler *handler.VersionHandler,
	paymentHandler *handler.PaymentHandler,
	authLimiter middleware.Limiter,
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
		addressHandler:  addressHandler,
		uploadHandler:   uploadHandler,
		adminHandler:    adminHandler,
		reviewHandler:   reviewHandler,
		earningsHandler: earningsHandler,
		searchHandler:   searchHandler,
		chatHandler:     chatHandler,
		versionHandler:  versionHandler,
		paymentHandler:  paymentHandler,
		authLimiter:     authLimiter,
	}
}

// Setup registers all routes and returns the configured handler.
func (r *Router) Setup() http.Handler {
	r.mux.HandleFunc("GET /health", r.healthHandler.HealthCheck)
	r.mux.HandleFunc("GET /version", r.versionHandler.Version)

	// Public auth routes. The credential/secret-bearing endpoints are rate
	// limited per IP to blunt brute-force / credential-stuffing attempts.
	throttle := middleware.RateLimit(r.authLimiter)
	limited := func(pattern string, h http.HandlerFunc) {
		r.mux.Handle(pattern, throttle(http.HandlerFunc(h)))
	}
	limited("POST /api/v2/auth/register", r.authHandler.Register)
	limited("POST /api/v2/auth/login", r.authHandler.Login)
	limited("POST /api/v2/auth/forgot-password", r.authHandler.ForgotPassword)
	limited("POST /api/v2/auth/reset-password", r.authHandler.ResetPassword)
	// Email verification: redeeming a token is public (the link is opened before
	// login); it is rate limited like the rest of the account surface.
	limited("POST /api/v2/auth/verify-email", r.authHandler.VerifyEmail)
	// Logout requires a valid token so it can revoke that exact token.
	r.handleAuth("POST /api/v2/auth/logout", r.authHandler.Logout)
	// Resending a link requires a session and is throttled per IP.
	r.mux.Handle("POST /api/v2/auth/resend-verification",
		throttle(r.auth.Require(http.HandlerFunc(r.authHandler.ResendVerification))))

	// Protected: requires a valid bearer token.
	r.mux.Handle("GET /api/v2/auth/me", r.auth.Require(http.HandlerFunc(r.authHandler.Me)))
	// Profile self-service: password change proves the current password; the
	// profile endpoint edits contact/location only (never email/username/role).
	r.handleAuth("PUT /api/v2/auth/password", r.authHandler.ChangePassword)
	r.handleAuth("PUT /api/v2/users/me", r.authHandler.UpdateProfile)

	// Chefs: reads are public; opening a profile requires the chef role.
	// The literal /nearby pattern is matched ahead of /{id} by ServeMux.
	r.mux.HandleFunc("GET /api/v2/chefs", r.chefHandler.List)
	r.mux.HandleFunc("GET /api/v2/chefs/nearby", r.chefHandler.Nearby)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}", r.chefHandler.Get)
	r.handleRole("POST /api/v2/chefs", r.chefHandler.Create)
	// The literal /chefs/me* patterns are matched ahead of /chefs/{id}.
	r.handleRole("GET /api/v2/chefs/me", r.chefHandler.Me)
	r.handleRole("PUT /api/v2/chefs/me", r.chefHandler.UpdateMe)
	r.handleRole("PATCH /api/v2/chefs/me/status", r.chefHandler.SetStatus)
	// Photos: chef-only uploads (ownership in the service); public serving is
	// gated on the store's generated-name pattern (no traversal).
	r.handleRole("POST /api/v2/chefs/me/image", r.uploadHandler.KitchenImage)
	// Working hours: the chef edits their own schedule; reads are public.
	r.handleRole("PUT /api/v2/chefs/me/hours", r.chefHandler.SetHours)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}/hours", r.chefHandler.Hours)
	r.handleRole("POST /api/v2/menu-items/{id}/image", r.uploadHandler.DishImage)
	r.handleRole("POST /api/v2/menu-items/{id}/images", r.uploadHandler.DishGalleryAdd)
	r.handleRole("DELETE /api/v2/menu-items/{id}/images", r.uploadHandler.DishGalleryRemove)
	r.mux.HandleFunc("GET /uploads/{file}", r.uploadHandler.Serve)
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

	// Card payments: opening a checkout requires the order's owner; the
	// gateway's browser callback is necessarily public (no bearer token on
	// that hop) — the outcome is verified server-to-server and the endpoint is
	// rate limited like the auth surface.
	r.handleAuth("POST /api/v2/orders/{id}/pay", r.paymentHandler.Pay)
	r.mux.Handle("POST /api/v2/payments/callback", throttle(http.HandlerFunc(r.paymentHandler.Callback)))

	// Favorites: a customer favoriting chefs (any authenticated user).
	r.handleAuth("GET /api/v2/favorites", r.favoriteHandler.List)
	r.handleAuth("POST /api/v2/favorites/{chefId}", r.favoriteHandler.Add)
	r.handleAuth("DELETE /api/v2/favorites/{chefId}", r.favoriteHandler.Remove)

	// Address book: auth-only; the service enforces per-address ownership.
	r.handleAuth("GET /api/v2/addresses", r.addressHandler.List)
	r.handleAuth("POST /api/v2/addresses", r.addressHandler.Create)
	r.handleAuth("PUT /api/v2/addresses/{id}", r.addressHandler.Update)
	r.handleAuth("DELETE /api/v2/addresses/{id}", r.addressHandler.Delete)

	// Reviews: customers rate chefs/dishes from their orders (auth); reads are
	// public.
	r.handleAuth("POST /api/v2/reviews", r.reviewHandler.Create)
	r.handleAuth("GET /api/v2/orders/{id}/reviews", r.reviewHandler.ListForOrder)
	r.mux.HandleFunc("GET /api/v2/chefs/{id}/reviews", r.reviewHandler.ListForChef)
	r.mux.HandleFunc("GET /api/v2/menu-items/{id}/reviews", r.reviewHandler.ListForMenuItem)

	// Search: dishes/chefs for any authenticated user; users for admins only.
	r.handleAuth("GET /api/v2/search", r.searchHandler.Search)

	// Admin: platform moderation + stats (admin role only). The service layer
	// carries no per-user ownership here — the role guard is the boundary.
	r.handleAdmin("GET /api/v2/admin/stats", r.adminHandler.Stats)
	r.handleAdmin("GET /api/v2/admin/users", r.adminHandler.ListUsers)
	r.handleAdmin("PATCH /api/v2/admin/users/{id}/active", r.adminHandler.SetUserActive)
	r.handleAdmin("GET /api/v2/admin/orders", r.adminHandler.ListOrders)
	r.handleAdmin("GET /api/v2/admin/chefs", r.adminHandler.ListChefs)
	r.handleAdmin("PATCH /api/v2/admin/chefs/{id}/active", r.adminHandler.SetChefActive)
	r.handleAdmin("GET /api/v2/admin/promos", r.adminHandler.ListPromos)
	r.handleAdmin("POST /api/v2/admin/promos", r.adminHandler.CreatePromo)
	r.handleAdmin("PATCH /api/v2/admin/promos/{id}/active", r.adminHandler.SetPromoActive)

	// Notification badge counts, polled by the SPA.
	r.handleAuth("GET /api/v2/notifications/summary", r.orderHandler.Summary)

	// Chat: customer <-> chef messaging (auth; only participants may access a
	// conversation). The /ws route upgrades to a WebSocket for live delivery.
	r.handleAuth("POST /api/v2/chat/conversations", r.chatHandler.StartConversation)
	r.handleAuth("GET /api/v2/chat/conversations", r.chatHandler.ListConversations)
	r.handleAuth("POST /api/v2/chat/conversations/{id}/messages", r.chatHandler.PostMessage)
	r.handleAuth("GET /api/v2/chat/conversations/{id}/messages", r.chatHandler.ListMessages)
	r.handleAuth("POST /api/v2/chat/conversations/{id}/read", r.chatHandler.MarkRead)
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

// handleAdmin registers an admin-only route (RequireRole(admin) → 403 otherwise).
func (r *Router) handleAdmin(pattern string, h http.HandlerFunc) {
	r.mux.Handle(pattern, r.auth.RequireRole(domain.RoleAdmin)(http.HandlerFunc(h)))
}
