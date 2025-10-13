package service

import (
	"errors"
	"fmt"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

const (
	defaultStudentListLimit = 200
	maxStudentListLimit     = 1000
)

type StudentService interface {
	Create(input req.CreateStudent) (*models.Student, error)
	List(limit, offset int) ([]models.Student, int64, int, int, error)
	Get(id string) (*models.Student, error)
	Update(id string, input req.UpdateStudent) (*models.Student, error)
	Delete(id string) error
}

type studentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) StudentService {
	return &studentService{repo: repo}
}

func (s *studentService) Create(input req.CreateStudent) (*models.Student, error) {
	if input.IndividualNumber == "" {
		return nil, errors.New("individual number is required")
	}
	if input.FullName == "" {
		return nil, errors.New("full name is required")
}
	st := &models.Student{
		IndividualNumber: input.IndividualNumber,
		FullName:         input.FullName,
		GroupID:          copyStringPtr(input.GroupID),
		StartYear:        copyIntPtr(input.StartYear),
		EndYear:          copyIntPtr(input.EndYear),
	}
	if err := s.repo.Create(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *studentService) List(limit, offset int) ([]models.Student, int64, int, int, error) {
	if limit <= 0 {
		limit = defaultStudentListLimit
	}
	if limit > maxStudentListLimit {
		limit = maxStudentListLimit
	}
	if offset < 0 {
		offset = 0
	}
	students, total, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return students, total, limit, offset, nil
}

func (s *studentService) Get(id string) (*models.Student, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	st, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func (s *studentService) Update(id string, input req.UpdateStudent) (*models.Student, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	st, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.IndividualNumber != nil && *input.IndividualNumber != "" {
		st.IndividualNumber = *input.IndividualNumber
	}
	if input.FullName != nil && *input.FullName != "" {
		st.FullName = *input.FullName
	}
	if input.GroupID != nil {
		if *input.GroupID == "" {
			st.GroupID = nil
		} else {
			st.GroupID = copyStringPtr(input.GroupID)
		}
	}
	if input.StartYear != nil {
		st.StartYear = copyIntPtr(input.StartYear)
	}
	if input.EndYear != nil {
		st.EndYear = copyIntPtr(input.EndYear)
	}
	if err := s.repo.Update(st); err != nil {
		return nil, fmt.Errorf("update student: %w", err)
	}
	return st, nil
}

func (s *studentService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func copyStringPtr(src *string) *string {
	if src == nil {
		return nil
	}
	val := *src
	return &val
}

func copyIntPtr(src *int) *int {
	if src == nil {
		return nil
	}
	val := *src
	return &val
}
