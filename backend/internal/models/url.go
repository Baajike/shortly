package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type URL struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"                  json:"id"`
	ShortCode   string         `gorm:"uniqueIndex;not null;size:20"           json:"short_code"`
	OriginalURL string         `gorm:"not null;type:text"                     json:"original_url"`
	CustomSlug  *string        `gorm:"uniqueIndex;size:100"                   json:"custom_slug,omitempty"`
	UserID      *uuid.UUID     `gorm:"type:uuid;index"                        json:"user_id,omitempty"`
	ExpiresAt   *time.Time     `                                              json:"expires_at,omitempty"`
	ClickCount  int64          `gorm:"default:0"                              json:"click_count"`
	CreatedAt   time.Time      `                                              json:"created_at"`
	UpdatedAt   time.Time      `                                              json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                  json:"-"`
}

func (u *URL) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *URL) IsExpired() bool {
	return u.ExpiresAt != nil && time.Now().After(*u.ExpiresAt)
}
