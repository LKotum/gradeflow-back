package repository

import (
	"fmt"
	"time"

	req "gradeflow/internal/domain/dto/request"
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type LessonRepository interface {
	Create(lesson *m.Lesson) error
	List(query req.ListLessonQuery, limit, offset int, from, to *time.Time) ([]m.Lesson, int64, error)
	GetByID(id string) (*m.Lesson, error)
	Update(lesson *m.Lesson) error
	Delete(id string) error
}

type lessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) LessonRepository {
	return &lessonRepository{db: db}
}

func (r *lessonRepository) Create(lesson *m.Lesson) error {
	if err := r.db.Create(lesson).Error; err != nil {
		return fmt.Errorf("create lesson: %w", err)
	}
	return nil
}

func (r *lessonRepository) List(query req.ListLessonQuery, limit, offset int, from, to *time.Time) ([]m.Lesson, int64, error) {
	base := r.db.Model(&m.Lesson{})
	base = applyLessonQuery(base, query, from, to)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count lessons: %w", err)
	}

	listQuery := applyLessonQuery(r.db.Model(&m.Lesson{}), query, from, to).
		Order("starts_at asc")
	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}

	var lessons []m.Lesson
	if err := listQuery.Find(&lessons).Error; err != nil {
		return nil, 0, fmt.Errorf("list lessons: %w", err)
	}
	return lessons, total, nil
}

func (r *lessonRepository) GetByID(id string) (*m.Lesson, error) {
	var lesson m.Lesson
	if err := r.db.First(&lesson, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *lessonRepository) Update(lesson *m.Lesson) error {
	if err := r.db.Save(lesson).Error; err != nil {
		return fmt.Errorf("update lesson: %w", err)
	}
	return nil
}

func (r *lessonRepository) Delete(id string) error {
	if err := r.db.Delete(&m.Lesson{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	return nil
}

func applyLessonQuery(db *gorm.DB, query req.ListLessonQuery, from, to *time.Time) *gorm.DB {
	if v := query.CourseID; v != "" {
		db = db.Where("course_id = ?", v)
	}
	if v := query.SessionID; v != "" {
		db = db.Joins("JOIN courses ON courses.id = lessons.course_id").Where("courses.academic_session_id = ?", v)
	}
	if from != nil {
		db = db.Where("ends_at >= ?", *from)
	}
	if to != nil {
		db = db.Where("starts_at <= ?", *to)
	}
	return db
}
