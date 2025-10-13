package service

import (
	"errors"
	"time"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

const (
	defaultExamAttemptLimit = 50
	maxExamAttemptLimit     = 200
)

type ExamAttemptService interface {
	Create(input req.CreateExamAttempt) (*models.ExamAttempt, error)
	List(query req.ListExamAttemptQuery) ([]models.ExamAttempt, int64, int, int, error)
	Get(id string) (*models.ExamAttempt, error)
	Update(id string, input req.UpdateExamAttempt) (*models.ExamAttempt, error)
	Delete(id string) error
}

type examAttemptService struct {
	repo repository.ExamAttemptRepository
}

func NewExamAttemptService(repo repository.ExamAttemptRepository) ExamAttemptService {
	return &examAttemptService{repo: repo}
}

func (s *examAttemptService) Create(input req.CreateExamAttempt) (*models.ExamAttempt, error) {
	if input.AssessmentID == "" || input.StudentID == "" || input.ResultScale == "" {
		return nil, errors.New("assessmentId, studentId and resultScale are required")
	}
	if err := validateExamAttemptScale(input.ResultScale, input.ValueNum, input.ValuePass); err != nil {
		return nil, err
	}
	attempt := &models.ExamAttempt{
		AssessmentID: input.AssessmentID,
		StudentID:    input.StudentID,
		AttemptNo:    input.AttemptNo,
		ResultScale:  input.ResultScale,
		ValueNum:     input.ValueNum,
		ValuePass:    input.ValuePass,
		Notes:        input.Notes,
		DateAt:       parseExamAttemptTimePtr(input.DateAt),
	}
	if err := s.repo.Create(attempt); err != nil {
		return nil, err
	}
	return attempt, nil
}

func (s *examAttemptService) List(query req.ListExamAttemptQuery) ([]models.ExamAttempt, int64, int, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultExamAttemptLimit
	}
	if limit > maxExamAttemptLimit {
		limit = maxExamAttemptLimit
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

func (s *examAttemptService) Get(id string) (*models.ExamAttempt, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(id)
}

func (s *examAttemptService) Update(id string, input req.UpdateExamAttempt) (*models.ExamAttempt, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	attempt, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.AttemptNo != nil {
		attempt.AttemptNo = *input.AttemptNo
	}
	if input.DateAt != nil {
		attempt.DateAt = parseExamAttemptTimePtr(input.DateAt)
	}
	if input.ResultScale != nil {
		if err := validateExamAttemptScale(*input.ResultScale, input.ValueNum, input.ValuePass); err != nil {
			return nil, err
		}
		attempt.ResultScale = *input.ResultScale
	} else if input.ValueNum != nil || input.ValuePass != nil {
		if err := validateExamAttemptScale(attempt.ResultScale, input.ValueNum, input.ValuePass); err != nil {
			return nil, err
		}
	}
	if input.ValueNum != nil {
		attempt.ValueNum = input.ValueNum
	}
	if input.ValuePass != nil {
		attempt.ValuePass = input.ValuePass
	}
	if input.Notes != nil {
		attempt.Notes = *input.Notes
	}
	if err := s.repo.Update(attempt); err != nil {
		return nil, err
	}
	return attempt, nil
}

func (s *examAttemptService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func validateExamAttemptScale(scale string, vnum *int, vpass *bool) error {
	switch scale {
	case "passfail":
		if vpass == nil || vnum != nil {
			return errors.New("invalid value for passfail")
		}
	case "five":
		if vnum == nil || vpass != nil {
			return errors.New("invalid value for five")
		}
		if *vnum < 2 || *vnum > 5 {
			return errors.New("value must be 2..5")
		}
	case "hundred":
		if vnum == nil || vpass != nil {
			return errors.New("invalid value for hundred")
		}
		if *vnum < 0 || *vnum > 100 {
			return errors.New("value must be 0..100")
		}
	default:
		return errors.New("invalid scale")
	}
	return nil
}

func parseExamAttemptTimePtr(src *string) *time.Time {
	if src == nil || *src == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, *src); err == nil {
		return &t
	}
	return nil
}
