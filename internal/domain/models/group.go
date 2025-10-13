package models

// Группа
type Group struct {
	Base
	ProgramID string  `gorm:"type:uuid;not null;index"`
	Program   Program `gorm:"foreignKey:ProgramID;constraint:OnDelete:RESTRICT"`
	Code      string  `gorm:"uniqueIndex;not null"`
	Name      string  `gorm:"not null;index"`
	Year      int     `gorm:"not null;index"`
}
