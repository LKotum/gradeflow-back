package gormrepo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

var _ repository.GradeRepository = (*GradeRepositoryGorm)(nil)

// GradeRepositoryGorm persists grade data.
type GradeRepositoryGorm struct {
	db *gorm.DB
}

// NewGradeRepository constructs repo.
func NewGradeRepository(db *gorm.DB) *GradeRepositoryGorm {
	return &GradeRepositoryGorm{db: db}
}

// Upsert creates or replaces grade for session+student.
func (r *GradeRepositoryGorm) Upsert(ctx context.Context, grade *models.Grade) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "student_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "notes", "assessed_at", "updated_at", "teacher_id"}),
	}).Create(grade).Error
}

func (r *GradeRepositoryGorm) Update(ctx context.Context, grade *models.Grade) error {
	return r.db.WithContext(ctx).Save(grade).Error
}

func (r *GradeRepositoryGorm) Delete(ctx context.Context, gradeID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Grade{}, "id = ?", gradeID).Error
}

func (r *GradeRepositoryGorm) DeleteByStudent(ctx context.Context, studentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Delete(&models.Grade{}).Error
}

func (r *GradeRepositoryGorm) DeleteByTeacher(ctx context.Context, teacherID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Delete(&models.Grade{}).Error
}

func (r *GradeRepositoryGorm) DeleteByGroup(ctx context.Context, groupID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("session_id IN (SELECT id FROM class_sessions WHERE group_id = ?)", groupID).
		Delete(&models.Grade{}).Error
}

func (r *GradeRepositoryGorm) DeleteBySubject(ctx context.Context, subjectID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Delete(&models.Grade{}).Error
}

func (r *GradeRepositoryGorm) GetByID(ctx context.Context, gradeID uuid.UUID) (*models.Grade, error) {
	var grade models.Grade
	if err := r.db.WithContext(ctx).First(&grade, "id = ?", gradeID).Error; err != nil {
		return nil, err
}
	return &grade, nil
}

func (r *GradeRepositoryGorm) GetBySessionAndStudent(ctx context.Context, sessionID, studentID uuid.UUID) (*models.Grade, error) {
	var grade models.Grade
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND student_id = ?", sessionID, studentID).
		First(&grade).Error; err != nil {
		return nil, err
	}
	return &grade, nil
}

func (r *GradeRepositoryGorm) ListBySubjectAndGroup(ctx context.Context, subjectID, groupID uuid.UUID) ([]models.Grade, error) {
	var grades []models.Grade
	if err := r.db.WithContext(ctx).
		Where("subject_id = ? AND session_id IN (SELECT id FROM class_sessions WHERE group_id = ?)", subjectID, groupID).
		Find(&grades).Error; err != nil {
		return nil, err
	}
	return grades, nil
}

func (r *GradeRepositoryGorm) ListByStudent(ctx context.Context, studentID uuid.UUID) ([]models.Grade, error) {
	var grades []models.Grade
	if err := r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("assessed_at ASC").
		Find(&grades).Error; err != nil {
		return nil, err
	}
	return grades, nil
}

func (r *GradeRepositoryGorm) aggregate(ctx context.Context, query string, args ...interface{}) (*float32, error) {
	type result struct {
		Avg *float32
	}
	var res result
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res.Avg, nil
}

func (r *GradeRepositoryGorm) GroupAverage(ctx context.Context, groupID uuid.UUID) (*float32, error) {
	return r.aggregate(ctx, `
		SELECT AVG(value) AS avg
		FROM grades
		WHERE session_id IN (SELECT id FROM class_sessions WHERE group_id = ?)`, groupID)
}

func (r *GradeRepositoryGorm) GroupSubjectAverage(ctx context.Context, groupID, subjectID uuid.UUID) (*float32, error) {
	return r.aggregate(ctx, `
		SELECT AVG(value) AS avg
		FROM grades
		WHERE subject_id = ? AND session_id IN (SELECT id FROM class_sessions WHERE group_id = ?)`, subjectID, groupID)
}

func (r *GradeRepositoryGorm) StudentSubjectAverage(ctx context.Context, studentID, subjectID uuid.UUID) (*float32, error) {
	return r.aggregate(ctx, `
		SELECT AVG(value) AS avg
		FROM grades
		WHERE student_id = ? AND subject_id = ?`, studentID, subjectID)
}

func (r *GradeRepositoryGorm) StudentOverallAverage(ctx context.Context, studentID uuid.UUID) (*float32, error) {
	return r.aggregate(ctx, `
		SELECT AVG(value) AS avg
		FROM grades
		WHERE student_id = ?`, studentID)
}
