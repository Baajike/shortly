package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shortly/backend/internal/config"
	"github.com/shortly/backend/internal/models"
	"github.com/shortly/backend/internal/repository"
	"github.com/shortly/backend/pkg/utils"
)

const (
	cachePrefix = "url:code:"
	cacheTTL    = time.Hour
	maxRetries  = 5
)

var (
	ErrNotFound      = errors.New("URL not found")
	ErrExpired       = errors.New("URL has expired")
	ErrDuplicateSlug = errors.New("custom slug is already taken")
	ErrInvalidURL    = errors.New("invalid URL")
)

// ShortenRequest is the input to ShortenURL.
type ShortenRequest struct {
	OriginalURL string     `json:"url"`
	CustomSlug  *string    `json:"custom_slug"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type URLService interface {
	ShortenURL(ctx context.Context, req ShortenRequest) (*models.URL, error)
	ResolveURL(ctx context.Context, shortCode string) (*models.URL, error)
	DeleteURL(ctx context.Context, id uuid.UUID) error
	ListURLs(ctx context.Context) ([]models.URL, error)
}

type urlService struct {
	repo   repository.URLRepository
	rdb    *redis.Client
	cfg    *config.Config
}

func NewURLService(repo repository.URLRepository, rdb *redis.Client, cfg *config.Config) URLService {
	return &urlService{repo: repo, rdb: rdb, cfg: cfg}
}

func (s *urlService) ShortenURL(ctx context.Context, req ShortenRequest) (*models.URL, error) {
	if err := utils.ValidateURL(req.OriginalURL); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	url := &models.URL{
		OriginalURL: req.OriginalURL,
		CustomSlug:  req.CustomSlug,
		ExpiresAt:   req.ExpiresAt,
	}

	if req.CustomSlug != nil && *req.CustomSlug != "" {
		url.ShortCode = *req.CustomSlug
		if err := s.repo.Create(ctx, url); err != nil {
			if errors.Is(err, repository.ErrDuplicateCode) {
				return nil, ErrDuplicateSlug
			}
			return nil, err
		}
		return url, nil
	}

	// Auto-generate a nanoid, retry on the rare collision.
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := utils.GenerateNanoID(s.cfg.ShortURL.Length)
		if err != nil {
			return nil, err
		}
		url.ShortCode = code

		if err := s.repo.Create(ctx, url); err != nil {
			if errors.Is(err, repository.ErrDuplicateCode) && attempt < maxRetries-1 {
				continue
			}
			return nil, err
		}
		return url, nil
	}

	return nil, errors.New("failed to generate a unique short code")
}

func (s *urlService) ResolveURL(ctx context.Context, shortCode string) (*models.URL, error) {
	// 1. Cache look-up.
	if cached, err := s.getFromCache(ctx, shortCode); err == nil {
		if cached.IsExpired() {
			return nil, ErrExpired
		}
		return cached, nil
	}

	// 2. Database fallback.
	url, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if url.IsExpired() {
		return nil, ErrExpired
	}

	// 3. Populate cache for next request.
	_ = s.setCache(ctx, url)

	return url, nil
}

func (s *urlService) DeleteURL(ctx context.Context, id uuid.UUID) error {
	url, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache entry.
	_ = s.rdb.Del(ctx, cachePrefix+url.ShortCode)
	return nil
}

func (s *urlService) ListURLs(ctx context.Context) ([]models.URL, error) {
	return s.repo.List(ctx)
}

func (s *urlService) getFromCache(ctx context.Context, shortCode string) (*models.URL, error) {
	val, err := s.rdb.Get(ctx, cachePrefix+shortCode).Bytes()
	if err != nil {
		return nil, err
	}
	var url models.URL
	if err := json.Unmarshal(val, &url); err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *urlService) setCache(ctx context.Context, url *models.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	ttl := cacheTTL
	if url.ExpiresAt != nil {
		remaining := time.Until(*url.ExpiresAt)
		if remaining < ttl {
			ttl = remaining
		}
	}
	return s.rdb.Set(ctx, cachePrefix+url.ShortCode, data, ttl).Err()
}
