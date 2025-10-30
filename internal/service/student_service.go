package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/utils"
)

// StudentService orchestrates student-facing use-cases.
type StudentService struct {
	users    repository.UserRepository
	groups   repository.GroupRepository
	subjects repository.SubjectRepository
	sessions repository.SessionRepository
	grades   repository.GradeRepository
}

// NewStudentService constructs service.
func NewStudentService(users repository.UserRepository, groups repository.GroupRepository, subjects repository.SubjectRepository, sessions repository.SessionRepository, grades repository.GradeRepository) *StudentService {
	return &StudentService{users: users, groups: groups, subjects: subjects, sessions: sessions, grades: grades}
}

// Dashboard composes student profile and aggregate metrics.
func (s *StudentService) Dashboard(ctx context.Context, studentID uuid.UUID) (*respdto.StudentDashboardResponse, error) {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	resp := &respdto.StudentDashboardResponse{
		Profile: respdto.UserProfile{
			ID:         user.ID.String(),
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			MiddleName: user.MiddleName,
			Email:      user.Email,
			INS:        user.INS,
			AvatarURL:  user.AvatarURL,
			Role:       string(user.Role),
		},
	}
	var groupID *uuid.UUID
	if user.Student != nil && user.Student.GroupID != nil {
		groupID = user.Student.GroupID
		group, err := s.groups.GetByID(ctx, *groupID)
		if err == nil {
			resp.Group = &respdto.GroupSummary{ID: group.ID.String(), Name: group.Name, Description: group.Description}
		}
	}
	overall, err := s.grades.StudentOverallAverage(ctx, studentID)
	if err == nil {
		resp.AverageGPA = overall
	}
	return resp, nil
}

// Subjects returns subjects with grade history for the student.
func (s *StudentService) Subjects(ctx context.Context, studentID uuid.UUID) ([]respdto.StudentSubjectGrade, error) {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	grades, err := s.grades.ListByStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("list student grades: %w", err)
	}
	gradesBySubject := make(map[uuid.UUID][]models.Grade)
	for _, gr := range grades {
		gradesBySubject[gr.SubjectID] = append(gradesBySubject[gr.SubjectID], gr)
	}
	var subjects []models.Subject
	if user.Student != nil && user.Student.GroupID != nil {
		subjects, err = s.subjects.ListByGroup(ctx, *user.Student.GroupID)
		if err != nil {
			return nil, fmt.Errorf("list subjects by group: %w", err)
		}
	}
	// ensure subjects from grades included even if group missing
	for subjectID := range gradesBySubject {
		found := false
		for _, subj := range subjects {
			if subj.ID == subjectID {
				found = true
				break
			}
		}
		if !found {
			subj, err := s.subjects.GetByID(ctx, subjectID)
			if err == nil {
				subjects = append(subjects, *subj)
			}
		}
	}
	response := make([]respdto.StudentSubjectGrade, 0, len(subjects))
	for _, subject := range subjects {
		var sessions []models.ClassSession
		if user.Student != nil && user.Student.GroupID != nil {
			sessions, err = s.sessions.ListBySubjectAndGroup(ctx, subject.ID, *user.Student.GroupID, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("list sessions: %w", err)
			}
		}
		gradesForSubject := gradesBySubject[subject.ID]
		gradeMap := make(map[uuid.UUID]models.Grade)
		for _, gr := range gradesForSubject {
			gradeMap[gr.SessionID] = gr
		}
		// Build session list fallback from grades when sessions absent
		if len(sessions) == 0 {
			for _, gr := range gradesForSubject {
				session, err := s.sessions.GetByID(ctx, gr.SessionID)
				if err == nil {
					sessions = append(sessions, *session)
				}
			}
		}
		sessionResponses := make([]respdto.StudentSessionGrade, 0, len(sessions))
		for _, session := range sessions {
			entry := respdto.StudentSessionGrade{
				Session: respdto.SessionSummary{
					ID:        session.ID.String(),
					StartsAt:  session.StartsAt,
					EndsAt:    session.EndsAt,
					Topic:     session.Topic,
					SubjectID: session.SubjectID.String(),
					GroupID:   session.GroupID.String(),
				},
			}
			if grade, ok := gradeMap[session.ID]; ok {
				value := grade.Value
				entry.Grade = &value
				entry.GradeID = utils.StringPtr(grade.ID.String())
				entry.Notes = grade.Notes
			}
			sessionResponses = append(sessionResponses, entry)
		}
		avg, _ := s.grades.StudentSubjectAverage(ctx, studentID, subject.ID)
		response = append(response, respdto.StudentSubjectGrade{
			Subject:  respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description},
			Sessions: sessionResponses,
			Average:  avg,
		})
	}
	return response, nil
}

// SubjectAverage returns average metrics for a student's subject.
func (s *StudentService) SubjectAverage(ctx context.Context, studentID, subjectID uuid.UUID) (*respdto.AverageMetricResponse, error) {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	studentAvg, err := s.grades.StudentSubjectAverage(ctx, studentID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("student subject average: %w", err)
	}
	var groupAvg *float32
	var overall *float32
	if user.Student != nil && user.Student.GroupID != nil {
		groupAvg, _ = s.grades.GroupSubjectAverage(ctx, *user.Student.GroupID, subjectID)
		overall, _ = s.grades.GroupAverage(ctx, *user.Student.GroupID)
	}
	return &respdto.AverageMetricResponse{SubjectAverage: studentAvg, GroupAverage: groupAvg, OverallAverage: overall}, nil
}

// Schedule returns student's timetable for their group.
func (s *StudentService) Schedule(ctx context.Context, studentID uuid.UUID, query reqdto.ScheduleQuery) ([]respdto.ScheduleEntry, error) {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	if user.Student == nil || user.Student.GroupID == nil {
		return []respdto.ScheduleEntry{}, nil
	}
	filter := repository.SessionFilter{GroupID: user.Student.GroupID}
	if id, err := uuidFromStringPtr(query.SubjectID); err != nil {
		return nil, fmt.Errorf("parse subjectId: %w", err)
	} else if id != nil {
		filter.SubjectID = id
	}
	filter.From = query.From
	filter.To = query.To
	sessions, err := s.sessions.ListByFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return buildScheduleEntries(sessions), nil
}
