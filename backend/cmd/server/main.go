package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/shortly/backend/internal/cache"
	"github.com/shortly/backend/internal/config"
	"github.com/shortly/backend/internal/database"
	"github.com/shortly/backend/internal/handlers"
	"github.com/shortly/backend/internal/models"
	"github.com/shortly/backend/internal/repository"
	"github.com/shortly/backend/internal/routes"
	"github.com/shortly/backend/internal/services"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	log.Printf("connected to postgres @ %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	if err := db.AutoMigrate(&models.URL{}, &models.Click{}); err != nil {
		log.Fatalf("automigrate: %v", err)
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb, err := cache.New(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	log.Printf("connected to redis @ %s", cfg.Redis.Addr())

	// ── Dependency graph ──────────────────────────────────────────────────────
	urlRepo   := repository.NewURLRepository(db)
	clickRepo := repository.NewClickRepository(db)

	urlSvc       := services.NewURLService(urlRepo, rdb, cfg)
	analyticsSvc := services.NewAnalyticsService(clickRepo, urlRepo)

	urlHandler      := handlers.NewURLHandler(urlSvc, analyticsSvc)
	redirectHandler := handlers.NewRedirectHandler(urlSvc, analyticsSvc)

	// ── Router ────────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	routes.Setup(r, cfg, rdb, urlHandler, redirectHandler)

	// ── HTTP server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("shortly listening on :%s (env=%s)", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
