package service

import (
	"errors"

	req "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

const (
	defaultCourseLimit = 100
	maxCourseLimit     = 500
)

type CourseService interface {
	Create(input req.CreateCourse) (*models.Course, error)
	List(query req.ListCourseQuery) ([]models.Course, int64, int, int, error)
	Get(id string) (*models.Course, error)
	Update(id string, input req.UpdateCourse) (*models.Course, error)
	Delete(id string) error
}

type courseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) CourseService {
	return &courseService{repo: repo}
}

func (s *courseService) Create(input req.CreateCourse) (*models.Course, error) {
	if input.SubjectID == "" || input.DepartmentID == "" || input.AcademicSessionID == "" || input.Title == "" {
		return nil, errors.New("subjectId, departmentId, academicSessionId, and title are required")
	}
	course := &models.Course{
		SubjectID:         input.SubjectID,
		DepartmentID:      input.DepartmentID,
		ProgramID:         input.ProgramID,
		AcademicSessionID: input.AcademicSessionID,
		Title:             input.Title,
		TeacherID:         input.TeacherID,
		Room:              input.Room,
	}
	if err := s.repo.Create(course); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *courseService) List(query req.ListCourseQuery) ([]models.Course, int64, int, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultCourseLimit
	}
	if limit > maxCourseLimit {
		limit = maxCourseLimit
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

func (s *courseService) Get(id string) (*models.Course, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(id)
}

func (s *courseService) Update(id string, input req.UpdateCourse) (*models.Course, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	course, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if input.SubjectID != "" {
		course.SubjectID = input.SubjectID
	}
	if input.DepartmentID != "" {
		course.DepartmentID = input.DepartmentID
	}
	if input.ProgramID != nil {
		course.ProgramID = input.ProgramID
	}
	if input.AcademicSessionID != "" {
		course.AcademicSessionID = input.AcademicSessionID
	}
	if input.Title != "" {
		course.Title = input.Title
	}
	if input.TeacherID != nil {
		course.TeacherID = input.TeacherID
	}
	if input.Room != "" {
		course.Room = input.Room
	}
	if err := s.repo.Update(course); err != nil {
		return nil, err
	}
	return course, nil
}

func (s *courseService) Delete(id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
