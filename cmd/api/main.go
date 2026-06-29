// Command api is the entry point for the food-delivery HTTP API.
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
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	if cfg.AutoMigrate {
		if err := database.RunMigrations(db.DB, "./migrations"); err != nil {
			log.Fatalf("migrations: %v", err)
		}
		log.Println("migrations applied")
	}

	app := initializeApp(db, cfg)

	log.Printf("starting server on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, app); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// initializeApp is the composition root: it constructs the concrete adapters
// and wires them into the core. As features are added, new repositories,
// services and handlers are assembled here.
func initializeApp(db *database.DB, cfg *config.Config) http.Handler {
	// Repositories (driven adapters).
	userRepo := repository.NewUserRepository(db.DB)
	chefRepo := repository.NewChefRepository(db.DB)
	menuRepo := repository.NewMenuRepository(db.DB)
	menuItemRepo := repository.NewMenuItemRepository(db.DB)
	orderRepo := repository.NewOrderRepository(db.DB)
	favoriteRepo := repository.NewFavoriteRepository(db.DB)
	reviewRepo := repository.NewReviewRepository(db.DB)
	earningsRepo := repository.NewEarningsRepository(db.DB)
	searchRepo := repository.NewSearchRepository(db.DB)
	passwordResetRepo := repository.NewPasswordResetRepository(db.DB)

	// Services (use cases).
	authService := service.NewAuthService(userRepo, passwordResetRepo, cfg.JWTSecret, cfg.JWTExpiration)
	chefService := service.NewChefService(chefRepo)
	menuService := service.NewMenuService(chefRepo, menuRepo, menuItemRepo)
	orderService := service.NewOrderService(orderRepo, menuItemRepo, chefRepo)
	favoriteService := service.NewFavoriteService(favoriteRepo, chefRepo)
	reviewService := service.NewReviewService(reviewRepo, orderRepo)
	earningsService := service.NewEarningsService(earningsRepo, chefRepo)
	searchService := service.NewSearchService(searchRepo)

	// Middleware.
	authMiddleware := middleware.NewAuth(authService)

	// Handlers (driving adapters).
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService, cfg.Env != "production")
	chefHandler := handler.NewChefHandler(chefService)
	menuHandler := handler.NewMenuHandler(menuService)
	orderHandler := handler.NewOrderHandler(orderService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	earningsHandler := handler.NewEarningsHandler(earningsService)
	searchHandler := handler.NewSearchHandler(searchService)

	r := router.NewRouter(authMiddleware, healthHandler, authHandler, chefHandler, menuHandler, orderHandler, favoriteHandler, reviewHandler, earningsHandler, searchHandler)
	return r.Setup()
}
