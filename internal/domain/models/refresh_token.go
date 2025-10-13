package models

import "time"

// Refresh токены для сессий
type RefreshToken struct {
	Base
	UserID    string    `gorm:"type:uuid;not null;index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	Revoked   bool      `gorm:"not null;default:false;index"`
}
