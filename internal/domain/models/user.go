package models

import "time"

// Пользователь и аутентификация
type User struct {
	Base
	Email        string     `gorm:"uniqueIndex;not null"`
	FullName     string     `gorm:"not null;index"`
	PasswordHash string     `gorm:"not null"`
	Role         string     `gorm:"not null;index"` // student|teacher|dean|admin
	Status       string     `gorm:"not null;index"` // pending|active|suspended
	TOTPSecret   *string    `gorm:"type:text"`
	TOTPEnabled  bool       `gorm:"not null;default:false"`
	LastLoginAt  *time.Time `gorm:"index"`
}
