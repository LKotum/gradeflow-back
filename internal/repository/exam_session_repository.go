package repository

import (
	"fmt"

	req "gradeflow/internal/domain/dto/request"
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type ExamSessionRepository interface {
	Create(session *m.ExamSession) error
	List(query req.ListExamSessionQuery, limit, offset int) ([]m.ExamSession, int64, error)
	GetByID(id string) (*m.ExamSession, error)
	Update(session *m.ExamSession) error
	Delete(id string) error
}

type examSessionRepository struct {
	db *gorm.DB
}

func NewExamSessionRepository(db *gorm.DB) ExamSessionRepository {
	return &examSessionRepository{db: db}
}

func (r *examSessionRepository) Create(session *m.ExamSession) error {
	if err := r.db.Create(session).Error; err != nil {
		return fmt.Errorf("create exam session: %w", err)
	}
	return nil
}

func (r *examSessionRepository) List(query req.ListExamSessionQuery, limit, offset int) ([]m.ExamSession, int64, error) {
	base := r.db.Model(&m.ExamSession{})
	if v := query.AcademicSessionID; v != "" {
		base = base.Where("academic_session_id = ?", v)
	}
	if v := query.Q; v != "" {
		base = base.Where("name ILIKE ?", "%"+v+"%")
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count exam sessions: %w", err)
	}
	listQuery := base.Order("starts_at asc nulls last")
	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}
	var items []m.ExamSession
	if err := listQuery.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list exam sessions: %w", err)
	}
	return items, total, nil
}

func (r *examSessionRepository) GetByID(id string) (*m.ExamSession, error) {
	var session m.ExamSession
	if err := r.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *examSessionRepository) Update(session *m.ExamSession) error {
	if err := r.db.Save(session).Error; err != nil {
		return fmt.Errorf("update exam session: %w", err)
	}
	return nil
}

func (r *examSessionRepository) Delete(id string) error {
	if err := r.db.Delete(&m.ExamSession{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete exam session: %w", err)
	}
	return nil
}
