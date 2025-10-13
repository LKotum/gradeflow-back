package service

import (
	"errors"
	"time"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	validation "gradeflow/internal/domain/validation"
	"gradeflow/internal/repository"

	"gorm.io/gorm"
)

const (
	defaultAssessmentGradeLimit = 500
	maxAssessmentGradeLimit     = 2000
)

type AssessmentGradeService interface {
	Create(input req.CreateAssessmentGrade, gradedBy *string) (*models.AssessmentGrade, error)
	List(query req.ListAssessmentGradeQuery) ([]models.AssessmentGrade, int64, int, int, error)
	ListByAssessment(assessmentID string, limit, offset int) ([]models.AssessmentGrade, int64, int, int, error)
	Get(id string) (*models.AssessmentGrade, error)
	Update(id string, input req.UpdateAssessmentGrade, gradedBy *string) (*models.AssessmentGrade, error)
	Delete(id string) error
	BulkUpsert(assessmentID string, items []req.GradeItem, gradedBy *string) ([]models.AssessmentGrade, error)
}

type assessmentGradeService struct {
	grades      repository.AssessmentGradeRepository
	assessments repository.AssessmentRepository
}

func NewAssessmentGradeService(gr repository.AssessmentGradeRepository, ar repository.AssessmentRepository) AssessmentGradeService {
	return &assessmentGradeService{grades: gr, assessments: ar}
}

func (s *assessmentGradeService) Create(input req.CreateAssessmentGrade, gradedBy *string) (*models.AssessmentGrade, error) {
	if input.AssessmentID == "" || input.StudentID == "" {
		return nil, errors.New("assessmentId and studentId are required")
	}
	assessment, err := s.assessments.GetByID(input.AssessmentID)
	if err != nil {
		return nil, err
	}
	scale := input.Scale
	if scale == "" {
		scale = assessment.Scale
	}
	if scale == "" {
		return nil, errors.New("scale is required")
	}
	if err := validation.ValidateAssessmentTypeScale(assessment.Type, scale); err != nil {
		return nil, err
	}
	if err := validation.ValidateGradeValues(scale, input.ValueNum, input.ValuePass); err != nil {
		return nil, err
	}
	grade := &models.AssessmentGrade{
		AssessmentID: input.AssessmentID,
		StudentID:    input.StudentID,
		Scale:        scale,
		ValueNum:     input.ValueNum,
		ValuePass:    input.ValuePass,
		GradedAt:     time.Now(),
	}
	if gradedBy != nil {
		grade.GradedBy = *gradedBy
	}
	if err := s.grades.Create(grade); err != nil {
		return nil, err
	}
	return grade, nil
}

func (s *assessmentGradeService) List(query req.ListAssessmentGradeQuery) ([]models.AssessmentGrade, int64, int, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultAssessmentGradeLimit
	}
	if limit > maxAssessmentGradeLimit {
		limit = maxAssessmentGradeLimit
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
	items, total, err := s.grades.List(query, limit, offset, fromPtr, toPtr)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, limit, offset, nil
}

func (s *assessmentGradeService) ListByAssessment(assessmentID string, limit, offset int) ([]models.AssessmentGrade, int64, int, int, error) {
	if assessmentID == "" {
		return nil, 0, 0, 0, errors.New("assessmentId is required")
	}
	if limit <= 0 {
		limit = defaultAssessmentGradeLimit
	}
	if limit > maxAssessmentGradeLimit {
		limit = maxAssessmentGradeLimit
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := s.grades.ListByAssessment(assessmentID, limit, offset)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, limit, offset, nil
}

func (s *assessmentGradeService) Get(id string) (*models.AssessmentGrade, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.grades.GetByID(id)
}

func (s *assessmentGradeService) Update(id string, input req.UpdateAssessmentGrade, gradedBy *string) (*models.AssessmentGrade, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	grade, err := s.grades.GetByID(id)
	if err != nil {
		return nil, err
	}
	assessment, err := s.assessments.GetByID(grade.AssessmentID)
	if err != nil {
		return nil, err
	}
	finalScale := grade.Scale
	if input.Scale != "" {
		finalScale = input.Scale
	}
	if finalScale == "" {
		finalScale = assessment.Scale
	}
	if finalScale == "" {
		return nil, errors.New("scale is required")
	}
	if err := validation.ValidateAssessmentTypeScale(assessment.Type, finalScale); err != nil {
		return nil, err
	}
	finalValueNum := grade.ValueNum
	if input.ValueNum != nil {
		finalValueNum = input.ValueNum
	}
	finalValuePass := grade.ValuePass
	if input.ValuePass != nil {
		finalValuePass = input.ValuePass
	}
	if err := validation.ValidateGradeValues(finalScale, finalValueNum, finalValuePass); err != nil {
		return nil, err
	}
	grade.Scale = finalScale
	grade.ValueNum = finalValueNum
	grade.ValuePass = finalValuePass
	if gradedBy != nil {
		grade.GradedBy = *gradedBy
	}
	grade.GradedAt = time.Now()
	if err := s.grades.Update(grade); err != nil {
		return nil, err
	}
	return grade, nil
}

func (s *assessmentGradeService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.grades.GetByID(id); err != nil {
		return err
	}
	return s.grades.Delete(id)
}

func (s *assessmentGradeService) BulkUpsert(assessmentID string, items []req.GradeItem, gradedBy *string) ([]models.AssessmentGrade, error) {
	if assessmentID == "" {
		return nil, errors.New("assessmentId is required")
	}
	if len(items) == 0 {
		return []models.AssessmentGrade{}, nil
	}
	assessment, err := s.assessments.GetByID(assessmentID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, item := range items {
		if item.StudentID == "" {
			return nil, errors.New("studentId is required")
		}
		existing, err := s.grades.GetByAssessmentAndStudent(assessmentID, item.StudentID)
		scale := ""
		if item.Scale != "" {
			scale = item.Scale
		} else if err == nil {
			scale = existing.Scale
		} else {
			scale = assessment.Scale
		}
		if scale == "" {
			return nil, errors.New("scale is required")
		}
		if err := validation.ValidateAssessmentTypeScale(assessment.Type, scale); err != nil {
			return nil, err
		}
		var valueNum *float64
		var valuePass *bool
		if item.ValueNum != nil {
			valueNum = item.ValueNum
		} else if err == nil {
			valueNum = existing.ValueNum
		}
		if item.ValuePass != nil {
			valuePass = item.ValuePass
		} else if err == nil {
			valuePass = existing.ValuePass
		}
		if err := validation.ValidateGradeValues(scale, valueNum, valuePass); err != nil {
			return nil, err
		}
		if err == nil {
			existing.Scale = scale
			existing.ValueNum = valueNum
			existing.ValuePass = valuePass
			if gradedBy != nil {
				existing.GradedBy = *gradedBy
			}
			existing.GradedAt = now
			if err := s.grades.Update(existing); err != nil {
				return nil, err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			grade := &models.AssessmentGrade{
				AssessmentID: assessmentID,
				StudentID:    item.StudentID,
				Scale:        scale,
				ValueNum:     valueNum,
				ValuePass:    valuePass,
				GradedAt:     now,
			}
			if gradedBy != nil {
				grade.GradedBy = *gradedBy
			}
			if err := s.grades.Create(grade); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	grades, _, err := s.grades.ListByAssessment(assessmentID, 0, 0)
	if err != nil {
		return nil, err
	}
	return grades, nil
}
