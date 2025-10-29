package gormrepo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

var _ repository.SubjectRepository = (*SubjectRepositoryGorm)(nil)

// SubjectRepositoryGorm persists subjects and related relations.
type SubjectRepositoryGorm struct {
	db *gorm.DB
}

// NewSubjectRepository creates the repository.
func NewSubjectRepository(db *gorm.DB) *SubjectRepositoryGorm {
	return &SubjectRepositoryGorm{db: db}
}

// Create inserts subject.
func (r *SubjectRepositoryGorm) Create(ctx context.Context, subject *models.Subject) error {
	return r.db.WithContext(ctx).Create(subject).Error
}

// GetByID loads subject by identifier.
func (r *SubjectRepositoryGorm) GetByID(ctx context.Context, id uuid.UUID) (*models.Subject, error) {
	var subject models.Subject
	if err := r.db.WithContext(ctx).
		Preload("Assignments").
		Preload("GroupLinks").
		First(&subject, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &subject, nil
}

// List returns all subjects.
func (r *SubjectRepositoryGorm) List(ctx context.Context, opts repository.ListOptions) ([]models.Subject, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Subject{})
	base = applySubjectFilters(base, opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query := r.db.WithContext(ctx).Model(&models.Subject{})
	query = applySubjectFilters(query, opts)
	query = query.Order("name ASC")
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	var subjects []models.Subject
	if err := query.Find(&subjects).Error; err != nil {
		return nil, 0, err
	}
	return subjects, total, nil
}

func (r *SubjectRepositoryGorm) ListDeleted(ctx context.Context, opts repository.ListOptions) ([]models.Subject, int64, error) {
	opts.IncludeDeleted = true
	base := r.db.WithContext(ctx).Model(&models.Subject{}).Unscoped().Where("deleted_at IS NOT NULL")
	base = applySubjectFilters(base, opts)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query := r.db.WithContext(ctx).Model(&models.Subject{}).Unscoped().Where("deleted_at IS NOT NULL")
	query = applySubjectFilters(query, opts)
	query = query.Order("name ASC")
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	var subjects []models.Subject
	if err := query.Find(&subjects).Error; err != nil {
		return nil, 0, err
	}
	return subjects, total, nil
}

// ListByGroup returns subjects attached to a group via subject_groups.
func (r *SubjectRepositoryGorm) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]models.Subject, error) {
	var subjects []models.Subject
	if err := r.db.WithContext(ctx).
		Joins("JOIN subject_groups sg ON sg.subject_id = subjects.id").
		Where("sg.group_id = ?", groupID).
		Group("subjects.id").
		Find(&subjects).Error; err != nil {
		return nil, err
	}
	return subjects, nil
}

func (r *SubjectRepositoryGorm) Update(ctx context.Context, subject *models.Subject) error {
	return r.db.WithContext(ctx).Save(subject).Error
}

func (r *SubjectRepositoryGorm) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.RemoveAssignmentsBySubject(ctx, id); err != nil {
		return err
	}
	if err := r.RemoveGroupLinks(ctx, id); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&models.Subject{}, "id = ?", id).Error
}

func (r *SubjectRepositoryGorm) Restore(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Model(&models.Subject{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// AssignTeacher associates teacher to subject.
func (r *SubjectRepositoryGorm) AssignTeacher(ctx context.Context, assignment *models.TeachingAssignment) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "teacher_id"}, {Name: "subject_id"}},
		DoNothing: true,
	}).Create(assignment).Error
}

// AttachGroup links group to subject.
func (r *SubjectRepositoryGorm) AttachGroup(ctx context.Context, link *models.SubjectGroup) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "subject_id"}, {Name: "group_id"}},
		DoNothing: true,
	}).Create(link).Error
}

// ListTeacherAssignments returns assignments for teacher.
func (r *SubjectRepositoryGorm) ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]models.TeachingAssignment, error) {
	var assignments []models.TeachingAssignment
	if err := r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *SubjectRepositoryGorm) ListSubjectAssignments(ctx context.Context, subjectID uuid.UUID) ([]models.TeachingAssignment, error) {
	var assignments []models.TeachingAssignment
	if err := r.db.WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

// ListSubjectGroups returns linked groups for subject.
func (r *SubjectRepositoryGorm) ListSubjectGroups(ctx context.Context, subjectID uuid.UUID) ([]models.SubjectGroup, error) {
	var links []models.SubjectGroup
	if err := r.db.WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (r *SubjectRepositoryGorm) RemoveTeacherAssignments(ctx context.Context, teacherID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Delete(&models.TeachingAssignment{}).Error
}

func (r *SubjectRepositoryGorm) RemoveAssignmentsBySubject(ctx context.Context, subjectID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Delete(&models.TeachingAssignment{}).Error
}

func (r *SubjectRepositoryGorm) RemoveGroupLinks(ctx context.Context, subjectID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Delete(&models.SubjectGroup{}).Error
}

func (r *SubjectRepositoryGorm) RemoveTeacherAssignment(ctx context.Context, teacherID, subjectID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("teacher_id = ? AND subject_id = ?", teacherID, subjectID).
		Delete(&models.TeachingAssignment{}).Error
}

func applySubjectFilters(tx *gorm.DB, opts repository.ListOptions) *gorm.DB {
	if opts.IncludeDeleted {
		tx = tx.Unscoped()
	}
	if opts.Search != nil && strings.TrimSpace(*opts.Search) != "" {
		pattern := "%" + strings.TrimSpace(*opts.Search) + "%"
		tx = tx.Where("name ILIKE ? OR code ILIKE ? OR COALESCE(description, '') ILIKE ?", pattern, pattern, pattern)
	}
	return tx
}
