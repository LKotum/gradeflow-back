package repository

import (
	"fmt"

	"gorm.io/gorm"

	m "gradeflow/internal/domain/models"
)

type StudentRepository interface {
	Create(student *m.Student) error
	List(limit, offset int) ([]m.Student, int64, error)
	GetByID(id string) (*m.Student, error)
	Update(student *m.Student) error
	Delete(id string) error
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) Create(student *m.Student) error {
	if err := r.db.Create(student).Error; err != nil {
		return fmt.Errorf("create student: %w", err)
	}
	return nil
}

func (r *studentRepository) List(limit, offset int) ([]m.Student, int64, error) {
	base := r.db.Model(&m.Student{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count students: %w", err)
	}
	var items []m.Student
	if err := base.Order("full_name asc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list students: %w", err)
	}
	return items, total, nil
}

func (r *studentRepository) GetByID(id string) (*m.Student, error) {
	var student m.Student
	if err := r.db.First(&student, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *studentRepository) Update(student *m.Student) error {
	if err := r.db.Save(student).Error; err != nil {
		return fmt.Errorf("update student: %w", err)
	}
	return nil
}

func (r *studentRepository) Delete(id string) error {
	if err := r.db.Delete(&m.Student{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete student: %w", err)
	}
	return nil
}
