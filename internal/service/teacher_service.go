package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

// TeacherService handles teacher-facing operations.
type TeacherService struct {
	users    repository.UserRepository
	groups   repository.GroupRepository
	subjects repository.SubjectRepository
	sessions repository.SessionRepository
	grades   repository.GradeRepository
	now      func() time.Time
}

// NewTeacherService constructs service.
func NewTeacherService(users repository.UserRepository, groups repository.GroupRepository, subjects repository.SubjectRepository, sessions repository.SessionRepository, grades repository.GradeRepository) *TeacherService {
	return &TeacherService{
		users:    users,
		groups:   groups,
		subjects: subjects,
		sessions: sessions,
		grades:   grades,
		now:      time.Now,
	}
}

// Dashboard aggregates teacher profile and assigned subjects/groups.
func (s *TeacherService) Dashboard(ctx context.Context, teacherID uuid.UUID) (*respdto.TeacherDashboardResponse, error) {
	user, err := s.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("load teacher: %w", err)
	}
	assignments, err := s.subjects.ListTeacherAssignments(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	subjectsCache := make(map[uuid.UUID]models.Subject)
	groupsCache := make(map[uuid.UUID]models.Group)
	resp := &respdto.TeacherDashboardResponse{
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
	for _, assignment := range assignments {
		subject, ok := subjectsCache[assignment.SubjectID]
		if !ok {
			subj, err := s.subjects.GetByID(ctx, assignment.SubjectID)
			if err != nil {
				return nil, fmt.Errorf("load subject: %w", err)
			}
			subjectsCache[assignment.SubjectID] = *subj
			subject = *subj
		}
		links, err := s.subjects.ListSubjectGroups(ctx, assignment.SubjectID)
		if err != nil {
			return nil, fmt.Errorf("list subject groups: %w", err)
		}
		groupSummaries := make([]respdto.GroupSummary, 0, len(links))
		for _, link := range links {
			group, ok := groupsCache[link.GroupID]
			if !ok {
				grp, err := s.groups.GetByID(ctx, link.GroupID)
				if err != nil {
					return nil, fmt.Errorf("load group: %w", err)
				}
				groupsCache[link.GroupID] = *grp
				group = *grp
			}
			groupSummaries = append(groupSummaries, respdto.GroupSummary{ID: group.ID.String(), Name: group.Name, Description: group.Description})
		}
		resp.Subjects = append(resp.Subjects, respdto.TeacherSubjectSummary{
			Subject: respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description},
			Groups:  groupSummaries,
		})
	}
	return resp, nil
}

// GradeTable returns grade information for subject/group pair.
func (s *TeacherService) GradeTable(ctx context.Context, teacherID, subjectID, groupID uuid.UUID) (*respdto.GradeTableResponse, error) {
	// ensure assignment
	assignments, err := s.subjects.ListTeacherAssignments(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	allowed := false
	for _, a := range assignments {
		if a.SubjectID == subjectID {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("teacher not assigned to subject")
	}
	subject, err := s.subjects.GetByID(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("load subject: %w", err)
	}
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	sessions, err := s.sessions.ListBySubjectAndGroup(ctx, subjectID, groupID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	filteredSessions := make([]models.ClassSession, 0, len(sessions))
	for _, session := range sessions {
		if session.TeacherID == teacherID {
			filteredSessions = append(filteredSessions, session)
		}
	}
	grades, err := s.grades.ListBySubjectAndGroup(ctx, subjectID, groupID)
	if err != nil {
		return nil, fmt.Errorf("list grades: %w", err)
	}
	resp := &respdto.GradeTableResponse{
		Subject: respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description},
		Group:   respdto.GroupSummary{ID: group.ID.String(), Name: group.Name, Description: group.Description},
	}
	resp.Sessions = make([]respdto.SessionSummary, 0, len(filteredSessions))
	for _, session := range filteredSessions {
		s := session // copy for pointer
		resp.Sessions = append(resp.Sessions, respdto.SessionSummary{
			ID:        s.ID.String(),
			StartsAt:  s.StartsAt,
			EndsAt:    s.EndsAt,
			Topic:     s.Topic,
			SubjectID: s.SubjectID.String(),
			GroupID:   s.GroupID.String(),
		})
	}

	studentProfiles := make([]respdto.UserProfile, 0, len(group.Students))
	studentIndex := make(map[uuid.UUID]models.User, len(group.Students))
	for _, profile := range group.Students {
		user := profile.User
		if user.ID == uuid.Nil {
			loaded, err := s.users.GetByID(ctx, profile.UserID)
			if err != nil {
				return nil, fmt.Errorf("load student: %w", err)
			}
			user = *loaded
		}
		studentIndex[user.ID] = user
		studentProfiles = append(studentProfiles, respdto.UserProfile{
			ID:         user.ID.String(),
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			MiddleName: user.MiddleName,
			Email:      user.Email,
			INS:        user.INS,
			AvatarURL:  user.AvatarURL,
			Role:       string(user.Role),
		})
	}
	resp.Students = studentProfiles
	resp.Meta = &respdto.PageMeta{Limit: len(studentProfiles), Offset: 0, Total: len(studentProfiles)}

	gradeLookup := make(map[uuid.UUID]map[uuid.UUID]models.Grade)
	for _, grade := range grades {
		if grade.TeacherID != teacherID {
			continue
		}
		if _, ok := gradeLookup[grade.SessionID]; !ok {
			gradeLookup[grade.SessionID] = make(map[uuid.UUID]models.Grade)
		}
		gradeLookup[grade.SessionID][grade.StudentID] = grade
	}

	for _, session := range filteredSessions {
		for _, profile := range group.Students {
			user, ok := studentIndex[profile.UserID]
			if !ok {
				loaded, err := s.users.GetByID(ctx, profile.UserID)
				if err != nil {
					return nil, fmt.Errorf("load student: %w", err)
				}
				user = *loaded
			}
			detail := respdto.GradeDetail{
				SessionID: session.ID.String(),
				Student: respdto.UserProfile{
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
			if byStudent, ok := gradeLookup[session.ID]; ok {
				if grade, ok := byStudent[user.ID]; ok {
					value := grade.Value
					detail.GradeID = stringPtr(grade.ID.String())
					detail.Value = &value
					detail.Notes = grade.Notes
					assessedAt := grade.AssessedAt
					detail.AssessedAt = &assessedAt
				}
			}
			resp.Grades = append(resp.Grades, detail)
		}
	}
	return resp, nil
}

// UpsertGrade creates or replaces a grade for specific session/student.
func (s *TeacherService) UpsertGrade(ctx context.Context, teacherID uuid.UUID, payload reqdto.GradeUpsertRequest) (*respdto.GradeDetail, error) {
	sessionID, err := uuid.Parse(payload.SessionID)
	if err != nil {
		return nil, fmt.Errorf("parse session id: %w", err)
	}
	studentID, err := uuid.Parse(payload.StudentID)
	if err != nil {
		return nil, fmt.Errorf("parse student id: %w", err)
	}
	session, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}
	if session.TeacherID != teacherID {
		return nil, errors.New("cannot grade session owned by another teacher")
	}
	grade := &models.Grade{
		SessionID:  sessionID,
		StudentID:  studentID,
		SubjectID:  session.SubjectID,
		TeacherID:  teacherID,
		Value:      payload.Value,
		Notes:      payload.Notes,
		AssessedAt: s.now(),
	}
	if err := s.grades.Upsert(ctx, grade); err != nil {
		return nil, fmt.Errorf("upsert grade: %w", err)
	}
	stored, err := s.grades.GetBySessionAndStudent(ctx, sessionID, studentID)
	if err != nil {
		return nil, fmt.Errorf("reload grade: %w", err)
	}
	student, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	value := stored.Value
	return &respdto.GradeDetail{
		GradeID:   stringPtr(stored.ID.String()),
		SessionID: stored.SessionID.String(),
		Student: respdto.UserProfile{
			ID:         student.ID.String(),
			FirstName:  student.FirstName,
			LastName:   student.LastName,
			MiddleName: student.MiddleName,
			Email:      student.Email,
			INS:        student.INS,
			AvatarURL:  student.AvatarURL,
			Role:       string(student.Role),
		},
		Value:      &value,
		Notes:      stored.Notes,
		AssessedAt: &stored.AssessedAt,
	}, nil
}

// UpdateGrade modifies an existing grade.
func (s *TeacherService) UpdateGrade(ctx context.Context, teacherID, gradeID uuid.UUID, payload reqdto.GradeUpdateRequest) (*respdto.GradeDetail, error) {
	grade, err := s.grades.GetByID(ctx, gradeID)
	if err != nil {
		return nil, fmt.Errorf("load grade: %w", err)
	}
	if grade.TeacherID != teacherID {
		return nil, errors.New("cannot modify grade from another teacher")
	}
	grade.Value = payload.Value
	grade.Notes = payload.Notes
	grade.AssessedAt = s.now()
	if err := s.grades.Update(ctx, grade); err != nil {
		return nil, fmt.Errorf("update grade: %w", err)
	}
	student, err := s.users.GetByID(ctx, grade.StudentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	value := grade.Value
	return &respdto.GradeDetail{
		GradeID:   stringPtr(grade.ID.String()),
		SessionID: grade.SessionID.String(),
		Student: respdto.UserProfile{
			ID:         student.ID.String(),
			FirstName:  student.FirstName,
			LastName:   student.LastName,
			MiddleName: student.MiddleName,
			Email:      student.Email,
			INS:        student.INS,
			AvatarURL:  student.AvatarURL,
			Role:       string(student.Role),
		},
		Value:      &value,
		Notes:      grade.Notes,
		AssessedAt: &grade.AssessedAt,
	}, nil
}

// DeleteGrade removes grade owned by teacher.
func (s *TeacherService) DeleteGrade(ctx context.Context, teacherID, gradeID uuid.UUID) error {
	grade, err := s.grades.GetByID(ctx, gradeID)
	if err != nil {
		return fmt.Errorf("load grade: %w", err)
	}
	if grade.TeacherID != teacherID {
		return errors.New("cannot delete grade from another teacher")
	}
	return s.grades.Delete(ctx, gradeID)
}

// SubjectAverages calculates averages for a student and group.
func (s *TeacherService) SubjectAverages(ctx context.Context, groupID, subjectID, studentID uuid.UUID) (*respdto.AverageMetricResponse, error) {
	studentAvg, err := s.grades.StudentSubjectAverage(ctx, studentID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("student subject avg: %w", err)
	}
	groupAvg, err := s.grades.GroupSubjectAverage(ctx, groupID, subjectID)
	if err != nil {
		return nil, fmt.Errorf("group subject avg: %w", err)
	}
	overall, err := s.grades.GroupAverage(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("group overall avg: %w", err)
	}
	return &respdto.AverageMetricResponse{
		SubjectAverage: studentAvg,
		GroupAverage:   groupAvg,
		OverallAverage: overall,
	}, nil
}

// Schedule returns teacher timetable filtered by optional parameters.
func (s *TeacherService) Schedule(ctx context.Context, teacherID uuid.UUID, query reqdto.ScheduleQuery) ([]respdto.ScheduleEntry, error) {
	filter := repository.SessionFilter{TeacherID: &teacherID}
	if id, err := uuidFromStringPtr(query.SubjectID); err != nil {
		return nil, fmt.Errorf("parse subjectId: %w", err)
	} else if id != nil {
		filter.SubjectID = id
	}
	if id, err := uuidFromStringPtr(query.GroupID); err != nil {
		return nil, fmt.Errorf("parse groupId: %w", err)
	} else if id != nil {
		filter.GroupID = id
	}
	filter.From = query.From
	filter.To = query.To
	sessions, err := s.sessions.ListByFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return buildScheduleEntries(sessions), nil
}
