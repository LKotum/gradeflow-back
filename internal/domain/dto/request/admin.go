package request

// CreateDeanRequest defines payload to create dean office staff.
type CreateDeanRequest struct {
	Password   string  `json:"password" binding:"required,min=8"`
	Email      *string `json:"email,omitempty"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	Position   *string `json:"position,omitempty"`
}

// ResetPasswordRequest is used by admin to set a new password for any user.
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateDeanRequest updates dean profile information.
type UpdateDeanRequest struct {
	Email      *string `json:"email,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	MiddleName *string `json:"middleName,omitempty"`
	Position   *string `json:"position,omitempty"`
}
