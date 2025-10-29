package repository

import (
	"time"

	"github.com/google/uuid"
)

// ListOptions describes common pagination and filtering options.
type ListOptions struct {
	Limit   int
	Offset  int
	Search  *string
	IncludeDeleted bool

	HashKey string
}

// SessionFilter describes optional filters for listing sessions.
type SessionFilter struct {
	SubjectID *uuid.UUID
	GroupID   *uuid.UUID
	TeacherID *uuid.UUID
	From      *time.Time
	To        *time.Time
}
