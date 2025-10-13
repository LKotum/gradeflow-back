package models

// Направление/программа
type Program struct {
	Base
	DepartmentID string     `gorm:"type:uuid;not null;index"`
	Department   Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:RESTRICT"`
	Code         string     `gorm:"uniqueIndex;not null"`
	Name         string     `gorm:"not null;index"`
}
