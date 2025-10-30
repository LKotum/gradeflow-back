package models

import "github.com/google/uuid"

// Subject describes a course managed within the system.
type Subject struct {
	Base
	Code        string               `gorm:"type:text;not null;uniqueIndex"`
	Name        string               `gorm:"type:text;not null"`
	Description *string              `gorm:"type:text"`
	Assignments []TeachingAssignment `gorm:"foreignKey:SubjectID;references:ID"`
	GroupLinks  []SubjectGroup       `gorm:"foreignKey:SubjectID;references:ID"`
	Sessions    []ClassSession       `gorm:"foreignKey:SubjectID;references:ID"`
}

// TeachingAssignment links teachers to subjects and optionally to groups.
type TeachingAssignment struct {
	Base
	TeacherID uuid.UUID `gorm:"type:uuid;not null;index"`
	SubjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Teacher   User      `gorm:"constraint:OnDelete:CASCADE;foreignKey:TeacherID;references:ID"`
	Subject   Subject   `gorm:"constraint:OnDelete:CASCADE;foreignKey:SubjectID;references:ID"`
}
