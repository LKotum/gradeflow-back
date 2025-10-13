package models

import "time"

// Занятие (расписание)
type Lesson struct {
	Base
	CourseID string    `gorm:"type:uuid;not null;index"`
	StartsAt time.Time `gorm:"not null;index"`
	EndsAt   time.Time `gorm:"not null;index"`
	Room     string    `gorm:"index"`
	Kind     string    `gorm:"not null;index"`
}
