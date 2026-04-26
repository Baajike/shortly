package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	ShortURL ShortURLConfig
	Rate     RateLimitConfig
}

type AppConfig struct {
	Env     string
	Port    string
	BaseURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type ShortURLConfig struct {
	Length  int
	BaseURL string
}

type RateLimitConfig struct {
	Requests      int
	WindowSeconds int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func Load() (*Config, error) {
	redisDB, _ := strconv.Atoi(env("REDIS_DB", "0"))
	jwtExpiry, _ := strconv.Atoi(env("JWT_EXPIRY_HOURS", "24"))
	shortURLLen, _ := strconv.Atoi(env("SHORT_URL_LENGTH", "7"))
	rateReqs, _ := strconv.Atoi(env("RATE_LIMIT_REQUESTS", "100"))
	rateWindow, _ := strconv.Atoi(env("RATE_LIMIT_WINDOW_SECONDS", "60"))

	cfg := &Config{
		App: AppConfig{
			Env:     env("APP_ENV", "development"),
			Port:    env("APP_PORT", "8080"),
			BaseURL: env("APP_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			User:     env("DB_USER", "shortly"),
			Password: env("DB_PASSWORD", ""),
			Name:     env("DB_NAME", "shortly"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     env("REDIS_HOST", "localhost"),
			Port:     env("REDIS_PORT", "6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret:      env("JWT_SECRET", ""),
			ExpiryHours: jwtExpiry,
		},
		ShortURL: ShortURLConfig{
			Length:  shortURLLen,
			BaseURL: env("SHORT_URL_BASE_URL", "http://localhost:8080"),
		},
		Rate: RateLimitConfig{
			Requests:      rateReqs,
			WindowSeconds: rateWindow,
		},
	}

	if cfg.JWT.Secret == "" && cfg.App.Env != "development" {
		return nil, fmt.Errorf("JWT_SECRET must be set in non-development environments")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
