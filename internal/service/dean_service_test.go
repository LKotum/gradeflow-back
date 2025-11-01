package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/cache"
)

type deanUserRepo struct {
	users    map[uuid.UUID]*models.User
	students map[uuid.UUID]*models.StudentProfile
}

func newDeanUserRepo() *deanUserRepo {
	return &deanUserRepo{
		users:    make(map[uuid.UUID]*models.User),
		students: make(map[uuid.UUID]*models.StudentProfile),
	}
}

func (r *deanUserRepo) Create(_ context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *deanUserRepo) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *deanUserRepo) GetByINS(context.Context, string) (*models.User, error) { return nil, nil }

func (r *deanUserRepo) ListByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *deanUserRepo) ListDeletedByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *deanUserRepo) Update(context.Context, *models.User) error { return nil }

func (r *deanUserRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }

func (r *deanUserRepo) Restore(context.Context, uuid.UUID) error { return nil }

func (r *deanUserRepo) RefreshToken(context.Context, uuid.UUID) (*models.RefreshToken, error) {
	return nil, nil
}

func (r *deanUserRepo) DeleteRefreshToken(context.Context, uuid.UUID) error { return nil }

func (r *deanUserRepo) AttachStudentProfile(_ context.Context, profile *models.StudentProfile) error {
	if existing, ok := r.users[profile.UserID]; ok {
		existing.Student = &models.StudentProfile{
			Base:    profile.Base,
			UserID:  profile.UserID,
			GroupID: profile.GroupID,
			Index:   profile.Index,
		}
	}
	r.students[profile.UserID] = &models.StudentProfile{
		Base:    profile.Base,
		UserID:  profile.UserID,
		GroupID: profile.GroupID,
		Index:   profile.Index,
	}
	return nil
}

func (r *deanUserRepo) AttachTeacherProfile(context.Context, *models.TeacherProfile) error {
	return nil
}

func (r *deanUserRepo) AttachStaffProfile(context.Context, *models.StaffProfile) error { return nil }

func (r *deanUserRepo) UpsertRefreshToken(context.Context, *models.RefreshToken) error { return nil }

func (r *deanUserRepo) NextINS(context.Context) (string, error) { return "00000042", nil }

type noopGroupRepo struct{}

func (noopGroupRepo) Create(context.Context, *models.Group) error { return nil }
func (noopGroupRepo) GetByID(context.Context, uuid.UUID) (*models.Group, error) {
	return &models.Group{}, nil
}
func (noopGroupRepo) List(context.Context, repository.ListOptions) ([]models.Group, int64, error) {
	return nil, 0, nil
}
func (noopGroupRepo) ListDeleted(context.Context, repository.ListOptions) ([]models.Group, int64, error) {
	return nil, 0, nil
}
func (noopGroupRepo) Update(context.Context, *models.Group) error         { return nil }
func (noopGroupRepo) SoftDelete(context.Context, uuid.UUID) error         { return nil }
func (noopGroupRepo) Restore(context.Context, uuid.UUID) error            { return nil }
func (noopGroupRepo) DetachStudents(context.Context, uuid.UUID) error     { return nil }
func (noopGroupRepo) RemoveSubjectLinks(context.Context, uuid.UUID) error { return nil }

type noopSubjectRepo struct{}

func (noopSubjectRepo) Create(context.Context, *models.Subject) error { return nil }
func (noopSubjectRepo) GetByID(context.Context, uuid.UUID) (*models.Subject, error) {
	return &models.Subject{}, nil
}
func (noopSubjectRepo) List(context.Context, repository.ListOptions) ([]models.Subject, int64, error) {
	return nil, 0, nil
}
func (noopSubjectRepo) ListDeleted(context.Context, repository.ListOptions) ([]models.Subject, int64, error) {
	return nil, 0, nil
}
func (noopSubjectRepo) ListByGroup(context.Context, uuid.UUID) ([]models.Subject, error) {
	return nil, nil
}
func (noopSubjectRepo) Update(context.Context, *models.Subject) error                   { return nil }
func (noopSubjectRepo) SoftDelete(context.Context, uuid.UUID) error                     { return nil }
func (noopSubjectRepo) Restore(context.Context, uuid.UUID) error                        { return nil }
func (noopSubjectRepo) AssignTeacher(context.Context, *models.TeachingAssignment) error { return nil }
func (noopSubjectRepo) AttachGroup(context.Context, *models.SubjectGroup) error         { return nil }
func (noopSubjectRepo) ListTeacherAssignments(context.Context, uuid.UUID) ([]models.TeachingAssignment, error) {
	return nil, nil
}
func (noopSubjectRepo) ListSubjectAssignments(context.Context, uuid.UUID) ([]models.TeachingAssignment, error) {
	return nil, nil
}
func (noopSubjectRepo) ListSubjectGroups(context.Context, uuid.UUID) ([]models.SubjectGroup, error) {
	return nil, nil
}
func (noopSubjectRepo) RemoveTeacherAssignments(context.Context, uuid.UUID) error   { return nil }
func (noopSubjectRepo) RemoveAssignmentsBySubject(context.Context, uuid.UUID) error { return nil }
func (noopSubjectRepo) RemoveGroupLinks(context.Context, uuid.UUID) error           { return nil }
func (noopSubjectRepo) RemoveTeacherAssignment(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type subjectRepoMock struct {
	noopSubjectRepo
	subjects    map[uuid.UUID]*models.Subject
	assignments map[uuid.UUID][]models.TeachingAssignment
}

func newSubjectRepoMock() *subjectRepoMock {
	return &subjectRepoMock{
		subjects:    make(map[uuid.UUID]*models.Subject),
		assignments: make(map[uuid.UUID][]models.TeachingAssignment),
	}
}

func (r *subjectRepoMock) GetByID(_ context.Context, id uuid.UUID) (*models.Subject, error) {
	subject, ok := r.subjects[id]
	if !ok {
		return nil, errors.New("subject not found")
	}
	return subject, nil
}

func (r *subjectRepoMock) AssignTeacher(_ context.Context, assignment *models.TeachingAssignment) error {
	if assignment == nil {
		return errors.New("assignment nil")
	}
	r.assignments[assignment.SubjectID] = append(r.assignments[assignment.SubjectID], *assignment)
	return nil
}

func (r *subjectRepoMock) ListSubjectAssignments(_ context.Context, subjectID uuid.UUID) ([]models.TeachingAssignment, error) {
	assignments, ok := r.assignments[subjectID]
	if !ok {
		return []models.TeachingAssignment{}, nil
	}
	out := make([]models.TeachingAssignment, len(assignments))
	copy(out, assignments)
	return out, nil
}

func (r *subjectRepoMock) RemoveTeacherAssignment(_ context.Context, teacherID, subjectID uuid.UUID) error {
	assignments := r.assignments[subjectID]
	filtered := assignments[:0]
	for _, assignment := range assignments {
		if assignment.TeacherID != teacherID {
			filtered = append(filtered, assignment)
		}
	}
	r.assignments[subjectID] = append([]models.TeachingAssignment(nil), filtered...)
	return nil
}

type noopSessionRepo struct{}

func (noopSessionRepo) Create(context.Context, *models.ClassSession) error { return nil }
func (noopSessionRepo) ListBySubjectAndGroup(context.Context, uuid.UUID, uuid.UUID, *time.Time, *time.Time) ([]models.ClassSession, error) {
	return nil, nil
}
func (noopSessionRepo) GetByID(context.Context, uuid.UUID) (*models.ClassSession, error) {
	return nil, nil
}
func (noopSessionRepo) ListByStudent(context.Context, uuid.UUID) ([]models.ClassSession, error) {
	return nil, nil
}
func (noopSessionRepo) DeleteByTeacher(context.Context, uuid.UUID) error { return nil }
func (noopSessionRepo) DeleteByGroup(context.Context, uuid.UUID) error   { return nil }
func (noopSessionRepo) DeleteBySubject(context.Context, uuid.UUID) error { return nil }
func (noopSessionRepo) ListByTeacher(context.Context, uuid.UUID, *time.Time, *time.Time) ([]models.ClassSession, error) {
	return nil, nil
}
func (noopSessionRepo) ListByGroup(context.Context, uuid.UUID, *time.Time, *time.Time) ([]models.ClassSession, error) {
	return nil, nil
}
func (noopSessionRepo) ListByFilter(context.Context, repository.SessionFilter) ([]models.ClassSession, error) {
	return nil, nil
}

type noopGradeRepo struct{}

func (noopGradeRepo) GroupAverage(context.Context, uuid.UUID) (*float32, error) { return nil, nil }
func (noopGradeRepo) GroupSubjectAverage(context.Context, uuid.UUID, uuid.UUID) (*float32, error) {
	return nil, nil
}
func (noopGradeRepo) StudentOverallAverage(context.Context, uuid.UUID) (*float32, error) {
	return nil, nil
}
func (noopGradeRepo) StudentSubjectAverage(context.Context, uuid.UUID, uuid.UUID) (*float32, error) {
	return nil, nil
}
func (noopGradeRepo) ListBySubjectAndGroup(context.Context, uuid.UUID, uuid.UUID) ([]models.Grade, error) {
	return nil, nil
}
func (noopGradeRepo) ListByStudent(context.Context, uuid.UUID) ([]models.Grade, error) {
	return nil, nil
}
func (noopGradeRepo) GetByID(context.Context, uuid.UUID) (*models.Grade, error) { return nil, nil }
func (noopGradeRepo) GetBySessionAndStudent(context.Context, uuid.UUID, uuid.UUID) (*models.Grade, error) {
	return nil, nil
}
func (noopGradeRepo) Upsert(context.Context, *models.Grade) error      { return nil }
func (noopGradeRepo) Update(context.Context, *models.Grade) error      { return nil }
func (noopGradeRepo) Delete(context.Context, uuid.UUID) error          { return nil }
func (noopGradeRepo) DeleteByStudent(context.Context, uuid.UUID) error { return nil }
func (noopGradeRepo) DeleteByTeacher(context.Context, uuid.UUID) error { return nil }
func (noopGradeRepo) DeleteByGroup(context.Context, uuid.UUID) error   { return nil }
func (noopGradeRepo) DeleteBySubject(context.Context, uuid.UUID) error { return nil }

func TestDeanServiceAssignStudentToGroupBulk(t *testing.T) {
	userRepo := newDeanUserRepo()
	studentID := uuid.New()
	groupID := uuid.New()
	studentProfileID := uuid.New()
	userRepo.users[studentID] = &models.User{
		Base: models.Base{ID: studentID},
		Role: models.UserRoleStudent,
		Student: &models.StudentProfile{
			Base:   models.Base{ID: studentProfileID},
			UserID: studentID,
			Index:  "ST-01",
		},
	}

    service := NewDeanService(userRepo, noopGroupRepo{}, noopSubjectRepo{}, noopSessionRepo{}, noopGradeRepo{}, cache.NewNoop(), nil)
	payload := reqdto.AssignStudentToGroupRequest{StudentIDs: []string{studentID.String()}}
	if err := service.AssignStudentToGroup(context.Background(), groupID, payload); err != nil {
		t.Fatalf("assign students: %v", err)
	}
	profile := userRepo.students[studentID]
	if profile == nil || profile.GroupID == nil || *profile.GroupID != groupID {
		t.Fatalf("expected student to be assigned to group %s", groupID)
	}
}

func TestDeanServiceAssignTeacherValidatesRole(t *testing.T) {
	userRepo := newDeanUserRepo()
	subjectRepo := newSubjectRepoMock()
	subjectID := uuid.New()
	subjectRepo.subjects[subjectID] = &models.Subject{Base: models.Base{ID: subjectID}}

	teacherID := uuid.New()
	userRepo.users[teacherID] = &models.User{
		Base: models.Base{ID: teacherID},
		Role: models.UserRoleTeacher,
	}

    service := NewDeanService(userRepo, noopGroupRepo{}, subjectRepo, noopSessionRepo{}, noopGradeRepo{}, cache.NewNoop(), nil)
	if err := service.AssignTeacher(context.Background(), subjectID, reqdto.AssignTeacherRequest{TeacherID: teacherID.String()}); err != nil {
		t.Fatalf("expected teacher assignment to succeed, got %v", err)
	}
	if len(subjectRepo.assignments[subjectID]) != 1 {
		t.Fatalf("expected one assignment, got %d", len(subjectRepo.assignments[subjectID]))
	}

	studentID := uuid.New()
	userRepo.users[studentID] = &models.User{
		Base: models.Base{ID: studentID},
		Role: models.UserRoleStudent,
	}
	err := service.AssignTeacher(context.Background(), subjectID, reqdto.AssignTeacherRequest{TeacherID: studentID.String()})
	if err == nil || !strings.Contains(err.Error(), "selected user is not a teacher") {
		t.Fatalf("expected role validation error, got %v", err)
	}
}

func TestDeanServiceSubjectTeachersSorted(t *testing.T) {
	userRepo := newDeanUserRepo()
	subjectRepo := newSubjectRepoMock()
	subjectID := uuid.New()
	subjectRepo.subjects[subjectID] = &models.Subject{Base: models.Base{ID: subjectID}}

	teacherA := uuid.New()
	teacherB := uuid.New()
	userRepo.users[teacherA] = &models.User{
		Base:      models.Base{ID: teacherA},
		Role:      models.UserRoleTeacher,
		FirstName: "Irina",
		LastName:  "Petrova",
	}
	userRepo.users[teacherB] = &models.User{
		Base:      models.Base{ID: teacherB},
		Role:      models.UserRoleTeacher,
		FirstName: "Andrey",
		LastName:  "Alekseev",
	}

	subjectRepo.assignments[subjectID] = []models.TeachingAssignment{
		{Base: models.Base{ID: uuid.New()}, SubjectID: subjectID, TeacherID: teacherA},
		{Base: models.Base{ID: uuid.New()}, SubjectID: subjectID, TeacherID: teacherA},
		{Base: models.Base{ID: uuid.New()}, SubjectID: subjectID, TeacherID: teacherB},
	}

    service := NewDeanService(userRepo, noopGroupRepo{}, subjectRepo, noopSessionRepo{}, noopGradeRepo{}, cache.NewNoop(), nil)
	teachers, err := service.SubjectTeachers(context.Background(), subjectID)
	if err != nil {
		t.Fatalf("subject teachers: %v", err)
	}
	if len(teachers) != 2 {
		t.Fatalf("expected 2 distinct teachers, got %d", len(teachers))
	}
	if teachers[0].LastName != "Alekseev" || teachers[1].LastName != "Petrova" {
		t.Fatalf("expected alphabetical order, got %+v", teachers)
	}
}
