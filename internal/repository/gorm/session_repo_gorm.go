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
	q := r.db.WithContext(ctx).
		Where("subject_id = ? AND group_id = ?", subjectID, groupID).
		Order("starts_at ASC")
	if from != nil {
		q = q.Where("starts_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("starts_at <= ?", *to)
	}
	var sessions []models.ClassSession
	if err := q.Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepositoryGorm) ListByStudent(ctx context.Context, studentID uuid.UUID) ([]models.ClassSession, error) {
	subQuery := r.db.Table("grades").Select("session_id").Where("student_id = ?", studentID)
	var sessions []models.ClassSession
	if err := r.db.WithContext(ctx).
		Where("id IN (?)", subQuery).
		Order("starts_at ASC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}
