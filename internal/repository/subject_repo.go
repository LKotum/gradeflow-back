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
	List(ctx context.Context, opts ListOptions) ([]models.Subject, int64, error)
	ListDeleted(ctx context.Context, opts ListOptions) ([]models.Subject, int64, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID) ([]models.Subject, error)
	Update(ctx context.Context, subject *models.Subject) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error

	AssignTeacher(ctx context.Context, assignment *models.TeachingAssignment) error
	AttachGroup(ctx context.Context, link *models.SubjectGroup) error
	ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]models.TeachingAssignment, error)
	ListSubjectAssignments(ctx context.Context, subjectID uuid.UUID) ([]models.TeachingAssignment, error)
	ListSubjectGroups(ctx context.Context, subjectID uuid.UUID) ([]models.SubjectGroup, error)

	RemoveTeacherAssignments(ctx context.Context, teacherID uuid.UUID) error
	RemoveAssignmentsBySubject(ctx context.Context, subjectID uuid.UUID) error
	RemoveGroupLinks(ctx context.Context, subjectID uuid.UUID) error
	RemoveTeacherAssignment(ctx context.Context, teacherID, subjectID uuid.UUID) error
}
