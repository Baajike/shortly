package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shortly/backend/internal/models"
	"gorm.io/gorm"
)

type ClickRepository interface {
	Create(ctx context.Context, click *models.Click) error
	GetClicksByURLID(ctx context.Context, urlID uuid.UUID) ([]models.Click, error)
	GetClicksGroupedByDay(ctx context.Context, urlID uuid.UUID) ([]models.DailyClick, error)
	GetClicksGroupedByDevice(ctx context.Context, urlID uuid.UUID) ([]models.GroupedStat, error)
	GetClicksGroupedByCountry(ctx context.Context, urlID uuid.UUID) ([]models.GroupedStat, error)
	GetClicksGroupedByBrowser(ctx context.Context, urlID uuid.UUID) ([]models.GroupedStat, error)
}

type clickRepo struct {
	db *gorm.DB
}

func NewClickRepository(db *gorm.DB) ClickRepository {
	return &clickRepo{db: db}
}

func (r *clickRepo) Create(ctx context.Context, click *models.Click) error {
	return r.db.WithContext(ctx).Create(click).Error
}

func (r *clickRepo) GetClicksByURLID(ctx context.Context, urlID uuid.UUID) ([]models.Click, error) {
	var clicks []models.Click
	err := r.db.WithContext(ctx).
		Where("url_id = ?", urlID).
		Order("clicked_at DESC").
		Find(&clicks).Error
	return clicks, err
}

// GetClicksGroupedByDay returns daily click counts for the past 30 days.
func (r *clickRepo) GetClicksGroupedByDay(ctx context.Context, urlID uuid.UUID) ([]models.DailyClick, error) {
	var results []models.DailyClick
	err := r.db.WithContext(ctx).
		Model(&models.Click{}).
		Select("TO_CHAR(clicked_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date, COUNT(*) AS count").
		Where("url_id = ? AND clicked_at >= NOW() - INTERVAL '30 days'", urlID).
		Group("TO_CHAR(clicked_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}

func (r *clickRepo) GetClicksGroupedByDevice(ctx context.Context, urlID uuid.UUID) ([]models.GroupedStat, error) {
	return r.groupBy(ctx, urlID, "device_type")
}

func (r *clickRepo) GetClicksGroupedByCountry(ctx context.Context, urlID uuid.UUID) ([]models.GroupedStat, error) {
	return r.groupBy(ctx, urlID, "country")
}

func (r *clickRepo) GetClicksGroupedByBrowser(ctx context.Context, urlID uuid.UUID) ([]models.GroupedStat, error) {
	return r.groupBy(ctx, urlID, "browser")
}

func (r *clickRepo) groupBy(ctx context.Context, urlID uuid.UUID, column string) ([]models.GroupedStat, error) {
	var results []models.GroupedStat
	err := r.db.WithContext(ctx).
		Model(&models.Click{}).
		Select(column+" AS label, COUNT(*) AS count").
		Where("url_id = ? AND "+column+" != ''", urlID).
		Group(column).
		Order("count DESC").
		Scan(&results).Error
	return results, err
}
