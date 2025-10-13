package models

// Сотрудник деканата
type StaffMember struct {
	Base
	FullName     string      `gorm:"not null;index"`
	DepartmentID *string     `gorm:"type:uuid;index"`
	Department   *Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:SET NULL"`
	Position     string      `gorm:"index"`
	UserID       *string     `gorm:"type:uuid;uniqueIndex"`
	User         *User       `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
