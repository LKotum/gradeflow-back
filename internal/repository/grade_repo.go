package repository

import (
	"context"

	"github.com/google/uuid"

	"gradeflow/internal/domain/models"
)

// GradeRepository handles grade persistence and queries.
type GradeRepository interface {
	Upsert(ctx context.Context, grade *models.Grade) error
	Update(ctx context.Context, grade *models.Grade) error
	Delete(ctx context.Context, gradeID uuid.UUID) error
	GetByID(ctx context.Context, gradeID uuid.UUID) (*models.Grade, error)
	GetBySessionAndStudent(ctx context.Context, sessionID, studentID uuid.UUID) (*models.Grade, error)
	ListBySubjectAndGroup(ctx context.Context, subjectID, groupID uuid.UUID) ([]models.Grade, error)
	ListByStudent(ctx context.Context, studentID uuid.UUID) ([]models.Grade, error)
	GroupAverage(ctx context.Context, groupID uuid.UUID) (*float32, error)
	GroupSubjectAverage(ctx context.Context, groupID, subjectID uuid.UUID) (*float32, error)
	StudentSubjectAverage(ctx context.Context, studentID, subjectID uuid.UUID) (*float32, error)
	StudentOverallAverage(ctx context.Context, studentID uuid.UUID) (*float32, error)
}
