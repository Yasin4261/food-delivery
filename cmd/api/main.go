// Command api is the entry point for the food-delivery HTTP API.
package main

import (
	"log"
	"net/http"

	"github.com/Yasin4261/food-delivery/config"
	"github.com/Yasin4261/food-delivery/database"
	"github.com/Yasin4261/food-delivery/internal/handler"
	"github.com/Yasin4261/food-delivery/internal/router"
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

	app := initializeApp(db)

	log.Printf("starting server on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, app); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// initializeApp is the composition root: it constructs the concrete adapters
// and wires them into the core. As features are added, new repositories,
// services and handlers are assembled here.
func initializeApp(db *database.DB) http.Handler {
	healthHandler := handler.NewHealthHandler(db)

	r := router.NewRouter(healthHandler)
	return r.Setup()
}
