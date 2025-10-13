package service

import (
	"errors"
	"time"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

const (
	defaultLessonLimit = 200
	maxLessonLimit     = 1000
)

type LessonService interface {
	Create(input req.CreateLesson) (*models.Lesson, error)
	List(query req.ListLessonQuery) ([]models.Lesson, int64, int, int, error)
	Get(id string) (*models.Lesson, error)
	Update(id string, input req.UpdateLesson) (*models.Lesson, error)
	Delete(id string) error
}

type lessonService struct {
	repo repository.LessonRepository
}

func NewLessonService(repo repository.LessonRepository) LessonService {
	return &lessonService{repo: repo}
}

func (s *lessonService) Create(input req.CreateLesson) (*models.Lesson, error) {
	lesson := &models.Lesson{
		CourseID: input.CourseID,
		StartsAt: input.StartsAt,
		EndsAt:   input.EndsAt,
		Room:     input.Room,
		Kind:     input.Kind,
	}
	if err := s.repo.Create(lesson); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (s *lessonService) List(query req.ListLessonQuery) ([]models.Lesson, int64, int, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultLessonLimit
	}
	if limit > maxLessonLimit {
		limit = maxLessonLimit
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

func (s *lessonService) Get(id string) (*models.Lesson, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(id)
}

func (s *lessonService) Update(id string, input req.UpdateLesson) (*models.Lesson, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	lesson, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.CourseID != "" {
		lesson.CourseID = input.CourseID
	}
	if input.StartsAt != nil {
		lesson.StartsAt = *input.StartsAt
	}
	if input.EndsAt != nil {
		lesson.EndsAt = *input.EndsAt
	}
	if input.Room != "" {
		lesson.Room = input.Room
	}
	if input.Kind != "" {
		lesson.Kind = input.Kind
	}
	if err := s.repo.Update(lesson); err != nil {
		return nil, err
	}
	return lesson, nil
}

func (s *lessonService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
