package service

import (
	"errors"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

const (
	defaultEnrollmentLimit = 200
	maxEnrollmentLimit     = 1000
)

type EnrollmentService interface {
	Create(input req.CreateEnrollment) (*models.Enrollment, error)
	List(query req.ListEnrollmentQuery) ([]models.Enrollment, int64, int, int, error)
	Get(id string) (*models.Enrollment, error)
	Update(id string, input req.UpdateEnrollment) (*models.Enrollment, error)
	Delete(id string) error
}

type enrollmentService struct {
	repo repository.EnrollmentRepository
}

func NewEnrollmentService(repo repository.EnrollmentRepository) EnrollmentService {
	return &enrollmentService{repo: repo}
}

func (s *enrollmentService) Create(input req.CreateEnrollment) (*models.Enrollment, error) {
	if input.CourseID == "" || input.StudentID == "" {
		return nil, errors.New("courseId and studentId are required")
	}
	en := &models.Enrollment{CourseID: input.CourseID, StudentID: input.StudentID}
	if err := s.repo.Create(en); err != nil {
		return nil, err
	}
	return en, nil
}

func (s *enrollmentService) List(query req.ListEnrollmentQuery) ([]models.Enrollment, int64, int, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultEnrollmentLimit
	}
	if limit > maxEnrollmentLimit {
		limit = maxEnrollmentLimit
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	items, total, err := s.repo.List(query, limit, offset)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, limit, offset, nil
}

func (s *enrollmentService) Get(id string) (*models.Enrollment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(id)
}

func (s *enrollmentService) Update(id string, input req.UpdateEnrollment) (*models.Enrollment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	enrollment, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.CourseID != "" {
		enrollment.CourseID = input.CourseID
	}
	if input.StudentID != "" {
		enrollment.StudentID = input.StudentID
	}
	if err := s.repo.Update(enrollment); err != nil {
		return nil, err
	}
	return enrollment, nil
}

func (s *enrollmentService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
