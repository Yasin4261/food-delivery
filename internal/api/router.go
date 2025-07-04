package api

import (
	"github.com/gin-gonic/gin"
	"ecommerce/internal/api/handler"
	"ecommerce/internal/api/middleware"
	"ecommerce/internal/auth"
)

// SetupRoutes tüm API route'larını ayarlar
func SetupRoutes(router *gin.Engine, jwtManager *auth.JWTManager) {
	// Global middleware'ler
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// API grubu
	api := router.Group("/api/v1")
	
	// Public route'lar (auth gerektirmez)
	public := api.Group("/")
	{
		public.POST("/auth/login", handler.Login)
		public.POST("/auth/register", handler.Register)
		public.GET("/products", handler.GetProducts)
		public.GET("/products/:id", handler.GetProduct)
	}

	// Auth required for logout
	auth := api.Group("/auth")
	auth.Use(middleware.AuthRequired(jwtManager))
	{
		auth.POST("/logout", handler.Logout)
	}

	// Protected route'lar (auth gerektirir)
	protected := api.Group("/")
	protected.Use(middleware.AuthRequired(jwtManager))
	{
		// User endpoints
		protected.GET("/profile", handler.GetProfile)
		protected.PUT("/profile", handler.UpdateProfile)
		
		// Cart endpoints
		protected.GET("/cart", handler.GetCart)
		protected.POST("/cart/items", handler.AddToCart)
		protected.DELETE("/cart/items/:id", handler.RemoveFromCart)
		
		// Order endpoints
		protected.GET("/orders", handler.GetOrders)
		protected.POST("/orders", handler.CreateOrder)
		protected.GET("/orders/:id", handler.GetOrder)
	}

	// Admin route'lar
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(jwtManager))
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/products", handler.AdminGetProducts)
		admin.POST("/products", handler.AdminCreateProduct)
		admin.PUT("/products/:id", handler.AdminUpdateProduct)
		admin.DELETE("/products/:id", handler.AdminDeleteProduct)
		
		admin.GET("/orders", handler.AdminGetOrders)
		admin.PUT("/orders/:id/status", handler.AdminUpdateOrderStatus)
	}
}
