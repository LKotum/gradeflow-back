package repository

import (
	"fmt"
	"time"

	req "gradeflow/internal/domain/dto/request"
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type AssessmentRepository interface {
	Create(a *m.Assessment) error
	List(query req.ListAssessmentQuery, limit, offset int, from, to *time.Time) ([]m.Assessment, int64, error)
	GetByID(id string) (*m.Assessment, error)
	Update(a *m.Assessment) error
	Delete(id string) error
}

type assessmentRepository struct {
	db *gorm.DB
}

func NewAssessmentRepository(db *gorm.DB) AssessmentRepository {
	return &assessmentRepository{db: db}
}

func (r *assessmentRepository) Create(a *m.Assessment) error {
	if err := r.db.Create(a).Error; err != nil {
		return fmt.Errorf("create assessment: %w", err)
	}
	return nil
}

func (r *assessmentRepository) List(query req.ListAssessmentQuery, limit, offset int, from, to *time.Time) ([]m.Assessment, int64, error) {
	base := applyAssessmentFilters(r.db.Model(&m.Assessment{}), query, from, to)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count assessments: %w", err)
	}
	listQuery := applyAssessmentFilters(r.db.Model(&m.Assessment{}), query, from, to).Order("date_at asc")
	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}
	var items []m.Assessment
	if err := listQuery.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list assessments: %w", err)
	}
	return items, total, nil
}

func (r *assessmentRepository) GetByID(id string) (*m.Assessment, error) {
	var a m.Assessment
	if err := r.db.First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *assessmentRepository) Update(a *m.Assessment) error {
	if err := r.db.Save(a).Error; err != nil {
		return fmt.Errorf("update assessment: %w", err)
	}
	return nil
}

func (r *assessmentRepository) Delete(id string) error {
	if err := r.db.Delete(&m.Assessment{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete assessment: %w", err)
	}
	return nil
}

func applyAssessmentFilters(db *gorm.DB, query req.ListAssessmentQuery, from, to *time.Time) *gorm.DB {
	if v := query.CourseID; v != "" {
		db = db.Where("course_id = ?", v)
	}
	if from != nil {
		db = db.Where("date_at >= ?", *from)
	}
	if to != nil {
		db = db.Where("date_at <= ?", *to)
	}
	return db
}
