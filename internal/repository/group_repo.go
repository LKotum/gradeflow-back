package repository

import (
	"context"

	"github.com/google/uuid"

	"gradeflow/internal/domain/models"
)

// GroupRepository abstracts group persistence.
type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error)
	List(ctx context.Context, opts ListOptions) ([]models.Group, int64, error)
	ListDeleted(ctx context.Context, opts ListOptions) ([]models.Group, int64, error)
	Update(ctx context.Context, group *models.Group) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error

	DetachStudents(ctx context.Context, groupID uuid.UUID) error
	RemoveSubjectLinks(ctx context.Context, groupID uuid.UUID) error
}
