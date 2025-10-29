package request

// AdminLoginRequest captures credentials for admin authentication via INS.
type AdminLoginRequest struct {
	INS      string `json:"ins" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// INSLoginRequest is used by students, teachers, and dean staff.
type INSLoginRequest struct {
	INS      string `json:"ins" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// RefreshTokenRequest requests a new access token from refresh token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ChangePasswordRequest allows a user to change password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,min=8"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}
