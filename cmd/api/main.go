// Command api is the entry point for the food-delivery HTTP API.
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Yasin4261/food-delivery/config"
	"github.com/Yasin4261/food-delivery/database"
	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/mailer"
	"github.com/Yasin4261/food-delivery/internal/metrics"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/payment"
	"github.com/Yasin4261/food-delivery/internal/redisstore"
	"github.com/Yasin4261/food-delivery/internal/repository"
	"github.com/Yasin4261/food-delivery/internal/router"
	"github.com/Yasin4261/food-delivery/internal/service"
	"github.com/Yasin4261/food-delivery/internal/storage"
)

// version is stamped at build time via:
//
//	go build -ldflags "-X main.version=$(git describe --tags --always --dirty)"
//
// It defaults to "dev" for plain `go run` / `go build`.
var version = "dev"

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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Prometheus instrumentation (#73): the middleware records HTTP metrics for
	// the app; DB-pool stats come from the connection handle.
	m := metrics.New()
	m.RegisterDB(db.DB)

	// Global middleware (outermost first): CORS, then structured per-request
	// logging, then metric recording, then the app.
	app := middleware.CORS(cfg.AllowedOrigins)(
		middleware.RequestLogger(logger)(
			m.Middleware(initializeApp(db, cfg, version, m)),
		),
	)

	// /metrics is served from a top mux OUTSIDE the logging + metrics
	// middleware, so scrapes neither inflate the counters nor spam the logs.
	// Caddy never proxies /metrics, so it is unreachable from the public
	// origin — only Prometheus on the internal network scrapes it.
	root := http.NewServeMux()
	root.Handle("/metrics", m.Handler())
	root.Handle("/", app)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Listen for SIGINT/SIGTERM so we can drain in-flight requests on deploy.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting server on %s (env=%s, version=%s)", srv.Addr, cfg.Env, version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received; draining in-flight requests")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}

// initializeApp is the composition root: it constructs the concrete adapters
// and wires them into the core. As features are added, new repositories,
// services and handlers are assembled here.
func initializeApp(db *database.DB, cfg *config.Config, version string, m *metrics.Metrics) http.Handler {
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
	chatRepo := repository.NewChatRepository(db.DB)
	paymentSessionRepo := repository.NewPaymentSessionRepository(db.DB)

	// Mailer (driven adapter): real SMTP when configured, else the dev logger.
	var mail domain.Mailer
	if cfg.SMTPHost != "" {
		mail = mailer.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.MailFrom)
	} else {
		slog.Warn("SMTP_HOST not set; using the dev logging mailer (emails are logged, not sent)")
		mail = mailer.NewLogging(slog.Default())
	}

	// Payment gateway (driven adapter): iyzico when configured, else the dev
	// mock that simulates the hosted-checkout dance.
	var gateway domain.PaymentGateway
	if cfg.IyzicoAPIKey != "" {
		gateway = payment.NewIyzico(cfg.IyzicoAPIKey, cfg.IyzicoSecretKey, cfg.IyzicoBaseURL)
	} else {
		slog.Warn("IYZICO_API_KEY not set; using the dev mock payment gateway (no real charges)")
		gateway = payment.NewMock(cfg.AppBaseURL)
	}

	// Platform time zone: chefs' working hours are written and evaluated in
	// it (falls back to UTC with a warning rather than failing the boot).
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		slog.Warn("invalid TIMEZONE; falling back to UTC", "timezone", cfg.Timezone, "error", err)
		loc = time.UTC
	}

	// Services (use cases).
	authService := service.NewAuthService(userRepo, passwordResetRepo, mail, cfg.JWTSecret, cfg.JWTExpiration, cfg.AppBaseURL)
	chefHoursRepo := repository.NewChefHoursRepository(db.DB)
	chefService := service.NewChefService(chefRepo, chefHoursRepo, loc)
	menuService := service.NewMenuService(chefRepo, menuRepo, menuItemRepo)
	paymentService := service.NewPaymentService(paymentSessionRepo, orderRepo, userRepo, gateway, cfg.AppBaseURL)
	addressRepo := repository.NewAddressRepository(db.DB)
	addressService := service.NewAddressService(addressRepo)
	adminService := service.NewAdminService(repository.NewAdminRepository(db.DB))
	fileStore, err := storage.NewLocal(cfg.UploadDir)
	if err != nil {
		log.Fatalf("upload storage: %v", err)
	}
	uploadService := service.NewUploadService(fileStore, chefRepo, menuItemRepo)
	orderNotifier := service.NewOrderNotifier(mail, userRepo, chefRepo)
	feePolicy := domain.FeePolicy{
		DeliveryBaseFee:  cfg.DeliveryBaseFee,
		DeliveryFeePerKm: cfg.DeliveryFeePerKm,
		CommissionRate:   cfg.CommissionPercent,
	}
	orderService := service.NewOrderService(orderRepo, menuItemRepo, chefRepo, addressRepo, chefHoursRepo, loc, feePolicy, paymentService, orderNotifier)
	favoriteService := service.NewFavoriteService(favoriteRepo, chefRepo)
	reviewService := service.NewReviewService(reviewRepo, orderRepo)
	earningsService := service.NewEarningsService(earningsRepo, chefRepo)
	searchService := service.NewSearchService(searchRepo)
	chatService := service.NewChatService(chatRepo, chefRepo)

	// Token denylist + auth rate limiter: in-memory by default (correct for a
	// single instance); Redis-backed when REDIS_URL is set, so revocation and
	// limits are shared across instances.
	var revoker service.TokenRevoker = service.NewTokenDenylist()
	var authLimiter middleware.Limiter = middleware.NewRateLimiter(10, time.Minute)
	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("redis: parse REDIS_URL: %v", err)
		}
		rdb := redis.NewClient(opts)
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := rdb.Ping(pingCtx).Err(); err != nil {
			log.Fatalf("redis: ping %s: %v", opts.Addr, err)
		}
		revoker = redisstore.NewDenylist(rdb)
		authLimiter = redisstore.NewRateLimiter(rdb, 10, time.Minute)
		slog.Info("redis-backed token denylist and rate limiter active", "addr", opts.Addr)
	}

	// Middleware.
	authMiddleware := middleware.NewAuth(authService, revoker)

	// Handlers (driving adapters).
	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService, revoker)
	chefHandler := handler.NewChefHandler(chefService)
	menuHandler := handler.NewMenuHandler(menuService)
	orderHandler := handler.NewOrderHandler(orderService, m)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	addressHandler := handler.NewAddressHandler(addressService)
	adminHandler := handler.NewAdminHandler(adminService)
	uploadHandler := handler.NewUploadHandler(uploadService, cfg.UploadDir)
	reviewHandler := handler.NewReviewHandler(reviewService)
	earningsHandler := handler.NewEarningsHandler(earningsService)
	searchHandler := handler.NewSearchHandler(searchService)
	chatHandler := handler.NewChatHandler(chatService)
	versionHandler := handler.NewVersionHandler(version)
	paymentHandler := handler.NewPaymentHandler(paymentService, m)

	r := router.NewRouter(authMiddleware, healthHandler, authHandler, chefHandler, menuHandler, orderHandler, favoriteHandler, addressHandler, uploadHandler, adminHandler, reviewHandler, earningsHandler, searchHandler, chatHandler, versionHandler, paymentHandler, authLimiter)
	return r.Setup()
}
