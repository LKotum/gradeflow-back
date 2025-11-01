package service

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net/url"
    "strconv"
    "strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/cache"
	"gradeflow/pkg/logger"
)

// AdminService handles administrator operations such as provisioning dean staff
// and restoring soft-deleted records.
type AdminService struct {
    users    repository.UserRepository
    groups   repository.GroupRepository
    subjects repository.SubjectRepository
    sessions repository.SessionRepository
    grades   repository.GradeRepository
    cache    cache.Store
    profiles *ProfileService
}

// NewAdminService creates the service.
func NewAdminService(
    users repository.UserRepository,
    groups repository.GroupRepository,
    subjects repository.SubjectRepository,
    sessions repository.SessionRepository,
    grades repository.GradeRepository,
    cacheStore cache.Store,
    profiles *ProfileService,
) *AdminService {
    if cacheStore == nil {
        cacheStore = cache.NewNoop()
    }
    return &AdminService{
        users:    users,
        groups:   groups,
        subjects: subjects,
        sessions: sessions,
        grades:   grades,
        cache:    cacheStore,
        profiles: profiles,
    }
}

func (s *AdminService) cacheKey(parts ...string) string {
	return strings.Join(append([]string{"admin"}, parts...), ":")
}

func sanitizeQuery(value *string) string {
	if value == nil {
		return ""
	}
	return url.QueryEscape(strings.TrimSpace(*value))
}

func (s *AdminService) invalidatePrefix(ctx context.Context, parts ...string) {
	prefix := s.cacheKey(parts...)
	if err := s.cache.InvalidatePrefix(ctx, prefix); err != nil {
		logger.Warn("cache invalidate failed", "prefix", prefix, "error", err)
	}
}

func (s *AdminService) userProfilesFromModels(users []models.User) []respdto.UserProfile {
	items := make([]respdto.UserProfile, 0, len(users))
	for _, u := range users {
		user := u
		items = append(items, userProfileFromModel(&user))
	}
	return items
}

// CreateDean provisions a dean office staff user.
func (s *AdminService) CreateDean(ctx context.Context, payload reqdto.CreateDeanRequest) (*respdto.UserProfile, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	ins, err := s.users.NextINS(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate INS: %w", err)
	}
	user := &models.User{
		Base:         models.Base{ID: uuid.New()},
		Role:         models.UserRoleDean,
		INS:          &ins,
		Email:        payload.Email,
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		MiddleName:   payload.MiddleName,
		PasswordHash: string(hashed),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create dean user: %w", err)
	}
	if err := s.users.AttachStaffProfile(ctx, &models.StaffProfile{
		Base:     models.Base{ID: uuid.New()},
		UserID:   user.ID,
		Position: payload.Position,
	}); err != nil {
		return nil, fmt.Errorf("create staff profile: %w", err)
	}
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "users")
	profile := userProfileFromModel(user)
	return &profile, nil
}

// ListDeans returns dean staff profiles with pagination.
func (s *AdminService) ListDeans(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.UserProfile], error) {
	return s.listUsersByRole(ctx, []string{"deans", "list"}, models.UserRoleDean, opts)
}

// UpdateDean updates dean staff profile.
func (s *AdminService) UpdateDean(ctx context.Context, deanID uuid.UUID, payload reqdto.UpdateDeanRequest) (*respdto.UserProfile, error) {
	user, err := s.users.GetByID(ctx, deanID)
	if err != nil {
		return nil, fmt.Errorf("load dean: %w", err)
	}
	if user.Role != models.UserRoleDean {
		return nil, fmt.Errorf("user is not dean")
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
		return nil, fmt.Errorf("update dean user: %w", err)
	}
	if payload.Position != nil {
		if err := s.users.AttachStaffProfile(ctx, &models.StaffProfile{
			UserID:   user.ID,
			Position: payload.Position,
		}); err != nil {
			return nil, fmt.Errorf("update staff profile: %w", err)
		}
	}
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "users")
	profile := userProfileFromModel(user)
	return &profile, nil
}

// DeleteDean performs soft delete for dean user.
func (s *AdminService) DeleteDean(ctx context.Context, deanID uuid.UUID) error {
	if err := s.users.SoftDelete(ctx, deanID); err != nil {
		return fmt.Errorf("soft delete dean: %w", err)
	}
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "users")
	s.invalidatePrefix(ctx, "users", "deleted")
	return nil
}

// UpdateUser updates fields for teacher/student/dean from admin panel.
func (s *AdminService) UpdateUser(ctx context.Context, userID uuid.UUID, payload reqdto.UpdateUserRequest) (*respdto.UserProfile, error) {
 user, err := s.users.GetByID(ctx, userID)
 if err != nil {
  return nil, fmt.Errorf("load user: %w", err)
 }
 normalize := func(value *string) *string {
  if value == nil {
   return nil
  }
  trimmed := strings.TrimSpace(*value)
  if trimmed == "" {
   return nil
  }
  return &trimmed
 }
 if payload.FirstName != nil {
  val := strings.TrimSpace(*payload.FirstName)
  if val == "" {
   return nil, errors.New("firstName cannot be empty")
  }
  user.FirstName = val
 }
 if payload.LastName != nil {
  val := strings.TrimSpace(*payload.LastName)
  if val == "" {
   return nil, errors.New("lastName cannot be empty")
  }
  user.LastName = val
 }
 if payload.MiddleName != nil {
  user.MiddleName = normalize(payload.MiddleName)
 }
 if payload.Email != nil {
  user.Email = normalize(payload.Email)
 }
 if err := s.users.Update(ctx, user); err != nil {
  return nil, fmt.Errorf("update user: %w", err)
 }
 switch user.Role {
 case models.UserRoleTeacher:
  if payload.Title != nil || payload.Bio != nil {
   profile := &models.TeacherProfile{UserID: user.ID}
   if payload.Title != nil {
    profile.Title = normalize(payload.Title)
   }
   if payload.Bio != nil {
    profile.Bio = normalize(payload.Bio)
   }
   if err := s.users.AttachTeacherProfile(ctx, profile); err != nil {
    return nil, fmt.Errorf("update teacher profile: %w", err)
   }
   if user.Teacher == nil {
    user.Teacher = &models.TeacherProfile{}
   }
   user.Teacher.Title = profile.Title
   user.Teacher.Bio = profile.Bio
  }
 case models.UserRoleDean:
  if payload.Position != nil {
   profile := &models.StaffProfile{UserID: user.ID, Position: normalize(payload.Position)}
   if err := s.users.AttachStaffProfile(ctx, profile); err != nil {
    return nil, fmt.Errorf("update staff profile: %w", err)
   }
   if user.Staff == nil {
    user.Staff = &models.StaffProfile{}
   }
   user.Staff.Position = profile.Position
  }
 }
 s.invalidatePrefix(ctx, "users")
 if user.Role == models.UserRoleTeacher {
  s.invalidatePrefix(ctx, "teachers")
 }
 if user.Role == models.UserRoleStudent {
  s.invalidatePrefix(ctx, "students")
 }
 if user.Role == models.UserRoleDean {
  s.invalidatePrefix(ctx, "deans")
 }
 profile := userProfileFromModel(user)
 return &profile, nil
}

// RestoreDean removes soft delete mark from dean user.
func (s *AdminService) RestoreDean(ctx context.Context, deanID uuid.UUID) error {
	if err := s.users.Restore(ctx, deanID); err != nil {
		return fmt.Errorf("restore dean: %w", err)
	}
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "users")
	return nil
}

// ListDeletedUsers returns soft-deleted users by role.
func (s *AdminService) ListDeletedUsers(ctx context.Context, role models.UserRole, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.UserProfile], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("users", "deleted", string(role), strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), sanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.UserProfile]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	users, total, err := s.users.ListDeletedByRole(ctx, role, repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	})
	if err != nil {
		return nil, fmt.Errorf("list deleted users: %w", err)
	}
	result := &respdto.Paginated[respdto.UserProfile]{
		Data: s.userProfilesFromModels(users),
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

// RestoreUser clears soft-delete flag for specified user.
func (s *AdminService) RestoreUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.users.Restore(ctx, userID); err != nil {
		return fmt.Errorf("restore user: %w", err)
	}
	s.invalidatePrefix(ctx, "users")
	s.invalidatePrefix(ctx, "users", "deleted")
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "students")
	return nil
}

// ListDeletedGroups returns soft-deleted groups.
func (s *AdminService) ListDeletedGroups(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.GroupSummary], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("groups", "deleted", strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), sanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.GroupSummary]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	groups, total, err := s.groups.ListDeleted(ctx, repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	})
	if err != nil {
		return nil, fmt.Errorf("list deleted groups: %w", err)
	}
	items := make([]respdto.GroupSummary, 0, len(groups))
	for _, g := range groups {
		items = append(items, respdto.GroupSummary{
			ID:          g.ID.String(),
			Name:        g.Name,
			Description: g.Description,
		})
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

// RestoreGroup restores a soft-deleted group.
func (s *AdminService) RestoreGroup(ctx context.Context, groupID uuid.UUID) error {
	if err := s.groups.Restore(ctx, groupID); err != nil {
		return fmt.Errorf("restore group: %w", err)
	}
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "groups", "deleted")
	return nil
}

// ListDeletedSubjects returns soft-deleted subjects.
func (s *AdminService) ListDeletedSubjects(ctx context.Context, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.SubjectSummary], error) {
	opts.Normalize(100)
	cacheKey := s.cacheKey("subjects", "deleted", strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), sanitizeQuery(opts.Search))
	var cached respdto.Paginated[respdto.SubjectSummary]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	subjects, total, err := s.subjects.ListDeleted(ctx, repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	})
	if err != nil {
		return nil, fmt.Errorf("list deleted subjects: %w", err)
	}
	items := make([]respdto.SubjectSummary, 0, len(subjects))
	for _, subj := range subjects {
		items = append(items, respdto.SubjectSummary{
			ID:          subj.ID.String(),
			Code:        subj.Code,
			Name:        subj.Name,
			Description: subj.Description,
		})
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

// RestoreSubject restores a soft-deleted subject.
func (s *AdminService) RestoreSubject(ctx context.Context, subjectID uuid.UUID) error {
	if err := s.subjects.Restore(ctx, subjectID); err != nil {
		return fmt.Errorf("restore subject: %w", err)
	}
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "subjects", "deleted")
	return nil
}

// ListUsers returns active users filtered by role.
func (s *AdminService) ListUsers(ctx context.Context, role models.UserRole, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.UserProfile], error) {
	switch role {
	case models.UserRoleDean, models.UserRoleTeacher, models.UserRoleStudent:
	default:
		return nil, fmt.Errorf("unsupported role: %s", role)
	}
	return s.listUsersByRole(ctx, []string{"users", "active"}, role, opts)
}

// DeleteUser performs soft delete for any non-admin user.
func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("load user: %w", err)
	}
	if user.Role == models.UserRoleAdmin {
		return errors.New("нельзя удалить системного администратора")
	}

	switch user.Role {
	case models.UserRoleStudent:
		if user.Student != nil {
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
		}
		if s.grades != nil {
			if err := s.grades.DeleteByStudent(ctx, user.ID); err != nil {
				return fmt.Errorf("delete student grades: %w", err)
			}
		}
	case models.UserRoleTeacher:
		if s.subjects != nil {
			if err := s.subjects.RemoveTeacherAssignments(ctx, user.ID); err != nil {
				return fmt.Errorf("remove teacher assignments: %w", err)
			}
		}
		if s.sessions != nil {
			if err := s.sessions.DeleteByTeacher(ctx, user.ID); err != nil {
				return fmt.Errorf("delete teacher sessions: %w", err)
			}
		}
		if s.grades != nil {
			if err := s.grades.DeleteByTeacher(ctx, user.ID); err != nil {
				return fmt.Errorf("delete teacher grades: %w", err)
			}
		}
	}

	if err := s.users.SoftDelete(ctx, userID); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if err := s.users.DeleteRefreshToken(ctx, userID); err != nil {
		return fmt.Errorf("clear refresh token: %w", err)
	}
	s.invalidatePrefix(ctx, "users")
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "students")
	s.invalidatePrefix(ctx, "subjects")
	s.invalidatePrefix(ctx, "groups")
	s.invalidatePrefix(ctx, "users", "deleted")
	return nil
}

// ResetPassword allows administrator to set a new password for any user.
func (s *AdminService) ResetPassword(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("load user: %w", err)
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hashed)
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := s.users.DeleteRefreshToken(ctx, userID); err != nil {
		return fmt.Errorf("clear refresh token: %w", err)
	}
	s.invalidatePrefix(ctx, "users")
	return nil
}

// UpdateUserAvatar stores a new avatar for any user.
func (s *AdminService) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, data io.Reader) (*respdto.UserProfile, error) {
	if s.profiles == nil {
		return nil, ErrAvatarNotConfigured
	}
	profile, err := s.profiles.UploadAvatarFor(ctx, userID, data)
	if err != nil {
		return nil, err
	}
	s.invalidatePrefix(ctx, "users")
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "students")
	return profile, nil
}

// DeleteUserAvatar removes avatar for the specified user.
func (s *AdminService) DeleteUserAvatar(ctx context.Context, userID uuid.UUID) (*respdto.UserProfile, error) {
	if s.profiles == nil {
		return nil, ErrAvatarNotConfigured
	}
	profile, err := s.profiles.DeleteAvatarFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.invalidatePrefix(ctx, "users")
	s.invalidatePrefix(ctx, "deans")
	s.invalidatePrefix(ctx, "teachers")
	s.invalidatePrefix(ctx, "students")
	return profile, nil
}

func (s *AdminService) listUsersByRole(ctx context.Context, scope []string, role models.UserRole, opts reqdto.PaginationQuery) (*respdto.Paginated[respdto.UserProfile], error) {
	opts.Normalize(100)
	parts := append(scope, string(role), strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset), sanitizeQuery(opts.Search))
	cacheKey := s.cacheKey(parts...)
	var cached respdto.Paginated[respdto.UserProfile]
	if ok, err := s.cache.Get(ctx, cacheKey, &cached); err == nil && ok {
		return &cached, nil
	} else if err != nil {
		logger.Warn("cache get failed", "key", cacheKey, "error", err)
	}
	users, total, err := s.users.ListByRole(ctx, role, repository.ListOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Search: opts.Search,
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	result := &respdto.Paginated[respdto.UserProfile]{
		Data: s.userProfilesFromModels(users),
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
