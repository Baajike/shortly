package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shortly/backend/internal/models"
	"github.com/shortly/backend/internal/repository"
	"github.com/shortly/backend/pkg/utils"
)

// ClickInput carries the raw HTTP request data needed to record a click.
type ClickInput struct {
	URLID     uuid.UUID
	IPAddress string
	UserAgent string
	Referer   string
}

type AnalyticsService interface {
	RecordClick(ctx context.Context, input ClickInput) error
	GetAnalytics(ctx context.Context, urlID uuid.UUID) (*models.Analytics, error)
}

type analyticsService struct {
	clickRepo repository.ClickRepository
	urlRepo   repository.URLRepository
	httpClient *http.Client
}

func NewAnalyticsService(clickRepo repository.ClickRepository, urlRepo repository.URLRepository) AnalyticsService {
	return &analyticsService{
		clickRepo:  clickRepo,
		urlRepo:    urlRepo,
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}
}

func (s *analyticsService) RecordClick(ctx context.Context, input ClickInput) error {
	parsed := utils.ParseUserAgent(input.UserAgent)
	country, city := s.geolocate(input.IPAddress)

	click := &models.Click{
		URLID:      input.URLID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		Referer:    input.Referer,
		Country:    country,
		City:       city,
		DeviceType: parsed.DeviceType,
		Browser:    parsed.Browser,
		OS:         parsed.OS,
	}

	if err := s.clickRepo.Create(ctx, click); err != nil {
		return err
	}

	// Increment the denormalized counter on the URL row.
	if err := s.urlRepo.IncrementClickCount(ctx, input.URLID); err != nil {
		log.Printf("warn: increment click count for %s: %v", input.URLID, err)
	}

	return nil
}

func (s *analyticsService) GetAnalytics(ctx context.Context, urlID uuid.UUID) (*models.Analytics, error) {
	if _, err := s.urlRepo.FindByID(ctx, urlID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	byDay, err := s.clickRepo.GetClicksGroupedByDay(ctx, urlID)
	if err != nil {
		return nil, err
	}

	byDevice, err := s.clickRepo.GetClicksGroupedByDevice(ctx, urlID)
	if err != nil {
		return nil, err
	}

	byCountry, err := s.clickRepo.GetClicksGroupedByCountry(ctx, urlID)
	if err != nil {
		return nil, err
	}

	byBrowser, err := s.clickRepo.GetClicksGroupedByBrowser(ctx, urlID)
	if err != nil {
		return nil, err
	}

	return &models.Analytics{
		// url.ClickCount is the true lifetime total; byDay only covers 30 days.
		TotalClicks:     url.ClickCount,
		ClicksByDay:     byDay,
		ClicksByDevice:  byDevice,
		ClicksByCountry: byCountry,
		ClicksByBrowser: byBrowser,
	}, nil
}

// geolocate calls ip-api.com (free tier, no key required).
// Returns empty strings on any error so click recording is never blocked.
func (s *analyticsService) geolocate(ipStr string) (country, city string) {
	if isPrivateIP(ipStr) {
		return "", ""
	}

	resp, err := s.httpClient.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,countryCode,city", ipStr))
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	var result struct {
		Status      string `json:"status"`
		CountryCode string `json:"countryCode"`
		City        string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", ""
	}
	if result.Status != "success" {
		return "", ""
	}
	return result.CountryCode, result.City
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
}
