package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Yasin4261/food-delivery/config"
    "github.com/Yasin4261/food-delivery/database"
    "github.com/Yasin4261/food-delivery/internal/handler"
    "github.com/Yasin4261/food-delivery/internal/repository"
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

	// Migrations would go here (if any)
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := database.RunMigrations(db.DB, "./migrations"); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
    }

	// DI
	// Repository Layer
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Service Layer
	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo)

	// Handler Layer
	userHandler := handler.NewUserHandler(userService)
	orderHandler := handler.NewOrderHandler(orderService)
	healthHandler := handler.NewHealthHandler()

	// HTTP Server setup
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", healthHandler.HealthCheck)

	// User routes
	mux.HandleFunc("POST /api/v1/users/register", userHandler.Register)
    mux.HandleFunc("POST /api/v1/users/login", userHandler.Login)
    mux.HandleFunc("GET /api/v1/users/profile", userHandler.GetProfile)
	
	// Order routes
	mux.HandleFunc("POST /api/v1/orders", orderHandler.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}", orderHandler.GetOrder)
	mux.HandleFunc("GET /api/v1/orders", orderHandler.ListOrders)


	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}