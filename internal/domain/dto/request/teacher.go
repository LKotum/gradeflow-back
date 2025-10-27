package request

// GradeUpsertRequest is used to create or update a grade for a session.
type GradeUpsertRequest struct {
	SessionID string   `json:"sessionId" binding:"required"`
	StudentID string   `json:"studentId" binding:"required"`
	Value     float32  `json:"value" binding:"required"`
	Notes     *string  `json:"notes,omitempty"`
}

// GradeUpdateRequest updates value/notes for an existing grade.
type GradeUpdateRequest struct {
	Value float32  `json:"value" binding:"required"`
	Notes *string  `json:"notes,omitempty"`
}
