package gormrepo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

var _ repository.GroupRepository = (*GroupRepositoryGorm)(nil)

// GroupRepositoryGorm persists groups.
type GroupRepositoryGorm struct {
	db *gorm.DB
}

// NewGroupRepository builds a repository.
func NewGroupRepository(db *gorm.DB) *GroupRepositoryGorm {
	return &GroupRepositoryGorm{db: db}
}

// Create stores a group.
func (r *GroupRepositoryGorm) Create(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetByID loads group by identifier.
func (r *GroupRepositoryGorm) GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error) {
	var group models.Group
	if err := r.db.WithContext(ctx).
		Preload("Students").
		Preload("Students.User").
		First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// List returns all groups.
func (r *GroupRepositoryGorm) List(ctx context.Context) ([]models.Group, error) {
	var groups []models.Group
	if err := r.db.WithContext(ctx).Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}
