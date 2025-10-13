package models

import "time"

// Практика (учебная/производственная)
type Practice struct {
	Base
	Title        string      `gorm:"not null;index"`
	Description  string      `gorm:"type:text"`
	DepartmentID *string     `gorm:"type:uuid;index"`
	Department   *Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:SET NULL"`
	ProgramID    *string     `gorm:"type:uuid;index"`
	Program      *Program    `gorm:"foreignKey:ProgramID;constraint:OnDelete:SET NULL"`
	StartDate    *time.Time  `gorm:"index"`
	EndDate      *time.Time  `gorm:"index"`
	SupervisorID *string     `gorm:"type:uuid;index"` // Teacher.ID
}
