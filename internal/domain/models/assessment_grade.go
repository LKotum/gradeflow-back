package models

import "time"

// Оценка по мероприятию
type AssessmentGrade struct {
	Base
	AssessmentID string     `gorm:"type:uuid;not null;index"`
	Assessment   Assessment `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE"`
	StudentID    string     `gorm:"type:uuid;not null;index;uniqueIndex:uniq_assessment_student,priority:2"`
	Scale        string     `gorm:"not null;index"`
	ValueNum     *float64
	ValuePass    *bool
	GradedBy     string    `gorm:"type:text;not null"`
	GradedAt     time.Time `gorm:"not null;default:now();index"`
	_            struct{}  `gorm:"uniqueIndex:uniq_assessment_student,priority:1"`
}
