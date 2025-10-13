package models

import "time"

// Учебная сессия (период)
type AcademicSession struct {
	Base
	Code     string    `gorm:"uniqueIndex;not null"`
	Kind     string    `gorm:"not null;index"`
	StartsAt time.Time `gorm:"not null"`
	EndsAt   time.Time `gorm:"not null"`
}
