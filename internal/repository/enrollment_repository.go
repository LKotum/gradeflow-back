package repository

import (
	"fmt"

	req "gradeflow/internal/domain/dto/request"
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type EnrollmentRepository interface {
	Create(enrollment *m.Enrollment) error
	List(query req.ListEnrollmentQuery, limit, offset int) ([]m.Enrollment, int64, error)
	GetByID(id string) (*m.Enrollment, error)
	Update(enrollment *m.Enrollment) error
	Delete(id string) error
}

type enrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) Create(enrollment *m.Enrollment) error {
	if err := r.db.Create(enrollment).Error; err != nil {
		return fmt.Errorf("create enrollment: %w", err)
	}
	return nil
}

func (r *enrollmentRepository) List(query req.ListEnrollmentQuery, limit, offset int) ([]m.Enrollment, int64, error) {
	base := r.db.Model(&m.Enrollment{})
	if v := query.CourseID; v != "" {
		base = base.Where("course_id = ?", v)
	}
	if v := query.StudentID; v != "" {
		base = base.Where("student_id = ?", v)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count enrollments: %w", err)
	}
	listQuery := base
	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}
	var enrollments []m.Enrollment
	if err := listQuery.Find(&enrollments).Error; err != nil {
		return nil, 0, fmt.Errorf("list enrollments: %w", err)
	}
	return enrollments, total, nil
}

func (r *enrollmentRepository) GetByID(id string) (*m.Enrollment, error) {
	var enrollment m.Enrollment
	if err := r.db.First(&enrollment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (r *enrollmentRepository) Update(enrollment *m.Enrollment) error {
	if err := r.db.Save(enrollment).Error; err != nil {
		return fmt.Errorf("update enrollment: %w", err)
	}
	return nil
}

func (r *enrollmentRepository) Delete(id string) error {
	if err := r.db.Delete(&m.Enrollment{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete enrollment: %w", err)
	}
	return nil
}
