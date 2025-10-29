package gormrepo

import (
	"context"
	"strings"

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

func (r *GroupRepositoryGorm) List(ctx context.Context, opts repository.ListOptions) ([]models.Group, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Group{})
	base = applyGroupFilters(base, opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query := r.db.WithContext(ctx).Model(&models.Group{})
	query = applyGroupFilters(query, opts)
	query = query.Order("name ASC")
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	var groups []models.Group
	if err := query.Find(&groups).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *GroupRepositoryGorm) ListDeleted(ctx context.Context, opts repository.ListOptions) ([]models.Group, int64, error) {
	opts.IncludeDeleted = true
	base := r.db.WithContext(ctx).Model(&models.Group{}).Unscoped().Where("deleted_at IS NOT NULL")
	base = applyGroupFilters(base, opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query := r.db.WithContext(ctx).Model(&models.Group{}).Unscoped().Where("deleted_at IS NOT NULL")
	query = applyGroupFilters(query, opts)
	query = query.Order("name ASC")
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	var groups []models.Group
	if err := query.Find(&groups).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *GroupRepositoryGorm) Update(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *GroupRepositoryGorm) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.DetachStudents(ctx, id); err != nil {
		return err
	}
	if err := r.RemoveSubjectLinks(ctx, id); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&models.Group{}, "id = ?", id).Error
}

func (r *GroupRepositoryGorm) Restore(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Model(&models.Group{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *GroupRepositoryGorm) DetachStudents(ctx context.Context, groupID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.StudentProfile{}).
		Where("group_id = ?", groupID).
		Update("group_id", nil).Error
}

func (r *GroupRepositoryGorm) RemoveSubjectLinks(ctx context.Context, groupID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Delete(&models.SubjectGroup{}).Error
}

func applyGroupFilters(tx *gorm.DB, opts repository.ListOptions) *gorm.DB {
	if opts.Search != nil && strings.TrimSpace(*opts.Search) != "" {
		pattern := "%" + strings.TrimSpace(*opts.Search) + "%"
		if opts.IncludeDeleted {
			tx = tx.Unscoped()
		}
		tx = tx.Where("name ILIKE ? OR COALESCE(description, '') ILIKE ?", pattern, pattern)
	} else if opts.IncludeDeleted {
		tx = tx.Unscoped()
	}
	return tx
}
