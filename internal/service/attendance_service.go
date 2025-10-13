package service

import (
	"errors"
	"fmt"
	"time"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/domain/validation"
	"gradeflow/internal/repository"
)

const (
	defaultAttendanceLimit = 100
	maxAttendanceLimit     = 1000
)

type AttendanceService interface {
	Create(input req.CreateAttendance, markedBy *string) (*models.Attendance, error)
	List(input req.ListAttendanceQuery) ([]models.Attendance, int64, int, int, error)
	Get(id string) (*models.Attendance, error)
	Update(id string, input req.UpdateAttendance, markedBy *string) (*models.Attendance, error)
	Delete(id string) error
	BulkUpsert(lessonID string, items []req.AttendanceBulkItem, markedBy *string) ([]models.Attendance, error)
}

type attendanceService struct {
	repo repository.AttendanceRepository
}

func NewAttendanceService(repo repository.AttendanceRepository) AttendanceService {
	return &attendanceService{repo: repo}
}

func (s *attendanceService) Create(input req.CreateAttendance, markedBy *string) (*models.Attendance, error) {
	if err := validation.ValidateAttendanceStatus(input.Status); err != nil {
		return nil, err
	}
	if input.LessonID == "" || input.StudentID == "" {
		return nil, errors.New("lessonId and studentId are required")
	}
	att := &models.Attendance{
		LessonID:  input.LessonID,
		StudentID: input.StudentID,
		Status:    input.Status,
		MarkedAt:  time.Now(),
	}
	if markedBy != nil {
		att.MarkedBy = *markedBy
	}
	if err := s.repo.Create(att); err != nil {
		return nil, err
	}
	return att, nil
}

func (s *attendanceService) List(input req.ListAttendanceQuery) ([]models.Attendance, int64, int, int, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultAttendanceLimit
	}
	if limit > maxAttendanceLimit {
		limit = maxAttendanceLimit
	}
	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	var fromPtr, toPtr *time.Time
	if input.From != "" {
		if t, err := time.Parse(time.RFC3339, input.From); err == nil {
			fromPtr = &t
		}
	}
	if input.To != "" {
		if t, err := time.Parse(time.RFC3339, input.To); err == nil {
			toPtr = &t
		}
	}

	filter := repository.AttendanceFilter{
		LessonID:  input.LessonID,
		StudentID: input.StudentID,
		CourseID:  input.CourseID,
		SessionID: input.SessionID,
		From:      fromPtr,
		To:        toPtr,
		Limit:     limit,
		Offset:    offset,
	}
	items, total, err := s.repo.List(filter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, limit, offset, nil
}

func (s *attendanceService) Get(id string) (*models.Attendance, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	att, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return att, nil
}

func (s *attendanceService) Update(id string, input req.UpdateAttendance, markedBy *string) (*models.Attendance, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	att, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.Status != "" {
		if err := validation.ValidateAttendanceStatus(input.Status); err != nil {
			return nil, err
		}
		att.Status = input.Status
	}
	if markedBy != nil {
		att.MarkedBy = *markedBy
	}
	if input.Status != "" || markedBy != nil {
		att.MarkedAt = time.Now()
	}
	if err := s.repo.Update(att); err != nil {
		return nil, fmt.Errorf("update attendance: %w", err)
	}
	return att, nil
}

func (s *attendanceService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *attendanceService) BulkUpsert(lessonID string, items []req.AttendanceBulkItem, markedBy *string) ([]models.Attendance, error) {
	if lessonID == "" {
		return nil, errors.New("lessonId is required")
	}
	if len(items) == 0 {
		return []models.Attendance{}, nil
	}
	upserts := make([]repository.AttendanceUpsert, 0, len(items))
	for _, item := range items {
		if err := validation.ValidateAttendanceStatus(item.Status); err != nil {
			return nil, err
		}
		upserts = append(upserts, repository.AttendanceUpsert{StudentID: item.StudentID, Status: item.Status})
	}
	if err := s.repo.BulkUpsert(lessonID, upserts, markedBy, time.Now()); err != nil {
		return nil, err
	}
	// return current state for lesson
	filter := repository.AttendanceFilter{LessonID: lessonID, Limit: maxAttendanceLimit, Offset: 0}
	itemsOut, _, err := s.repo.List(filter)
	if err != nil {
		return nil, err
	}
	return itemsOut, nil
}
