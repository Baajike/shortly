package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shortly/backend/internal/middleware"
	"github.com/shortly/backend/internal/services"
)

type RedirectHandler struct {
	urlService       services.URLService
	analyticsService services.AnalyticsService
}

func NewRedirectHandler(us services.URLService, as services.AnalyticsService) *RedirectHandler {
	return &RedirectHandler{urlService: us, analyticsService: as}
}

// GET /:shortCode
func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")

	url, err := h.urlService.ResolveURL(c.Request.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "short URL not found",
				"request_id": c.GetString(middleware.RequestIDKey),
			})
		case errors.Is(err, services.ErrExpired):
			c.JSON(http.StatusGone, gin.H{
				"error":      "this short URL has expired",
				"request_id": c.GetString(middleware.RequestIDKey),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	// Snapshot everything from the request before the goroutine runs —
	// Gin recycles the context after the handler returns.
	input := services.ClickInput{
		URLID:     url.ID,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.analyticsService.RecordClick(ctx, input); err != nil {
			log.Printf("record click for %s: %v", shortCode, err)
		}
	}()

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}
