package models

// Group represents a cohort of students.
type Group struct {
	Base
	Name        string           `gorm:"type:text;not null;uniqueIndex"`
	Description *string          `gorm:"type:text"`
	Students    []StudentProfile `gorm:"foreignKey:GroupID"`
	SubjectLinks []SubjectGroup  `gorm:"foreignKey:GroupID"`
}
