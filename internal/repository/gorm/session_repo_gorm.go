package gormrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

var _ repository.SessionRepository = (*SessionRepositoryGorm)(nil)

// SessionRepositoryGorm persists class sessions.
type SessionRepositoryGorm struct {
	db *gorm.DB
}

// NewSessionRepository builds the repo.
func NewSessionRepository(db *gorm.DB) *SessionRepositoryGorm {
	return &SessionRepositoryGorm{db: db}
}

func (r *SessionRepositoryGorm) Create(ctx context.Context, session *models.ClassSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *SessionRepositoryGorm) GetByID(ctx context.Context, id uuid.UUID) (*models.ClassSession, error) {
	var session models.ClassSession
	if err := r.db.WithContext(ctx).
		Preload("Subject").
		Preload("Group").
		Preload("Teacher").
		First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepositoryGorm) ListBySubjectAndGroup(ctx context.Context, subjectID, groupID uuid.UUID, from *time.Time, to *time.Time) ([]models.ClassSession, error) {
	filter := repository.SessionFilter{SubjectID: &subjectID, GroupID: &groupID, From: from, To: to}
	return r.ListByFilter(ctx, filter)
}

func (r *SessionRepositoryGorm) ListByStudent(ctx context.Context, studentID uuid.UUID) ([]models.ClassSession, error) {
	subQuery := r.db.Table("grades").Select("session_id").Where("student_id = ?", studentID)
	var sessions []models.ClassSession
	if err := r.db.WithContext(ctx).
		Preload("Subject").
		Preload("Group").
		Preload("Teacher").
		Where("id IN (?)", subQuery).
		Order("starts_at ASC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepositoryGorm) ListByTeacher(ctx context.Context, teacherID uuid.UUID, from *time.Time, to *time.Time) ([]models.ClassSession, error) {
	filter := repository.SessionFilter{TeacherID: &teacherID, From: from, To: to}
	return r.ListByFilter(ctx, filter)
}

func (r *SessionRepositoryGorm) ListByGroup(ctx context.Context, groupID uuid.UUID, from *time.Time, to *time.Time) ([]models.ClassSession, error) {
	filter := repository.SessionFilter{GroupID: &groupID, From: from, To: to}
	return r.ListByFilter(ctx, filter)
}

func (r *SessionRepositoryGorm) ListByFilter(ctx context.Context, filter repository.SessionFilter) ([]models.ClassSession, error) {
	q := r.db.WithContext(ctx).
		Preload("Subject").
		Preload("Group").
		Preload("Teacher").
		Order("starts_at ASC")

	if filter.SubjectID != nil {
		q = q.Where("subject_id = ?", *filter.SubjectID)
	}
	if filter.GroupID != nil {
		q = q.Where("group_id = ?", *filter.GroupID)
	}
	if filter.TeacherID != nil {
		q = q.Where("teacher_id = ?", *filter.TeacherID)
	}
	if filter.From != nil {
		q = q.Where("starts_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("starts_at <= ?", *filter.To)
	}

	var sessions []models.ClassSession
	if err := q.Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepositoryGorm) DeleteByTeacher(ctx context.Context, teacherID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Delete(&models.ClassSession{}).Error
}

func (r *SessionRepositoryGorm) DeleteByGroup(ctx context.Context, groupID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Delete(&models.ClassSession{}).Error
}

func (r *SessionRepositoryGorm) DeleteBySubject(ctx context.Context, subjectID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("subject_id = ?", subjectID).
		Delete(&models.ClassSession{}).Error
}
