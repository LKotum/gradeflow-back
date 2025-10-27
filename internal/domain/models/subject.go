package models

import (
	"time"

	"github.com/google/uuid"
)

// Subject describes a course managed within the system.
type Subject struct {
	ID           uuid.UUID           `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Code         string              `gorm:"type:text;not null;uniqueIndex"`
	Name         string              `gorm:"type:text;not null"`
	Description  *string             `gorm:"type:text"`
	Assignments  []TeachingAssignment `gorm:"foreignKey:SubjectID"`
	GroupLinks   []SubjectGroup       `gorm:"foreignKey:SubjectID"`
	Sessions     []ClassSession       `gorm:"foreignKey:SubjectID"`
	CreatedAt    time.Time           `gorm:"not null"`
	UpdatedAt    time.Time           `gorm:"not null"`
}

// TeachingAssignment links teachers to subjects and optionally to groups.
type TeachingAssignment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TeacherID uuid.UUID `gorm:"type:uuid;not null;index"`
	SubjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Teacher   User       `gorm:"foreignKey:TeacherID;constraint:OnDelete:CASCADE"`
	Subject   Subject    `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
}
