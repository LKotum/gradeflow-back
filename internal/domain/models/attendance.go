package models

import "time"

// Посещаемость
type Attendance struct {
	Base
	LessonID  string    `gorm:"type:uuid;not null;index"`
	StudentID string    `gorm:"type:uuid;not null;index;uniqueIndex:uniq_attendance_lesson_student,priority:2"`
	Status    string    `gorm:"not null;index"`
	MarkedBy  string    `gorm:"type:text;not null"`
	MarkedAt  time.Time `gorm:"not null;default:now()"`
	_         struct{}  `gorm:"uniqueIndex:uniq_attendance_lesson_student,priority:1"`
}
