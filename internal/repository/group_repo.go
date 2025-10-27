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
	List(ctx context.Context) ([]models.Group, error)
}
