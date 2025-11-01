package service

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net/url"
    "sort"
    "strconv"
    "strings"
    "time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/cache"
	"gradeflow/pkg/logger"
	"gradeflow/pkg/utils"
)

// DeanService exposes dean office orchestration use-cases.
type DeanService struct {
    users    repository.UserRepository
    groups   repository.GroupRepository
    subjects repository.SubjectRepository
    sessions repository.SessionRepository
    grades   repository.GradeRepository
    cache    cache.Store
    profiles *ProfileService
}

// NewDeanService constructs the service.
func NewDeanService(users repository.UserRepository, groups repository.GroupRepository, subjects repository.SubjectRepository, sessions repository.SessionRepository, grades repository.GradeRepository, cacheStore cache.Store, profiles *ProfileService) *DeanService {
    if cacheStore == nil {
        cacheStore = cache.NewNoop()
    }
    return &DeanService{
        users:    users,
        groups:   groups,
        subjects: subjects,
        sessions: sessions,
        grades:   grades,
        cache:    cacheStore,
        profiles: profiles,
    }
}

func (s *DeanService) cacheKey(parts ...string) string {
	return strings.Join(append([]string{"dean"}, parts...), ":")
}

func deanSanitizeQuery(value *string) string {
	if value == nil {
		return ""
	}
	return url.QueryEscape(strings.TrimSpace(*value))
}

func (s *DeanService) invalidatePrefix(ctx context.Context, parts ...string) {
	prefix := s.cacheKey(parts...)
	if err := s.cache.InvalidatePrefix(ctx, prefix); err != nil {
		logger.Warn("cache invalidate failed", "prefix", prefix, "error", err)
	}
}

func uuidFromStringPtr(value *string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &id, nil
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
	s.invalidatePrefix(ctx, "groups")
	return &respdto.GroupSummary{ID: group.ID.String(), Name: group.Name, Description: group.Description}, nil
}

// ListGroups returns available groups.
func (s *DeanService) ListGroups(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.GroupSummary], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("groups", "list", strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), deanSanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.GroupSummary]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	listOpts := repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	}
	groups, total, err := s.groups.List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	items := make([]respdto.GroupSummary, 0, len(groups))
	for _, g := range groups {
		items = append(items, respdto.GroupSummary{ID: g.ID.String(), Name: g.Name, Description: g.Description})
	}
	result := &respdto.Paginated[respdto.GroupSummary]{
		Data: items,
		Meta: respdto.PageMeta{
			Limit:  opts.Limit,
			Offset: opts.Offset,
			Total:  int(total),
		},
	}
	if err := s.cache.Set(ctx, cacheKey, result, 0); err != nil {
		logger.Warn("cache set failed", "key", cacheKey, "error", err)
	}
	return result, nil
}

// UpdateGroup modifies group metadata.
func (s *DeanService) UpdateGroup(ctx context.Context, groupID uuid.UUID, payload reqdto.UpdateGroupRequest) (*respdto.GroupSummary, error) {
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("load group: %w", err)
	}
	if payload.Name != nil {
		group.Name = *payload.Name
	}
	if payload.Description != nil {
		group.Description = payload.Description
	}
	if err := s.groups.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	s.invalidatePrefix(ctx, "groups")
	return &respdto.GroupSummary{ID: group.ID.String(), Name: group.Name, Description: group.Description}, nil
}

// DeleteGroup soft deletes a group.
func (s *DeanService) DeleteGroup(ctx context.Context, groupID uuid.UUID) error {
	if s.sessions != nil {
		if err := s.sessions.DeleteByGroup(ctx, groupID); err != nil {
			return fmt.Errorf("delete group sessions: %w", err)
		}
	}
	if s.grades != nil {
		if err := s.grades.DeleteByGroup(ctx, groupID); err != nil {
			return fmt.Errorf("delete group grades: %w", err)
		}
	}
	if err := s.groups.SoftDelete(ctx, groupID); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "groups", "deleted")
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "subjects")
	return nil
}

// RestoreGroup restores a soft deleted group.
func (s *DeanService) RestoreGroup(ctx context.Context, groupID uuid.UUID) error {
	if err := s.groups.Restore(ctx, groupID); err != nil {
		return fmt.Errorf("restore group: %w", err)
	}
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "groups", "deleted")
	return nil
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
	s.invalidatePrefix(ctx, "subjects")
	return &respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description}, nil
}

// ListSubjects returns subjects.
func (s *DeanService) ListSubjects(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.SubjectSummary], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("subjects", "list", strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), deanSanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.SubjectSummary]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	listOpts := repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	}
	subjects, total, err := s.subjects.List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	items := make([]respdto.SubjectSummary, 0, len(subjects))
	for _, subj := range subjects {
		items = append(items, respdto.SubjectSummary{ID: subj.ID.String(), Code: subj.Code, Name: subj.Name, Description: subj.Description})
	}
	result := &respdto.Paginated[respdto.SubjectSummary]{
		Data: items,
		Meta: respdto.PageMeta{
			Limit:  opts.Limit,
			Offset: opts.Offset,
			Total:  int(total),
		},
	}
	if err := s.cache.Set(ctx, cacheKey, result, 0); err != nil {
		logger.Warn("cache set failed", "key", cacheKey, "error", err)
	}
	return result, nil
}

// UpdateSubject modifies subject metadata.
func (s *DeanService) UpdateSubject(ctx context.Context, subjectID uuid.UUID, payload reqdto.UpdateSubjectRequest) (*respdto.SubjectSummary, error) {
	subject, err := s.subjects.GetByID(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("load subject: %w", err)
	}
	if payload.Code != nil {
		subject.Code = *payload.Code
	}
	if payload.Name != nil {
		subject.Name = *payload.Name
	}
	if payload.Description != nil {
		subject.Description = payload.Description
	}
	if err := s.subjects.Update(ctx, subject); err != nil {
		return nil, fmt.Errorf("update subject: %w", err)
	}
	s.invalidatePrefix(ctx, "subjects")
	return &respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description}, nil
}

// DeleteSubject soft deletes a subject.
func (s *DeanService) DeleteSubject(ctx context.Context, subjectID uuid.UUID) error {
	if s.sessions != nil {
		if err := s.sessions.DeleteBySubject(ctx, subjectID); err != nil {
			return fmt.Errorf("delete subject sessions: %w", err)
		}
	}
	if s.grades != nil {
		if err := s.grades.DeleteBySubject(ctx, subjectID); err != nil {
			return fmt.Errorf("delete subject grades: %w", err)
		}
	}
	if err := s.subjects.SoftDelete(ctx, subjectID); err != nil {
		return fmt.Errorf("delete subject: %w", err)
	}
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "subjects", "deleted")
	return nil
}

// RestoreSubject restores a soft deleted subject.
func (s *DeanService) RestoreSubject(ctx context.Context, subjectID uuid.UUID) error {
	if err := s.subjects.Restore(ctx, subjectID); err != nil {
		return fmt.Errorf("restore subject: %w", err)
	}
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "subjects", "deleted")
	return nil
}

// CreateTeacher provisions teacher account.
func (s *DeanService) CreateTeacher(ctx context.Context, payload reqdto.CreateTeacherRequest) (*respdto.UserProfile, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	ins, err := s.users.NextINS(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate INS: %w", err)
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
		Base:   models.Base{ID: uuid.New()},
		UserID: user.ID,
		Title:  payload.Title,
		Bio:    payload.Bio,
	}); err != nil {
		return nil, fmt.Errorf("create teacher profile: %w", err)
	}
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "users")
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
func (s *DeanService) ListTeachers(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.UserProfile], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("teachers", "list", strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), deanSanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.UserProfile]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	listOpts := repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	}
	teachers, total, err := s.users.ListByRole(ctx, models.UserRoleTeacher, listOpts)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	items := make([]respdto.UserProfile, 0, len(teachers))
	for _, t := range teachers {
		items = append(items, respdto.UserProfile{
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
	result := &respdto.Paginated[respdto.UserProfile]{
		Data: items,
		Meta: respdto.PageMeta{
			Limit:  opts.Limit,
			Offset: opts.Offset,
			Total:  int(total),
		},
	}
	if err := s.cache.Set(ctx, cacheKey, result, 0); err != nil {
		logger.Warn("cache set failed", "key", cacheKey, "error", err)
	}
	return result, nil
}

// UpdateTeacher modifies teacher profile and metadata.
func (s *DeanService) UpdateTeacher(ctx context.Context, teacherID uuid.UUID, payload reqdto.UpdateTeacherRequest) (*respdto.UserProfile, error) {
	user, err := s.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("load teacher: %w", err)
	}
	if user.Role != models.UserRoleTeacher {
		return nil, errors.New("user is not a teacher")
	}
	if payload.Email != nil {
		user.Email = payload.Email
	}
	if payload.FirstName != nil {
		user.FirstName = *payload.FirstName
	}
	if payload.LastName != nil {
		user.LastName = *payload.LastName
	}
	if payload.MiddleName != nil {
		user.MiddleName = payload.MiddleName
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update teacher user: %w", err)
	}
	if payload.Title != nil || payload.Bio != nil {
		if err := s.users.AttachTeacherProfile(ctx, &models.TeacherProfile{
			UserID: user.ID,
			Title:  payload.Title,
			Bio:    payload.Bio,
		}); err != nil {
			return nil, fmt.Errorf("update teacher profile: %w", err)
		}
	}
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "users")
	profile := userProfileFromModel(user)
	return &profile, nil
}

// DeleteTeacher performs soft delete for teacher account.
func (s *DeanService) DeleteTeacher(ctx context.Context, teacherID uuid.UUID) error {
	return errors.New("удаление преподавателя доступно только администратору")
}

// CreateStudent provisions student account.
func (s *DeanService) CreateStudent(ctx context.Context, payload reqdto.CreateStudentRequest) (*respdto.UserProfile, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	ins, err := s.users.NextINS(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate INS: %w", err)
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
		Index:  fmt.Sprintf("ST-%s", ins),
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
	user.Student = profile
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "users")
	result := userProfileFromModel(user)
	return &result, nil
}

// ListStudents returns student summaries.
func (s *DeanService) ListStudents(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.UserProfile], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("students", "list", strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), deanSanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.UserProfile]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	listOpts := repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	}
	students, total, err := s.users.ListByRole(ctx, models.UserRoleStudent, listOpts)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	items := make([]respdto.UserProfile, 0, len(students))
	for idx := range students {
		profile := userProfileFromModel(&students[idx])
		items = append(items, profile)
	}
	result := &respdto.Paginated[respdto.UserProfile]{
		Data: items,
		Meta: respdto.PageMeta{
			Limit:  opts.Limit,
			Offset: opts.Offset,
			Total:  int(total),
		},
	}
	if err := s.cache.Set(ctx, cacheKey, result, 0); err != nil {
		logger.Warn("cache set failed", "key", cacheKey, "error", err)
	}
	return result, nil
}

// UpdateStudent modifies student profile attributes.
func (s *DeanService) UpdateStudent(ctx context.Context, studentID uuid.UUID, payload reqdto.UpdateStudentRequest) (*respdto.UserProfile, error) {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	if user.Role != models.UserRoleStudent {
		return nil, errors.New("user is not a student")
	}
	if payload.Email != nil {
		user.Email = payload.Email
	}
	if payload.FirstName != nil {
		user.FirstName = *payload.FirstName
	}
	if payload.LastName != nil {
		user.LastName = *payload.LastName
	}
	if payload.MiddleName != nil {
		user.MiddleName = payload.MiddleName
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update student user: %w", err)
	}

	if payload.GroupID != nil {
		profile := &models.StudentProfile{
			UserID: user.ID,
			Index:  studentIndexOrDefault(user),
		}
		if user.Student != nil {
			profile.ID = user.Student.ID
		}
		if strings.TrimSpace(*payload.GroupID) == "" {
			profile.GroupID = nil
		} else {
			groupID, err := uuid.Parse(*payload.GroupID)
			if err != nil {
				return nil, fmt.Errorf("parse groupId: %w", err)
			}
			profile.GroupID = &groupID
		}
		if err := s.users.AttachStudentProfile(ctx, profile); err != nil {
			return nil, fmt.Errorf("update student profile: %w", err)
		}
		if user.Student != nil {
			user.Student.GroupID = profile.GroupID
			user.Student.Index = profile.Index
		}
	}
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "users")
	profile := userProfileFromModel(user)
	return &profile, nil
}

// DeleteStudent performs soft delete for student account.
func (s *DeanService) DeleteStudent(ctx context.Context, studentID uuid.UUID) error {
	return errors.New("удаление студента доступно только администратору")
}

// AssignTeacher connects a teacher to a subject.
func (s *DeanService) AssignTeacher(ctx context.Context, subjectID uuid.UUID, payload reqdto.AssignTeacherRequest) error {
	teacherID, err := uuid.Parse(payload.TeacherID)
	if err != nil {
		return fmt.Errorf("parse teacher id: %w", err)
	}
	if _, err := s.subjects.GetByID(ctx, subjectID); err != nil {
		return fmt.Errorf("load subject: %w", err)
	}
	teacher, err := s.users.GetByID(ctx, teacherID)
	if err != nil {
		return fmt.Errorf("load teacher: %w", err)
	}
	if teacher.Role != models.UserRoleTeacher {
		return errors.New("selected user is not a teacher")
	}
	assignment := &models.TeachingAssignment{
		Base:      models.Base{ID: uuid.New()},
		SubjectID: subjectID,
		TeacherID: teacherID,
	}
	if err := s.subjects.AssignTeacher(ctx, assignment); err != nil {
		return fmt.Errorf("assign teacher: %w", err)
	}
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "subjects", subjectID.String(), "teachers")
	return nil
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
	if err := s.subjects.AttachGroup(ctx, link); err != nil {
		return err
	}
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "groups")
	return nil
}

// AssignStudentToGroup moves student into a group.
func (s *DeanService) AssignStudentToGroup(ctx context.Context, groupID uuid.UUID, payload reqdto.AssignStudentToGroupRequest) error {
	if len(payload.StudentIDs) == 0 {
		return errors.New("studentIds required")
	}
	seen := make(map[uuid.UUID]struct{})
	for _, id := range payload.StudentIDs {
		studentID, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("parse student id: %w", err)
		}
		if _, dup := seen[studentID]; dup {
			continue
		}
		seen[studentID] = struct{}{}

		user, err := s.users.GetByID(ctx, studentID)
		if err != nil {
			return fmt.Errorf("load student: %w", err)
		}
		if user.Student == nil {
			return errors.New("user has no student profile")
		}
		index := studentIndexOrDefault(user)
		profile := &models.StudentProfile{
			UserID:  user.ID,
			Index:   index,
			GroupID: &groupID,
		}
		if user.Student.ID != uuid.Nil {
			profile.ID = user.Student.ID
		}
		if err := s.users.AttachStudentProfile(ctx, profile); err != nil {
			return err
		}
		if user.Student != nil {
			user.Student.GroupID = &groupID
			user.Student.Index = profile.Index
		}
	}
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "users")
	return nil
}

// StudentSubjects returns subjects with grade history for given student.
func (s *DeanService) StudentSubjects(ctx context.Context, studentID uuid.UUID, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.StudentSubjectGrade], error) {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	grades, err := s.grades.ListByStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("list student grades: %w", err)
	}
	gradesBySubject := make(map[uuid.UUID][]models.Grade)
	for _, grade := range grades {
		gradesBySubject[grade.SubjectID] = append(gradesBySubject[grade.SubjectID], grade)
	}
	var subjects []models.Subject
	if user.Student != nil && user.Student.GroupID != nil {
		if list, err := s.subjects.ListByGroup(ctx, *user.Student.GroupID); err == nil {
			subjects = append(subjects, list...)
		}
	}
	// ensure subjects referenced by grades are present
	for subjectID := range gradesBySubject {
		found := false
		for _, subj := range subjects {
			if subj.ID == subjectID {
				found = true
				break
			}
		}
		if !found {
			if subj, err := s.subjects.GetByID(ctx, subjectID); err == nil {
				subjects = append(subjects, *subj)
			}
		}
	}
	response := make([]respdto.StudentSubjectGrade, 0, len(subjects))
	for _, subject := range subjects {
		var sessions []models.ClassSession
		if user.Student != nil && user.Student.GroupID != nil {
			if list, err := s.sessions.ListBySubjectAndGroup(ctx, subject.ID, *user.Student.GroupID, nil, nil); err == nil {
				sessions = append(sessions, list...)
			}
		}
		gradesForSubject := gradesBySubject[subject.ID]
		gradeMap := make(map[uuid.UUID]models.Grade)
		for _, grade := range gradesForSubject {
			gradeMap[grade.SessionID] = grade
		}
		if len(sessions) == 0 {
			for _, grade := range gradesForSubject {
				if session, err := s.sessions.GetByID(ctx, grade.SessionID); err == nil {
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
		subjectAverage, _ := s.grades.StudentSubjectAverage(ctx, studentID, subject.ID)
		response = append(response, respdto.StudentSubjectGrade{
			Subject:  respdto.SubjectSummary{ID: subject.ID.String(), Code: subject.Code, Name: subject.Name, Description: subject.Description},
			Sessions: sessionResponses,
			Average:  subjectAverage,
		})
	}

	filtered := response
	if opts.Search != nil && strings.TrimSpace(*opts.Search) != "" {
		needle := strings.ToLower(strings.TrimSpace(*opts.Search))
		tmp := make([]respdto.StudentSubjectGrade, 0, len(filtered))
		for _, subj := range filtered {
			if strings.Contains(strings.ToLower(subj.Subject.Name), needle) ||
				strings.Contains(strings.ToLower(subj.Subject.Code), needle) {
				tmp = append(tmp, subj)
			}
		}
		filtered = tmp
	}

	total := len(filtered)
	opts.Normalize(100)
	if total == 0 {
		return &respdto.Paginated[respdto.StudentSubjectGrade]{
			Data: []respdto.StudentSubjectGrade{},
			Meta: respdto.PageMeta{Limit: opts.Limit, Offset: 0, Total: 0},
		}, nil
	}
	if opts.Offset > total {
		opts.Offset = total
	}
	end := opts.Offset + opts.Limit
	if end > total {
		end = total
	}
	paged := filtered[opts.Offset:end]
	return &respdto.Paginated[respdto.StudentSubjectGrade]{
		Data: paged,
		Meta: respdto.PageMeta{Limit: opts.Limit, Offset: opts.Offset, Total: total},
	}, nil
}

// DetachStudentFromGroup removes student assignment from a group.
func (s *DeanService) DetachStudentFromGroup(ctx context.Context, groupID, studentID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return fmt.Errorf("load student: %w", err)
	}
	if user.Student == nil {
		return errors.New("user has no student profile")
	}
	profile := &models.StudentProfile{
		UserID:  user.ID,
		Index:   studentIndexOrDefault(user),
		GroupID: nil,
	}
	if user.Student.ID != uuid.Nil {
		profile.ID = user.Student.ID
	}
	if err := s.users.AttachStudentProfile(ctx, profile); err != nil {
		return fmt.Errorf("detach student profile: %w", err)
	}
	if user.Student != nil {
		user.Student.GroupID = nil
		user.Student.Index = profile.Index
	}
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "users")
	return nil
}

// UpdateGrade allows dean staff to adjust any grade.
func (s *DeanService) UpdateGrade(ctx context.Context, gradeID uuid.UUID, payload reqdto.UpdateGradeRequest) (*respdto.GradeDetail, error) {
	if s.grades == nil {
		return nil, errors.New("grade repository unavailable")
	}
	grade, err := s.grades.GetByID(ctx, gradeID)
	if err != nil {
		return nil, fmt.Errorf("load grade: %w", err)
	}
	value := payload.Value
	if !isGradeValueAllowed(value) {
		return nil, errors.New("grade value must be one of 2, 3, 4, 5")
	}
	grade.Value = value
	grade.Notes = payload.Notes
	assessedAt := time.Now()
	grade.AssessedAt = assessedAt
	if err := s.grades.Update(ctx, grade); err != nil {
		return nil, fmt.Errorf("update grade: %w", err)
	}
	student, err := s.users.GetByID(ctx, grade.StudentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	value = grade.Value
	profile := userProfileFromModel(student)
	return &respdto.GradeDetail{
		GradeID:    utils.StringPtr(grade.ID.String()),
		SessionID:  grade.SessionID.String(),
		Student:    profile,
		Value:      &value,
		Notes:      grade.Notes,
		AssessedAt: &grade.AssessedAt,
	}, nil
}

// DetachTeacherFromSubject removes teacher assignment from subject.
func (s *DeanService) DetachTeacherFromSubject(ctx context.Context, subjectID, teacherID uuid.UUID) error {
	if err := s.subjects.RemoveTeacherAssignment(ctx, teacherID, subjectID); err != nil {
		return fmt.Errorf("remove teacher assignment: %w", err)
	}
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "subjects", subjectID.String(), "teachers")
	return nil
}

// UpdateStudentAvatar lets dean staff update student photo.
func (s *DeanService) UpdateStudentAvatar(ctx context.Context, studentID uuid.UUID, data io.Reader) (*respdto.UserProfile, error) {
	if s.profiles == nil {
		return nil, ErrAvatarNotConfigured
	}
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	if user.Role != models.UserRoleStudent {
		return nil, errors.New("user is not a student")
	}
	profile, err := s.profiles.UploadAvatarFor(ctx, studentID, data)
	if err != nil {
		return nil, err
	}
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "groups")
	return profile, nil
}

// DeleteStudentAvatar removes student photo.
func (s *DeanService) DeleteStudentAvatar(ctx context.Context, studentID uuid.UUID) (*respdto.UserProfile, error) {
	if s.profiles == nil {
		return nil, ErrAvatarNotConfigured
	}
	user, err := s.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}
	if user.Role != models.UserRoleStudent {
		return nil, errors.New("user is not a student")
	}
	profile, err := s.profiles.DeleteAvatarFor(ctx, studentID)
	if err != nil {
		return nil, err
	}
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "groups")
	return profile, nil
}

// UpdateTeacherAvatar lets dean staff update teacher photo.
func (s *DeanService) UpdateTeacherAvatar(ctx context.Context, teacherID uuid.UUID, data io.Reader) (*respdto.UserProfile, error) {
	if s.profiles == nil {
		return nil, ErrAvatarNotConfigured
	}
	user, err := s.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("load teacher: %w", err)
	}
	if user.Role != models.UserRoleTeacher {
		return nil, errors.New("user is not a teacher")
	}
	profile, err := s.profiles.UploadAvatarFor(ctx, teacherID, data)
	if err != nil {
		return nil, err
	}
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "subjects")
	return profile, nil
}

// DeleteTeacherAvatar removes teacher photo.
func (s *DeanService) DeleteTeacherAvatar(ctx context.Context, teacherID uuid.UUID) (*respdto.UserProfile, error) {
	if s.profiles == nil {
		return nil, ErrAvatarNotConfigured
	}
	user, err := s.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("load teacher: %w", err)
	}
	if user.Role != models.UserRoleTeacher {
		return nil, errors.New("user is not a teacher")
	}
	profile, err := s.profiles.DeleteAvatarFor(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "subjects")
	return profile, nil
}

// SubjectTeachers lists teachers assigned to a subject.
func (s *DeanService) SubjectTeachers(ctx context.Context, subjectID uuid.UUID) ([]respdto.UserProfile, error) {
	cacheKey := s.cacheKey("subjects", subjectID.String(), "teachers")
	var cached []respdto.UserProfile
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	assignments, err := s.subjects.ListSubjectAssignments(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list subject assignments: %w", err)
	}
	if len(assignments) == 0 {
		return []respdto.UserProfile{}, nil
	}
	unique := make(map[uuid.UUID]struct{}, len(assignments))
	for _, a := range assignments {
		unique[a.TeacherID] = struct{}{}
	}
	result := make([]respdto.UserProfile, 0, len(unique))
	for teacherID := range unique {
		user, err := s.users.GetByID(ctx, teacherID)
		if err != nil {
			return nil, fmt.Errorf("load teacher: %w", err)
		}
		result = append(result, userProfileFromModel(user))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].LastName == result[j].LastName {
			return result[i].FirstName < result[j].FirstName
		}
		return result[i].LastName < result[j].LastName
	})
	if err := s.cache.Set(ctx, cacheKey, result, 5*time.Minute); err != nil {
		logger.Warn("cache set failed", "key", cacheKey, "error", err)
	}
	return result, nil
}

// ScheduleSession creates lesson slots for one subject and multiple groups.
func (s *DeanService) ScheduleSession(ctx context.Context, payload reqdto.ScheduleSessionRequest) ([]respdto.SessionSummary, error) {
	subjectID, err := uuid.Parse(payload.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("parse subject id: %w", err)
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
	if len(payload.GroupIDs) == 0 {
		return nil, errors.New("groupIds required")
	}
	if payload.Slot < 1 || payload.Slot > 6 {
		return nil, errors.New("slot must be between 1 and 6")
	}
	slotDuration := 90 * time.Minute
	sessionDate, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}
	startBase := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(), 9, 0, 0, 0, time.Local)
	startsAt := startBase.Add(time.Duration(payload.Slot-1) * slotDuration)
	endsAt := startsAt.Add(slotDuration)

	summaries := make([]respdto.SessionSummary, 0, len(payload.GroupIDs))
	for _, gidStr := range payload.GroupIDs {
		groupID, err := uuid.Parse(gidStr)
		if err != nil {
			return nil, fmt.Errorf("parse group id: %w", err)
		}
		end := endsAt
		session := &models.ClassSession{
			Base:      models.Base{ID: uuid.New()},
			SubjectID: subjectID,
			GroupID:   groupID,
			TeacherID: teacherID,
			StartsAt:  startsAt,
			EndsAt:    &end,
			Topic:     payload.Topic,
		}
		if err := s.sessions.Create(ctx, session); err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
		summaries = append(summaries, respdto.SessionSummary{
			ID:        session.ID.String(),
			StartsAt:  session.StartsAt,
			EndsAt:    session.EndsAt,
			Topic:     session.Topic,
			SubjectID: subjectID.String(),
			GroupID:   session.GroupID.String(),
		})
	}
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "teachers")
	return summaries, nil
}

// Schedule returns sessions filtered by subject, group or teacher.
func (s *DeanService) Schedule(ctx context.Context, query reqdto.ScheduleQuery) ([]respdto.ScheduleEntry, error) {
	if s.sessions == nil {
		return []respdto.ScheduleEntry{}, nil
	}
	filter := repository.SessionFilter{}
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
	if id, err := uuidFromStringPtr(query.TeacherID); err != nil {
		return nil, fmt.Errorf("parse teacherId: %w", err)
	} else if id != nil {
		filter.TeacherID = id
	}
	filter.From = query.From
	filter.To = query.To

	sessions, err := s.sessions.ListByFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return buildScheduleEntries(sessions), nil
}

// GroupRanking composes ranking by average grade.
func (s *DeanService) GroupRanking(ctx context.Context) (*respdto.GroupRankingResponse, error) {
	groups, _, err := s.groups.List(ctx, repository.ListOptions{})
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
