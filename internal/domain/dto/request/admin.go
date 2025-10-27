package request

// CreateDeanRequest defines payload to create dean office staff.
type CreateDeanRequest struct {
	INS        string  `json:"ins" binding:"required"`
	Password   string  `json:"password" binding:"required,min=8"`
	Email      *string `json:"email,omitempty"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	Position   *string `json:"position,omitempty"`
}
