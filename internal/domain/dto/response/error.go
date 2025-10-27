package response

// ErrorResponse represents a consistent error payload.
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]any    `json:"details,omitempty"`
}
