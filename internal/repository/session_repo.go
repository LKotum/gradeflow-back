package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"gradeflow/internal/domain/models"
)

// SessionRepository handles class sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *models.ClassSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ClassSession, error)
	ListBySubjectAndGroup(ctx context.Context, subjectID, groupID uuid.UUID, from *time.Time, to *time.Time) ([]models.ClassSession, error)
	ListByStudent(ctx context.Context, studentID uuid.UUID) ([]models.ClassSession, error)
	DeleteByTeacher(ctx context.Context, teacherID uuid.UUID) error
	DeleteByGroup(ctx context.Context, groupID uuid.UUID) error
	DeleteBySubject(ctx context.Context, subjectID uuid.UUID) error
	ListByTeacher(ctx context.Context, teacherID uuid.UUID, from *time.Time, to *time.Time) ([]models.ClassSession, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID, from *time.Time, to *time.Time) ([]models.ClassSession, error)
	ListByFilter(ctx context.Context, filter SessionFilter) ([]models.ClassSession, error)
}
