package models

import (
	"time"

	"gorm.io/gorm"
)

// Base common fields with UUID
type Base struct {
	ID        string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
