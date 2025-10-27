package response

import (
	"time"

	"gradeflow/internal/domain/models"
)

// AuthResponse returns access credentials after login or registration.
type AuthResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken,omitempty"`
	User         UserSummary `json:"user"`
	ExpiresAt    time.Time   `json:"expiresAt"`
}

// UserSummary exposes a minimal set of user information.
type UserSummary struct {
	ID         string          `json:"id"`
	Role       models.UserRole `json:"role"`
	INS        *string         `json:"ins,omitempty"`
	Username   *string         `json:"username,omitempty"`
	Email      *string         `json:"email,omitempty"`
	FirstName  string          `json:"firstName"`
	LastName   string          `json:"lastName"`
	MiddleName *string         `json:"middleName,omitempty"`
	AvatarURL  *string         `json:"avatarUrl,omitempty"`
}
