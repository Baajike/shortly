package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shortly/backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicateCode = errors.New("short code already taken")
)

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.URL, error)
	List(ctx context.Context) ([]models.URL, error)
	IncrementClickCount(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type urlRepo struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) URLRepository {
	return &urlRepo{db: db}
}

func (r *urlRepo) Create(ctx context.Context, url *models.URL) error {
	if err := r.db.WithContext(ctx).Create(url).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateCode
		}
		return err
	}
	return nil
}

func (r *urlRepo) FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	var url models.URL
	err := r.db.WithContext(ctx).
		Where("short_code = ?", shortCode).
		First(&url).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &url, err
}

func (r *urlRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.URL, error) {
	var url models.URL
	err := r.db.WithContext(ctx).
		First(&url, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &url, err
}

func (r *urlRepo) List(ctx context.Context) ([]models.URL, error) {
	var urls []models.URL
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&urls).Error
	return urls, err
}

func (r *urlRepo) IncrementClickCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.URL{}).
		Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func (r *urlRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.URL{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// isDuplicateKeyError detects Postgres unique-constraint violations (SQLSTATE 23505).
func isDuplicateKeyError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}
