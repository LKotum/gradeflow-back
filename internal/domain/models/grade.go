package models

import (
	"time"

	"github.com/google/uuid"
)

// Grade stores an assessment assigned by a teacher to a student for a subject.
type Grade struct {
	Base
	SessionID  uuid.UUID    `gorm:"type:uuid;not null;index"`
	StudentID  uuid.UUID    `gorm:"type:uuid;not null;index"`
	SubjectID  uuid.UUID    `gorm:"type:uuid;not null;index"`
	TeacherID  uuid.UUID    `gorm:"type:uuid;not null;index"`
	Value      float32      `gorm:"type:real;not null"`
	Notes      *string      `gorm:"type:text"`
	AssessedAt time.Time    `gorm:"not null;index"`
	Session    ClassSession `gorm:"constraint:OnDelete:CASCADE;foreignKey:SessionID;references:ID"`
	Student    User         `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`
	Subject    Subject      `gorm:"constraint:OnDelete:CASCADE;foreignKey:SubjectID;references:ID"`
	Teacher    User         `gorm:"constraint:OnDelete:CASCADE;foreignKey:TeacherID;references:ID"`
}
