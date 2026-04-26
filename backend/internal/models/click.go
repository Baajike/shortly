package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Click struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"          json:"id"`
	URLID      uuid.UUID `gorm:"type:uuid;not null;index"      json:"url_id"`
	ClickedAt  time.Time `gorm:"not null;index"                json:"clicked_at"`
	IPAddress  string    `gorm:"size:45"                       json:"ip_address"`
	UserAgent  string    `gorm:"type:text"                     json:"-"`
	Referer    string    `gorm:"type:text"                     json:"referer,omitempty"`
	Country    string    `gorm:"size:2"                        json:"country,omitempty"`
	City       string    `gorm:"size:100"                      json:"city,omitempty"`
	DeviceType string    `gorm:"size:20"                       json:"device_type,omitempty"`
	Browser    string    `gorm:"size:50"                       json:"browser,omitempty"`
	OS         string    `gorm:"size:50"                       json:"os,omitempty"`
}

func (c *Click) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.ClickedAt.IsZero() {
		c.ClickedAt = time.Now().UTC()
	}
	return nil
}
