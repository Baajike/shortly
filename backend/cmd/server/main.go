package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/shortly/backend/internal/cache"
	"github.com/shortly/backend/internal/config"
	"github.com/shortly/backend/internal/database"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	log.Printf("connected to postgres @ %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	rdb, err := cache.New(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	log.Printf("connected to redis @ %s", cfg.Redis.Addr())

	// TODO: wire up router, run migrations, start HTTP server
	_ = db
	_ = rdb

	log.Printf("shortly ready — listening on :%s (env=%s)", cfg.App.Port, cfg.App.Env)
}
