package models

import "time"

// Экзаменационная сессия (зимняя/летняя и пр.)
type ExamSession struct {
	Base
	AcademicSessionID string          `gorm:"type:uuid;not null;index"`
	AcademicSession   AcademicSession `gorm:"foreignKey:AcademicSessionID;constraint:OnDelete:RESTRICT"`
	Name              string          `gorm:"not null;index"` // Winter 2025, etc.
	StartsAt          *time.Time      `gorm:"index"`
	EndsAt            *time.Time      `gorm:"index"`
}
