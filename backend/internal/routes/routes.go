package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shortly/backend/internal/config"
	"github.com/shortly/backend/internal/handlers"
	"github.com/shortly/backend/internal/middleware"
)

// Setup wires all routes onto r.
// The redirect route (/:shortCode) is registered last so Gin's static
// routes always take priority over the wildcard parameter.
func Setup(
	r *gin.Engine,
	cfg *config.Config,
	rdb *redis.Client,
	urlHandler *handlers.URLHandler,
	redirectHandler *handlers.RedirectHandler,
) {
	// In development allow any origin so the Vite dev server (port 5173)
	// can reach the backend (port 8080) without extra config.
	corsOrigin := cfg.App.BaseURL
	if cfg.App.Env != "production" {
		corsOrigin = "*"
	}

	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(corsOrigin))

	// Health / readiness probe
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "env": cfg.App.Env})
	})

	// REST API
	api := r.Group("/api/v1")
	{
		limiter := middleware.NewRateLimiter(rdb, cfg.Rate.Requests, time.Duration(cfg.Rate.WindowSeconds)*time.Second)

		api.POST("/shorten", limiter.Limit(), urlHandler.Shorten)
		api.GET("/urls", urlHandler.List)
		api.DELETE("/urls/:id", urlHandler.Delete)
		api.GET("/urls/:id/analytics", urlHandler.GetAnalytics)
	}

	// Short-URL redirect — registered after all static routes
	r.GET("/:shortCode", redirectHandler.Redirect)
}
