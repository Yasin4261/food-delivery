package main

import (
	"log"
	"net/http"

	"github.com/Yasin4261/food-delivery/config"
	"github.com/Yasin4261/food-delivery/database"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/repository"
	"github.com/Yasin4261/food-delivery/internal/router"
	"github.com/Yasin4261/food-delivery/internal/service"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %s", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	defer db.Close()

	// Run migrations if enabled
	if cfg.AutoMigrate {
		if err := database.RunMigrations(db.DB, "./migrations"); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")
	}

	// Initialize dependencies (Dependency Injection)
	app := initializeApp(db, cfg)

	// Setup and start server
	log.Printf("Starting server on port %s (environment: %s)", cfg.Port, cfg.Env)
	log.Printf("API Version: v2")
	log.Printf("JWT Expiration: %s", cfg.JWTExpiration)
	
	if err := http.ListenAndServe(":"+cfg.Port, app); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}

// initializeApp sets up all dependencies and returns the configured router
func initializeApp(db *database.DB, cfg *config.Config) http.Handler {
	// Repository Layer
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	chefRepo := repository.NewChefRepository(db)

	// Service Layer
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	orderService := service.NewOrderService(orderRepo)
	chefService := service.NewChefService(chefRepo)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(userService)

	// Handler Layer
	userHandler := handler.NewUserHandler(userService)
	orderHandler := handler.NewOrderHandler(orderService)
	chefHandler := handler.NewChefHandler(chefService)
	healthHandler := handler.NewHealthHandler()

	// Router
	r := router.NewRouter(
		authMiddleware,
		userHandler,
		orderHandler,
		chefHandler,
		healthHandler,
	)

	return r.SetupRoutes()
}