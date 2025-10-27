package models

import (
	"time"

	"github.com/google/uuid"
)

// SubjectGroup links a subject with a specific student group.
type SubjectGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SubjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	GroupID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Subject   Subject   `gorm:"constraint:OnDelete:CASCADE"`
	Group     Group     `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// ClassSession represents a concrete lesson scheduled for a group and subject.
type ClassSession struct {
	ID         uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SubjectID  uuid.UUID   `gorm:"type:uuid;not null;index"`
	GroupID    uuid.UUID   `gorm:"type:uuid;not null;index"`
	TeacherID  uuid.UUID   `gorm:"type:uuid;not null;index"`
	StartsAt   time.Time   `gorm:"not null;index"`
	EndsAt     *time.Time  `gorm:""`
	Topic      *string     `gorm:"type:text"`
	Subject    Subject     `gorm:"constraint:OnDelete:CASCADE"`
	Group      Group       `gorm:"constraint:OnDelete:CASCADE"`
	Teacher    User        `gorm:"foreignKey:TeacherID;constraint:OnDelete:CASCADE"`
	Grades     []Grade     `gorm:"foreignKey:SessionID"`
	CreatedAt  time.Time   `gorm:"not null"`
	UpdatedAt  time.Time   `gorm:"not null"`
}
