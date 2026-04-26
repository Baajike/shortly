package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shortly/backend/internal/middleware"
	"github.com/shortly/backend/internal/services"
)

type URLHandler struct {
	urlService       services.URLService
	analyticsService services.AnalyticsService
}

func NewURLHandler(us services.URLService, as services.AnalyticsService) *URLHandler {
	return &URLHandler{urlService: us, analyticsService: as}
}

// POST /api/v1/shorten
func (h *URLHandler) Shorten(c *gin.Context) {
	var req struct {
		URL        string  `json:"url"         binding:"required"`
		CustomSlug *string `json:"custom_slug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.error(c, http.StatusBadRequest, err.Error())
		return
	}

	url, err := h.urlService.ShortenURL(c.Request.Context(), services.ShortenRequest{
		OriginalURL: req.URL,
		CustomSlug:  req.CustomSlug,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidURL):
			h.error(c, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, services.ErrDuplicateSlug):
			h.error(c, http.StatusConflict, err.Error())
		default:
			h.error(c, http.StatusInternalServerError, "failed to shorten URL")
		}
		return
	}

	c.JSON(http.StatusCreated, url)
}

// GET /api/v1/urls
func (h *URLHandler) List(c *gin.Context) {
	urls, err := h.urlService.ListURLs(c.Request.Context())
	if err != nil {
		h.error(c, http.StatusInternalServerError, "failed to list URLs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"urls": urls, "count": len(urls)})
}

// DELETE /api/v1/urls/:id
func (h *URLHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.error(c, http.StatusBadRequest, "invalid URL id")
		return
	}

	if err := h.urlService.DeleteURL(c.Request.Context(), id); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.error(c, http.StatusNotFound, "URL not found")
			return
		}
		h.error(c, http.StatusInternalServerError, "failed to delete URL")
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/urls/:id/analytics
func (h *URLHandler) GetAnalytics(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.error(c, http.StatusBadRequest, "invalid URL id")
		return
	}

	analytics, err := h.analyticsService.GetAnalytics(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.error(c, http.StatusNotFound, "URL not found")
			return
		}
		h.error(c, http.StatusInternalServerError, "failed to fetch analytics")
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *URLHandler) error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error":      message,
		"request_id": c.GetString(middleware.RequestIDKey),
	})
}
