package models

import "time"

// Токены для почты (подтверждение / сброс пароля)
type EmailToken struct {
	Base
	UserID    string    `gorm:"type:uuid;not null;index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Kind      string    `gorm:"not null;index"` // verify|reset
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
}
