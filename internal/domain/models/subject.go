package models

// Предмет (дисциплина)
type Subject struct {
	Base
	DepartmentID string     `gorm:"type:uuid;not null;index"`
	Department   Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT"`
	Code         string     `gorm:"uniqueIndex;not null"`
	Title        string     `gorm:"not null;index"`
	Credits      int        `gorm:"not null;default:0"`
}
