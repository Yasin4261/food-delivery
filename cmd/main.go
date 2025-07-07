package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	
	"ecommerce/config"
	"ecommerce/internal/api"
	"ecommerce/internal/auth"
	"ecommerce/internal/api/handler"
	"ecommerce/internal/repository"
	"ecommerce/internal/service"
)

func main() {
	// Konfigürasyon yükleme
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config yüklenemedi:", err)
	}

	// PostgreSQL bağlantısı
	dbConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal("Veritabanı bağlantısı kurulamadı:", err)
	}
	defer db.Close()

	// Bağlantıyı test et
	if err := db.Ping(); err != nil {
		log.Fatal("Veritabanına bağlanamadı:", err)
	}
	fmt.Println("Veritabanı bağlantısı başarılı!")

	// JWT Manager oluştur
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	// Repository'leri oluştur
	userRepo := repository.NewUserRepository(db)
	chefRepo := repository.NewChefRepository(db)
	mealRepo := repository.NewMealRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(db)

	// Service'leri oluştur
	userService := service.NewUserService(userRepo, jwtManager)
	chefService := service.NewChefService(chefRepo, userRepo)
	mealService := service.NewMealService(mealRepo, chefRepo)
	orderService := service.NewOrderService(orderRepo, mealRepo, cartRepo)
	cartService := service.NewCartService(cartRepo, mealRepo)
	adminService := service.NewAdminService(userRepo, chefRepo, mealRepo, orderRepo)

	// Handler bağımlılıklarını ayarla
	handler.SetDependencies(&handler.HandlerDependencies{
		UserService:  userService,
		ChefService:  chefService,
		MealService:  mealService,
		OrderService: orderService,
		CartService:  cartService,
		AdminService: adminService,
	})

	// Gin router oluşturma
	router := gin.Default()

	// API route'larını ayarlama (JWT manager ile)
	api.SetupRoutes(router, jwtManager)

	// Server başlatma
	port := fmt.Sprintf(":%s", cfg.Server.Port)
	fmt.Printf("Server %s portunda başlatılıyor...\n", port)
	
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal("Server başlatılamadı:", err)
	}
}
