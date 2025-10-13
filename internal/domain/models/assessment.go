package models

import "time"

// ОЦЕНОЧНОЕ МЕРОПРИЯТИЕ
type Assessment struct {
	Base
	CourseID string    `gorm:"type:uuid;not null;index"`
	Type     string    `gorm:"not null;index"`
	DateAt   time.Time `gorm:"not null;index"`
	Room     string    `gorm:"index"`
	Scale    string    `gorm:"not null;index"`
	MaxPts   *float64
}
