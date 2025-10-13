package response

// Error describes an error payload
type Error struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// OK represents a generic success response with ok=true
type OK struct {
	OK bool `json:"ok"`
}

// CreatedID represents a resource creation response with its ID
type CreatedID struct {
	ID string `json:"id"`
}

// Page contains pagination metadata
type Page struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

// List wraps list results with pagination metadata
type List[T any] struct {
	Items []T  `json:"items"`
	Page  Page `json:"page"`
}
