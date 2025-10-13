package models

// Администратор системы
type Admin struct {
	Base
	FullName string  `gorm:"not null;index"`
	UserID   *string `gorm:"type:uuid;uniqueIndex"`
	User     *User   `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
