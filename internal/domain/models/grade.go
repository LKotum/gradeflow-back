package models

import (
	"time"

	"github.com/google/uuid"
)

// Grade stores an assessment assigned by a teacher to a student for a subject.
type Grade struct {
    Base
    SessionID  uuid.UUID      `gorm:"type:uuid;not null;index"`
    StudentID  uuid.UUID      `gorm:"type:uuid;not null;index"`
    SubjectID  uuid.UUID      `gorm:"type:uuid;not null;index"`
    TeacherID  uuid.UUID      `gorm:"type:uuid;not null;index"`
    Value      float32        `gorm:"type:real;not null"`
    Notes      *string        `gorm:"type:text"`
    AssessedAt time.Time      `gorm:"not null;index"`
    Session    ClassSession   `gorm:"constraint:OnDelete:CASCADE"`
    Student    StudentProfile `gorm:"constraint:OnDelete:CASCADE"`
    Subject    Subject        `gorm:"constraint:OnDelete:CASCADE"`
    Teacher    User           `gorm:"foreignKey:TeacherID;constraint:OnDelete:CASCADE"`
}
