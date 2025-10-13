package models

// Запись студента на практику
type PracticeEnrollment struct {
	Base
	PracticeID string   `gorm:"type:uuid;not null;index"`
	Practice   Practice `gorm:"foreignKey:PracticeID;constraint:OnDelete:CASCADE"`
	StudentID  string   `gorm:"type:uuid;not null;index"`
	Student    Student  `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Place      string   `gorm:"index"`
	Status     string   `gorm:"not null;default:'enrolled';index"` // enrolled|completed|failed
}
