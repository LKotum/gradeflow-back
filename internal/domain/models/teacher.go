package models

// Преподаватель (профиль преподавателя, привязанный к пользователю с ролью teacher)
type Teacher struct {
	Base
	FullName     string      `gorm:"not null;index"`
	DepartmentID *string     `gorm:"type:uuid;index"`
	Department   *Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:SET NULL"`
	Title        string      `gorm:"index"` // должность, например: доцент
	Rank         string      `gorm:"index"` // ученое звание
	UserID       *string     `gorm:"type:uuid;uniqueIndex"`
	User         *User       `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
