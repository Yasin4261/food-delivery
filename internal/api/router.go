package api

import (
	"github.com/gin-gonic/gin"
	"ecommerce/internal/api/handler"
	"ecommerce/internal/api/middleware"
	"ecommerce/internal/auth"
	
	// Swagger imports
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
	_ "ecommerce/docs" // swagger docs
)

// SetupRoutes tüm API route'larını ayarlar (Ev yemekleri platformu için)
func SetupRoutes(router *gin.Engine, jwtManager *auth.JWTManager) {
	// Global middleware'ler
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Swagger route'u
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Version info endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":     "1.0.0",
			"app_name":    "Özgür Mutfak API", 
			"build_date":  "2025-07-14",
			"description": "Professional Home-Cooked Meal Marketplace Platform",
			"api_path":    "/api/v1",
			"swagger_url": "/swagger/index.html",
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": "2025-07-14T12:00:00Z",
			"version":   "1.0.0",
		})
	})

	// Get dependencies
	deps := handler.GetDependencies()
	adminHandler := handler.NewAdminHandler(deps.AdminService)

	// API grubu
	api := router.Group("/api/v1")
	
	// Public route'lar (auth gerektirmez)
	public := api.Group("/")
	{
		public.POST("/auth/login", handler.Login)
		public.POST("/auth/register", handler.Register)
		
		// Public meal browsing
		public.GET("/meals", handler.GetMeals)
		public.GET("/meals/:id", handler.GetMeal)
		public.GET("/chefs", handler.GetChefs)
		public.GET("/chefs/:id", handler.GetChef)
		public.GET("/chefs/:id/meals", handler.GetChefMeals)
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

	// Chef specific routes
	chef := api.Group("/chef")
	chef.Use(middleware.AuthRequired(jwtManager))
	chef.Use(middleware.RoleRequired("chef"))
	{
		chef.GET("/profile", handler.GetChefProfile)
		chef.POST("/profile", handler.CreateChefProfile)
		chef.PUT("/profile", handler.UpdateChefProfile)
		
		chef.GET("/meals", handler.GetMyMeals)
		chef.POST("/meals", handler.CreateMeal)
		chef.PUT("/meals/:id", handler.UpdateMeal)
		chef.DELETE("/meals/:id", handler.DeleteMeal)
		chef.PUT("/meals/:id/toggle", handler.ToggleMealAvailability)
		
		chef.GET("/orders", handler.GetChefOrders)
		chef.PUT("/orders/:id/status", handler.UpdateOrderStatus)
	}

	// Admin route'lar
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(jwtManager))
	admin.Use(middleware.AdminRequired())
	{
		// Dashboard
		admin.GET("/dashboard", adminHandler.AdminGetDashboard)
		
		// User management
		admin.GET("/users", adminHandler.AdminGetUsers)
		admin.GET("/users/:id", adminHandler.AdminGetUser)
		
		// Chef management
		admin.GET("/chefs", adminHandler.AdminGetChefs)
		admin.GET("/chefs/pending", adminHandler.AdminGetPendingChefs)
		admin.PUT("/chefs/:id/verify", adminHandler.AdminVerifyChef)
		
		// Order management
		admin.GET("/orders", adminHandler.AdminGetOrders)
		admin.PUT("/orders/:id/status", adminHandler.AdminUpdateOrderStatus)
		
		// Meal management
		admin.GET("/meals", adminHandler.AdminGetMeals)
		admin.PUT("/meals/:id/approve", adminHandler.AdminApproveMeal)
		admin.DELETE("/meals/:id", adminHandler.AdminDeleteMeal)
	}
}
