package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/middleware"
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
)

type studentUserRepo struct {
	user *models.User
}

func (r studentUserRepo) Create(context.Context, *models.User) error { return nil }
func (r studentUserRepo) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return r.user, nil
}
func (r studentUserRepo) GetByINS(context.Context, string) (*models.User, error) { return nil, nil }
func (r studentUserRepo) ListByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r studentUserRepo) ListDeletedByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r studentUserRepo) Update(context.Context, *models.User) error  { return nil }
func (r studentUserRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }
func (r studentUserRepo) Restore(context.Context, uuid.UUID) error    { return nil }
func (r studentUserRepo) AttachStudentProfile(context.Context, *models.StudentProfile) error {
	return nil
}
func (r studentUserRepo) AttachTeacherProfile(context.Context, *models.TeacherProfile) error {
	return nil
}
func (r studentUserRepo) AttachStaffProfile(context.Context, *models.StaffProfile) error { return nil }
func (r studentUserRepo) UpsertRefreshToken(context.Context, *models.RefreshToken) error { return nil }
func (r studentUserRepo) DeleteRefreshToken(context.Context, uuid.UUID) error            { return nil }
func (r studentUserRepo) NextINS(context.Context) (string, error)                        { return "00000010", nil }

type studentGroupRepo struct{}

func (studentGroupRepo) Create(context.Context, *models.Group) error { return nil }
func (studentGroupRepo) GetByID(context.Context, uuid.UUID) (*models.Group, error) {
	return &models.Group{}, nil
}
func (studentGroupRepo) List(context.Context, repository.ListOptions) ([]models.Group, int64, error) {
	return nil, 0, nil
}
func (studentGroupRepo) ListDeleted(context.Context, repository.ListOptions) ([]models.Group, int64, error) {
	return nil, 0, nil
}
func (studentGroupRepo) Update(context.Context, *models.Group) error         { return nil }
func (studentGroupRepo) SoftDelete(context.Context, uuid.UUID) error         { return nil }
func (studentGroupRepo) Restore(context.Context, uuid.UUID) error            { return nil }
func (studentGroupRepo) DetachStudents(context.Context, uuid.UUID) error     { return nil }
func (studentGroupRepo) RemoveSubjectLinks(context.Context, uuid.UUID) error { return nil }

type studentSubjectRepo struct {
	subjects []models.Subject
}

func (r studentSubjectRepo) Create(context.Context, *models.Subject) error { return nil }
func (r studentSubjectRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Subject, error) {
	for _, subj := range r.subjects {
		if subj.ID == id {
			copy := subj
			return &copy, nil
		}
	}
	return nil, nil
}
func (r studentSubjectRepo) List(context.Context, repository.ListOptions) ([]models.Subject, int64, error) {
	return nil, 0, nil
}
func (r studentSubjectRepo) ListDeleted(context.Context, repository.ListOptions) ([]models.Subject, int64, error) {
	return nil, 0, nil
}
func (r studentSubjectRepo) ListByGroup(context.Context, uuid.UUID) ([]models.Subject, error) {
	return r.subjects, nil
}
func (studentSubjectRepo) Update(context.Context, *models.Subject) error { return nil }
func (studentSubjectRepo) SoftDelete(context.Context, uuid.UUID) error   { return nil }
func (studentSubjectRepo) Restore(context.Context, uuid.UUID) error      { return nil }
func (studentSubjectRepo) AssignTeacher(context.Context, *models.TeachingAssignment) error {
	return nil
}
func (studentSubjectRepo) AttachGroup(context.Context, *models.SubjectGroup) error { return nil }
func (studentSubjectRepo) ListTeacherAssignments(context.Context, uuid.UUID) ([]models.TeachingAssignment, error) {
	return nil, nil
}
func (studentSubjectRepo) ListSubjectAssignments(context.Context, uuid.UUID) ([]models.TeachingAssignment, error) {
	return nil, nil
}
func (studentSubjectRepo) ListSubjectGroups(context.Context, uuid.UUID) ([]models.SubjectGroup, error) {
	return nil, nil
}
func (studentSubjectRepo) RemoveTeacherAssignments(context.Context, uuid.UUID) error   { return nil }
func (studentSubjectRepo) RemoveAssignmentsBySubject(context.Context, uuid.UUID) error { return nil }
func (studentSubjectRepo) RemoveGroupLinks(context.Context, uuid.UUID) error           { return nil }
func (studentSubjectRepo) RemoveTeacherAssignment(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type studentSessionRepo struct {
	sessions []models.ClassSession
}

func (r studentSessionRepo) Create(context.Context, *models.ClassSession) error { return nil }
func (r studentSessionRepo) ListBySubjectAndGroup(_ context.Context, subjectID uuid.UUID, _ uuid.UUID, _, _ *time.Time) ([]models.ClassSession, error) {
	var result []models.ClassSession
	for _, session := range r.sessions {
		if session.SubjectID == subjectID {
			result = append(result, session)
		}
	}
	return result, nil
}
func (r studentSessionRepo) GetByID(_ context.Context, id uuid.UUID) (*models.ClassSession, error) {
	for _, session := range r.sessions {
		if session.ID == id {
			copy := session
			return &copy, nil
		}
	}
	return nil, nil
}

// ListByStudent implements the repository.SessionRepository interface; return all sessions for tests.
func (r studentSessionRepo) ListByStudent(context.Context, uuid.UUID) ([]models.ClassSession, error) {
	return r.sessions, nil
}
func (studentSessionRepo) DeleteByTeacher(context.Context, uuid.UUID) error { return nil }
func (studentSessionRepo) DeleteByGroup(context.Context, uuid.UUID) error   { return nil }
func (studentSessionRepo) DeleteBySubject(context.Context, uuid.UUID) error { return nil }
func (studentSessionRepo) ListByTeacher(context.Context, uuid.UUID, *time.Time, *time.Time) ([]models.ClassSession, error) {
	return nil, nil
}
func (studentSessionRepo) ListByGroup(context.Context, uuid.UUID, *time.Time, *time.Time) ([]models.ClassSession, error) {
	return nil, nil
}
func (studentSessionRepo) ListByFilter(context.Context, repository.SessionFilter) ([]models.ClassSession, error) {
	return nil, nil
}

type studentGradeRepo struct {
	grades []models.Grade
}

func (r studentGradeRepo) GroupAverage(context.Context, uuid.UUID) (*float32, error) { return nil, nil }
func (r studentGradeRepo) GroupSubjectAverage(context.Context, uuid.UUID, uuid.UUID) (*float32, error) {
	return nil, nil
}
func (r studentGradeRepo) StudentOverallAverage(context.Context, uuid.UUID) (*float32, error) {
	return nil, nil
}
func (r studentGradeRepo) StudentSubjectAverage(context.Context, uuid.UUID, uuid.UUID) (*float32, error) {
	return nil, nil
}
func (r studentGradeRepo) ListByStudent(context.Context, uuid.UUID) ([]models.Grade, error) {
	return r.grades, nil
}
func (studentGradeRepo) ListBySubjectAndGroup(context.Context, uuid.UUID, uuid.UUID) ([]models.Grade, error) {
	return nil, nil
}
func (r studentGradeRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Grade, error) {
	for _, g := range r.grades {
		if g.ID == id {
			copy := g
			return &copy, nil
		}
	}
	return nil, nil
}
func (r studentGradeRepo) GetBySessionAndStudent(_ context.Context, sessionID uuid.UUID, studentID uuid.UUID) (*models.Grade, error) {
	for _, g := range r.grades {
		if g.SessionID == sessionID && g.StudentID == studentID {
			copy := g
			return &copy, nil
		}
	}
	return nil, nil
}
func (studentGradeRepo) Upsert(context.Context, *models.Grade) error      { return nil }
func (studentGradeRepo) Update(context.Context, *models.Grade) error      { return nil }
func (studentGradeRepo) Delete(context.Context, uuid.UUID) error          { return nil }
func (studentGradeRepo) DeleteByStudent(context.Context, uuid.UUID) error { return nil }
func (studentGradeRepo) DeleteByTeacher(context.Context, uuid.UUID) error { return nil }
func (studentGradeRepo) DeleteByGroup(context.Context, uuid.UUID) error   { return nil }
func (studentGradeRepo) DeleteBySubject(context.Context, uuid.UUID) error { return nil }

func TestStudentControllerSubjectsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	studentID := uuid.New()
	groupID := uuid.New()
	subject1 := models.Subject{Base: models.Base{ID: uuid.New()}, Code: "MAT101", Name: "Математика"}
	subject2 := models.Subject{Base: models.Base{ID: uuid.New()}, Code: "PHY101", Name: "Физика"}
	sessions := []models.ClassSession{
		{Base: models.Base{ID: uuid.New()}, SubjectID: subject1.ID, GroupID: groupID, StartsAt: time.Now()},
		{Base: models.Base{ID: uuid.New()}, SubjectID: subject2.ID, GroupID: groupID, StartsAt: time.Now()},
	}
	grades := []models.Grade{
		{Base: models.Base{ID: uuid.New()}, SubjectID: subject1.ID, SessionID: sessions[0].ID, StudentID: studentID, Value: 4.5},
	}
	student := &models.User{
		Base: models.Base{ID: studentID},
		Role: models.UserRoleStudent,
		Student: &models.StudentProfile{
			Base:    models.Base{ID: uuid.New()},
			UserID:  studentID,
			GroupID: &groupID,
			Index:   "ST-100",
		},
	}
	studentSvc := service.NewStudentService(
		studentUserRepo{user: student},
		studentGroupRepo{},
		studentSubjectRepo{subjects: []models.Subject{subject1, subject2}},
		studentSessionRepo{sessions: sessions},
		studentGradeRepo{grades: grades},
	)
	controller := NewStudentController(studentSvc)
	router := gin.New()
	group := router.Group("/student")
	group.Use(func(ctx *gin.Context) {
		ctx.Set(middleware.ContextUserIDKey, studentID.String())
	})
	controller.RegisterRoutes(group)
	req := httptest.NewRequest(http.MethodGet, "/student/subjects?limit=1&search=мат", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Data []struct {
			Subject respdto.SubjectSummary `json:"subject"`
		}
		Meta respdto.PageMeta `json:"meta"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Meta.Total != 1 {
		t.Fatalf("expected total 1, got %d", payload.Meta.Total)
	}
	if len(payload.Data) != 1 || payload.Data[0].Subject.Name != "Математика" {
		t.Fatalf("unexpected subjects payload: %+v", payload.Data)
	}
}
