package models

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a cohort of students.
type Group struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string           `gorm:"type:text;not null;uniqueIndex"`
	Description *string          `gorm:"type:text"`
	Students    []StudentProfile `gorm:"foreignKey:GroupID"`
	SubjectLinks []SubjectGroup  `gorm:"foreignKey:GroupID"`
	CreatedAt   time.Time        `gorm:"not null"`
	UpdatedAt   time.Time        `gorm:"not null"`
}
