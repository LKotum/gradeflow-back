package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

// DeanService exposes dean office orchestration use-cases.
type DeanService struct {
	users    repository.UserRepository
	groups   repository.GroupRepository
	subjects repository.SubjectRepository
	sessions repository.SessionRepository
	grades   repository.GradeRepository
}

// NewDeanService constructs the service.
func NewDeanService(users repository.UserRepository, groups repository.GroupRepository, subjects repository.SubjectRepository, sessions repository.SessionRepository, grades repository.GradeRepository) *DeanService {
	return &DeanService{
		users:    users,
		groups:   groups,
		subjects: subjects,
		sessions: sessions,
		grades:   grades,
	}
}

// CreateGroup provisions a new student group.
func (s *DeanService) CreateGroup(ctx context.Context, payload reqdto.CreateGroupRequest) (*respdto.GroupSummary, error) {
    group := &models.Group{
        Base:        models.Base{ID: uuid.New()},
        Name:        payload.Name,
        Description: payload.Description,
    }
	if err := s.groups.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	return &respdto.GroupSummary{ID: group.ID.String(), Name: group.Name, Description: group.Description}, nil
}

// ListGroups returns available groups.
func (s *DeanService) ListGroups(ctx context.Context) ([]respdto.GroupSummary, error) {
	groups, err := s.groups.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	resp := make([]respdto.GroupSummary, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, respdto.GroupSummary{ID: g.ID.String(), Name: g.Name, Description: g.Description})
	}
	return resp, nil
}

// CreateSubject registers a new subject.
func (s *DeanService) CreateSubject(ctx context.Context, payload reqdto.CreateSubjectRequest) (*respdto.SubjectSummary, error) {
    subject := &models.Subject{
        Base:        models.Base{ID: uuid.New()},
        Code:        payload.Code,
        Name:        payload.Name,
        Description: payload.Description,
    }
	if err := s.subjects.Create(ctx, subject); err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}
	return &respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description}, nil
}

// ListSubjects returns subjects.
func (s *DeanService) ListSubjects(ctx context.Context) ([]respdto.SubjectSummary, error) {
	subjects, err := s.subjects.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	resp := make([]respdto.SubjectSummary, 0, len(subjects))
	for _, subj := range subjects {
		resp = append(resp, respdto.SubjectSummary{ID: subj.ID.String(), Code: subj.Code, Name: subj.Name, Description: subj.Description})
	}
	return resp, nil
}

// CreateTeacher provisions teacher account.
func (s *DeanService) CreateTeacher(ctx context.Context, payload reqdto.CreateTeacherRequest) (*respdto.UserProfile, error) {
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }
    ins := payload.INS
    if ins == "" {
        generated, err := s.users.NextINS(ctx)
        if err != nil {
            return nil, fmt.Errorf("generate INS: %w", err)
        }
        ins = generated
    }
    user := &models.User{
        Base:         models.Base{ID: uuid.New()},
        Role:         models.UserRoleTeacher,
        INS:          &ins,
        Email:        payload.Email,
        FirstName:    payload.FirstName,
        LastName:     payload.LastName,
        MiddleName:   payload.MiddleName,
        PasswordHash: string(passwordHash),
    }
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create teacher user: %w", err)
	}
    if err := s.users.AttachTeacherProfile(ctx, &models.TeacherProfile{
        Base:    models.Base{ID: uuid.New()},
        UserID:  user.ID,
        Title:   payload.Title,
        Bio:     payload.Bio,
    }); err != nil {
        return nil, fmt.Errorf("create teacher profile: %w", err)
    }
	return &respdto.UserProfile{
		ID:         user.ID.String(),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		INS:        user.INS,
		Role:       string(user.Role),
	}, nil
}

// ListTeachers returns teacher summaries.
func (s *DeanService) ListTeachers(ctx context.Context) ([]respdto.UserProfile, error) {
	teachers, err := s.users.ListByRole(ctx, models.UserRoleTeacher)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	resp := make([]respdto.UserProfile, 0, len(teachers))
	for _, t := range teachers {
		resp = append(resp, respdto.UserProfile{
			ID:         t.ID.String(),
			FirstName:  t.FirstName,
			LastName:   t.LastName,
			MiddleName: t.MiddleName,
			Email:      t.Email,
			INS:        t.INS,
			AvatarURL:  t.AvatarURL,
			Role:       string(t.Role),
		})
	}
	return resp, nil
}

// CreateStudent provisions student account.
func (s *DeanService) CreateStudent(ctx context.Context, payload reqdto.CreateStudentRequest) (*respdto.UserProfile, error) {
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }
    ins := payload.INS
    if ins == "" {
        generated, err := s.users.NextINS(ctx)
        if err != nil {
            return nil, fmt.Errorf("generate INS: %w", err)
        }
        ins = generated
    }
    user := &models.User{
        Base:         models.Base{ID: uuid.New()},
        Role:         models.UserRoleStudent,
        INS:          &ins,
        Email:        payload.Email,
        FirstName:    payload.FirstName,
        LastName:     payload.LastName,
        MiddleName:   payload.MiddleName,
        PasswordHash: string(passwordHash),
    }
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create student user: %w", err)
	}
    profile := &models.StudentProfile{
        Base:   models.Base{ID: uuid.New()},
        UserID: user.ID,
        Index:  payload.Index,
    }
	if payload.GroupID != nil {
		gid, err := uuid.Parse(*payload.GroupID)
		if err != nil {
			return nil, fmt.Errorf("parse groupId: %w", err)
		}
		profile.GroupID = &gid
	}
	if err := s.users.AttachStudentProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("attach student profile: %w", err)
	}
	return &respdto.UserProfile{
		ID:         user.ID.String(),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		INS:        user.INS,
		Role:       string(user.Role),
	}, nil
}

// ListStudents returns student summaries.
func (s *DeanService) ListStudents(ctx context.Context) ([]respdto.UserProfile, error) {
	students, err := s.users.ListByRole(ctx, models.UserRoleStudent)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	resp := make([]respdto.UserProfile, 0, len(students))
	for _, st := range students {
		resp = append(resp, respdto.UserProfile{
			ID:         st.ID.String(),
			FirstName:  st.FirstName,
			LastName:   st.LastName,
			MiddleName: st.MiddleName,
			Email:      st.Email,
			INS:        st.INS,
			AvatarURL:  st.AvatarURL,
			Role:       string(st.Role),
		})
	}
	return resp, nil
}

// AssignTeacher connects a teacher to a subject.
func (s *DeanService) AssignTeacher(ctx context.Context, subjectID uuid.UUID, payload reqdto.AssignTeacherRequest) error {
	teacherID, err := uuid.Parse(payload.TeacherID)
	if err != nil {
		return fmt.Errorf("parse teacher id: %w", err)
	}
    assignment := &models.TeachingAssignment{
        Base:      models.Base{ID: uuid.New()},
        SubjectID: subjectID,
        TeacherID: teacherID,
    }
	return s.subjects.AssignTeacher(ctx, assignment)
}

// AttachGroup links group to subject.
func (s *DeanService) AttachGroup(ctx context.Context, subjectID uuid.UUID, payload reqdto.AttachGroupRequest) error {
	groupID, err := uuid.Parse(payload.GroupID)
	if err != nil {
		return fmt.Errorf("parse group id: %w", err)
	}
    link := &models.SubjectGroup{
        Base:      models.Base{ID: uuid.New()},
        SubjectID: subjectID,
        GroupID:   groupID,
    }
	return s.subjects.AttachGroup(ctx, link)
}

// AssignStudentToGroup moves student into a group.
func (s *DeanService) AssignStudentToGroup(ctx context.Context, groupID uuid.UUID, payload reqdto.AssignStudentToGroupRequest) error {
	studentID, err := uuid.Parse(payload.StudentID)
	if err != nil {
		return fmt.Errorf("parse student id: %w", err)
	}
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return fmt.Errorf("load student: %w", err)
	}
	if user.Student == nil {
		return errors.New("user has no student profile")
	}
	profile := &models.StudentProfile{
		UserID:  user.ID,
		Index:   user.Student.Index,
		GroupID: &groupID,
	}
	if user.Student.ID != uuid.Nil {
		profile.ID = user.Student.ID
	}
	return s.users.AttachStudentProfile(ctx, profile)
}

// ScheduleSession creates a lesson for subject/group/teacher.
func (s *DeanService) ScheduleSession(ctx context.Context, payload reqdto.ScheduleSessionRequest) (*respdto.SessionSummary, error) {
	subjectID, err := uuid.Parse(payload.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("parse subject id: %w", err)
	}
	groupID, err := uuid.Parse(payload.GroupID)
	if err != nil {
		return nil, fmt.Errorf("parse group id: %w", err)
	}
	teacherID, err := uuid.Parse(payload.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("parse teacher id: %w", err)
	}
	assignments, err := s.subjects.ListTeacherAssignments(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	valid := false
	for _, a := range assignments {
		if a.SubjectID == subjectID {
			valid = true
			break
		}
	}
	if !valid {
		return nil, errors.New("teacher not assigned to subject")
	}
    session := &models.ClassSession{
        Base:      models.Base{ID: uuid.New()},
        SubjectID: subjectID,
        GroupID:   groupID,
        TeacherID: teacherID,
        StartsAt:  payload.StartsAt,
        EndsAt:    payload.EndsAt,
        Topic:     payload.Topic,
    }
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &respdto.SessionSummary{
		ID:        session.ID.String(),
		StartsAt:  session.StartsAt,
		EndsAt:    session.EndsAt,
		Topic:     session.Topic,
		SubjectID: session.SubjectID.String(),
		GroupID:   session.GroupID.String(),
	}, nil
}

// GroupRanking composes ranking by average grade.
func (s *DeanService) GroupRanking(ctx context.Context) (*respdto.GroupRankingResponse, error) {
	groups, err := s.groups.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	resp := respdto.GroupRankingResponse{Items: make([]respdto.GroupRankingItem, 0, len(groups))}
	for _, g := range groups {
		avg, err := s.grades.GroupAverage(ctx, g.ID)
		if err != nil {
			return nil, fmt.Errorf("group average: %w", err)
		}
		resp.Items = append(resp.Items, respdto.GroupRankingItem{
			Group:   respdto.GroupSummary{ID: g.ID.String(), Name: g.Name, Description: g.Description},
			Average: avg,
		})
	}
	return &resp, nil
}
