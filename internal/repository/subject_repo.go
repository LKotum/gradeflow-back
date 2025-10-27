package repository

import (
	"context"

	"github.com/google/uuid"

	"gradeflow/internal/domain/models"
)

// SubjectRepository handles subject persistence.
type SubjectRepository interface {
	Create(ctx context.Context, subject *models.Subject) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subject, error)
	List(ctx context.Context) ([]models.Subject, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID) ([]models.Subject, error)

	AssignTeacher(ctx context.Context, assignment *models.TeachingAssignment) error
	AttachGroup(ctx context.Context, link *models.SubjectGroup) error
	ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]models.TeachingAssignment, error)
	ListSubjectGroups(ctx context.Context, subjectID uuid.UUID) ([]models.SubjectGroup, error)
}
