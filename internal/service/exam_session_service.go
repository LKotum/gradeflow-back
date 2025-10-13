package service

import (
    "errors"
    "time"

    req "gradeflow/internal/domain/dto/request"
    "gradeflow/internal/domain/models"
    "gradeflow/internal/repository"
)

const (
    defaultExamSessionLimit = 50
    maxExamSessionLimit     = 200
)

type ExamSessionService interface {
    Create(input req.CreateExamSession) (*models.ExamSession, error)
    List(query req.ListExamSessionQuery) ([]models.ExamSession, int64, int, int, error)
    Get(id string) (*models.ExamSession, error)
    Update(id string, input req.UpdateExamSession) (*models.ExamSession, error)
    Delete(id string) error
}

type examSessionService struct {
    repo repository.ExamSessionRepository
}

func NewExamSessionService(repo repository.ExamSessionRepository) ExamSessionService {
    return &examSessionService{repo: repo}
}

func (s *examSessionService) Create(input req.CreateExamSession) (*models.ExamSession, error) {
    if input.AcademicSessionID == "" || input.Name == "" {
        return nil, errors.New("academicSessionId and name are required")
    }
    session := &models.ExamSession{AcademicSessionID: input.AcademicSessionID, Name: input.Name}
    session.StartsAt = parseTimePtr(input.StartsAt)
    session.EndsAt = parseTimePtr(input.EndsAt)
    if err := s.repo.Create(session); err != nil {
        return nil, err
    }
    return session, nil
}

func (s *examSessionService) List(query req.ListExamSessionQuery) ([]models.ExamSession, int64, int, int, error) {
    limit := query.Limit
    if limit <= 0 {
        limit = defaultExamSessionLimit
    }
    if limit > maxExamSessionLimit {
        limit = maxExamSessionLimit
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

func (s *examSessionService) Get(id string) (*models.ExamSession, error) {
    if id == "" {
        return nil, errors.New("id is required")
    }
    return s.repo.GetByID(id)
}

func (s *examSessionService) Update(id string, input req.UpdateExamSession) (*models.ExamSession, error) {
    if id == "" {
        return nil, errors.New("id is required")
    }
    session, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    if input.Name != nil {
        session.Name = *input.Name
    }
    if input.StartsAt != nil {
        session.StartsAt = parseTimePtr(input.StartsAt)
    }
    if input.EndsAt != nil {
        session.EndsAt = parseTimePtr(input.EndsAt)
    }
    if err := s.repo.Update(session); err != nil {
        return nil, err
    }
    return session, nil
}

func (s *examSessionService) Delete(id string) error {
    if id == "" {
        return errors.New("id is required")
    }
    if _, err := s.repo.GetByID(id); err != nil {
        return err
    }
    return s.repo.Delete(id)
}

func parseTimePtr(src *string) *time.Time {
    if src == nil || *src == "" {
        return nil
    }
    if t, err := time.Parse(time.RFC3339, *src); err == nil {
        return &t
    }
    return nil
}
