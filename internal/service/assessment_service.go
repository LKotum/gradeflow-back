package service

import (
	"errors"
	"time"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	validation "gradeflow/internal/domain/validation"
	"gradeflow/internal/repository"
)

const (
	defaultAssessmentLimit = 200
	maxAssessmentLimit     = 1000
)

type AssessmentService interface {
	Create(input req.CreateAssessment) (*models.Assessment, error)
	List(query req.ListAssessmentQuery) ([]models.Assessment, int64, int, int, error)
	Get(id string) (*models.Assessment, error)
	Update(id string, input req.UpdateAssessment) (*models.Assessment, error)
	Delete(id string) error
}

type assessmentService struct {
	repo repository.AssessmentRepository
}

func NewAssessmentService(repo repository.AssessmentRepository) AssessmentService {
	return &assessmentService{repo: repo}
}

func (s *assessmentService) Create(input req.CreateAssessment) (*models.Assessment, error) {
	if err := validation.ValidateAssessmentTypeScale(input.Type, input.Scale); err != nil {
		return nil, err
	}
	a := &models.Assessment{
		CourseID: input.CourseID,
		Type:     input.Type,
		DateAt:   input.DateAt,
		Room:     input.Room,
		Scale:    input.Scale,
		MaxPts:   input.MaxPts,
	}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *assessmentService) List(query req.ListAssessmentQuery) ([]models.Assessment, int64, int, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultAssessmentLimit
	}
	if limit > maxAssessmentLimit {
		limit = maxAssessmentLimit
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	var fromPtr, toPtr *time.Time
	if query.From != "" {
		if t, err := time.Parse(time.RFC3339, query.From); err == nil {
			fromPtr = &t
		}
	}
	if query.To != "" {
		if t, err := time.Parse(time.RFC3339, query.To); err == nil {
			toPtr = &t
		}
	}
	items, total, err := s.repo.List(query, limit, offset, fromPtr, toPtr)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, limit, offset, nil
}

func (s *assessmentService) Get(id string) (*models.Assessment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(id)
}

func (s *assessmentService) Update(id string, input req.UpdateAssessment) (*models.Assessment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	a, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.CourseID != "" {
		a.CourseID = input.CourseID
	}
	if input.Type != "" {
		a.Type = input.Type
	}
	if input.DateAt != nil {
		a.DateAt = *input.DateAt
	}
	if input.Room != "" {
		a.Room = input.Room
	}
	if input.Scale != "" {
		if err := validation.ValidateAssessmentTypeScale(a.Type, input.Scale); err != nil {
			return nil, err
		}
		a.Scale = input.Scale
	} else {
		if err := validation.ValidateAssessmentTypeScale(a.Type, a.Scale); err != nil {
			return nil, err
		}
	}
	if input.MaxPts != nil {
		a.MaxPts = input.MaxPts
	}
	if err := s.repo.Update(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *assessmentService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
