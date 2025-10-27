package gormrepo

import (
	"context"

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
func (r *SubjectRepositoryGorm) List(ctx context.Context) ([]models.Subject, error) {
	var subjects []models.Subject
	if err := r.db.WithContext(ctx).Find(&subjects).Error; err != nil {
		return nil, err
	}
	return subjects, nil
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
