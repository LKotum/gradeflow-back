package models

import "time"

// Попытка сдачи (пересдача/перезачет)
type ExamAttempt struct {
	Base
	AssessmentID string     `gorm:"type:uuid;not null;index"`
	Assessment   Assessment `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE"`
	StudentID    string     `gorm:"type:uuid;not null;index"`
	Student      Student    `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	AttemptNo    int        `gorm:"not null;default:1"`
	DateAt       *time.Time `gorm:"index"`
	ResultScale  string     `gorm:"not null;index"` // passfail|five|hundred
	ValueNum     *int       `gorm:"type:int"`
	ValuePass    *bool
	Notes        string `gorm:"type:text"`
}
