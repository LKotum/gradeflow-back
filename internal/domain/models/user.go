package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole enumerates allowed user roles in the system.
type UserRole string

const (
	// UserRoleAdmin identifies an administrator account.
	UserRoleAdmin UserRole = "admin"
	// UserRoleDean identifies dean office staff.
	UserRoleDean UserRole = "dean"
	// UserRoleTeacher identifies a teacher.
	UserRoleTeacher UserRole = "teacher"
	// UserRoleStudent identifies a student.
	UserRoleStudent UserRole = "student"
)

// User represents any authenticated identity in the system.
type User struct {
	ID           uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Role         UserRole         `gorm:"type:text;not null;index"`
	INS          *string          `gorm:"type:text;uniqueIndex"`
	Username     *string          `gorm:"type:text;uniqueIndex"`
	Email        *string          `gorm:"type:text;uniqueIndex"`
	PasswordHash string           `gorm:"type:text;not null"`
	FirstName    string           `gorm:"type:text;not null"`
	LastName     string           `gorm:"type:text;not null"`
	MiddleName   *string          `gorm:"type:text"`
	AvatarURL    *string          `gorm:"type:text"`
	Student      *StudentProfile  `gorm:"constraint:OnDelete:SET NULL"`
	Teacher      *TeacherProfile  `gorm:"constraint:OnDelete:SET NULL"`
	Staff        *StaffProfile    `gorm:"constraint:OnDelete:SET NULL"`
	RefreshToken *RefreshToken    `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time        `gorm:"not null"`
	UpdatedAt    time.Time        `gorm:"not null"`
}

// StudentProfile stores student-specific data kept separate from base user record.
type StudentProfile struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"`
	GroupID   *uuid.UUID `gorm:"type:uuid;index"`
	Index     string     `gorm:"type:text;uniqueIndex"`
	User      User       `gorm:"constraint:OnDelete:CASCADE"`
	Group     *Group     `gorm:"constraint:OnDelete:SET NULL"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
}

// TeacherProfile augments a user with teacher-specific metadata.
type TeacherProfile struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Title     *string   `gorm:"type:text"`
	Bio       *string   `gorm:"type:text"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// StaffProfile stores dean office staff related information.
type StaffProfile struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Position  *string   `gorm:"type:text"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// RefreshToken allows issuing long-lived refresh tokens for sessions.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Token     string    `gorm:"type:text;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
