package repository

import (
	"fmt"

	req "gradeflow/internal/domain/dto/request"
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type CourseRepository interface {
	Create(course *m.Course) error
	List(query req.ListCourseQuery, limit, offset int) ([]m.Course, int64, error)
	GetByID(id string) (*m.Course, error)
	Update(course *m.Course) error
	Delete(id string) error
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) Create(course *m.Course) error {
	if err := r.db.Create(course).Error; err != nil {
		return fmt.Errorf("create course: %w", err)
	}
	return nil
}

func (r *courseRepository) List(query req.ListCourseQuery, limit, offset int) ([]m.Course, int64, error) {
	base := r.db.Model(&m.Course{})
	if v := query.DepartmentID; v != "" {
		base = base.Where("department_id = ?", v)
	}
	if v := query.ProgramID; v != "" {
		base = base.Where("program_id = ?", v)
	}
	if v := query.SubjectID; v != "" {
		base = base.Where("subject_id = ?", v)
	}
	if v := query.AcademicSessionID; v != "" {
		base = base.Where("academic_session_id = ?", v)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count courses: %w", err)
	}
	listQuery := base.Order("title asc")
	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}
	var courses []m.Course
	if err := listQuery.Find(&courses).Error; err != nil {
		return nil, 0, fmt.Errorf("list courses: %w", err)
	}
	return courses, total, nil
}

func (r *courseRepository) GetByID(id string) (*m.Course, error) {
	var course m.Course
	if err := r.db.First(&course, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) Update(course *m.Course) error {
	if err := r.db.Save(course).Error; err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	return nil
}

func (r *courseRepository) Delete(id string) error {
	if err := r.db.Delete(&m.Course{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	return nil
}
