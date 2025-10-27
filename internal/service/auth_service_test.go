package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
)

var errNotFound = errors.New("not found")

type fakeUserRepo struct {
	users         map[uuid.UUID]*models.User
	byINS         map[string]uuid.UUID
	byUsername    map[string]uuid.UUID
	profiles      map[uuid.UUID]*models.StudentProfile
	refreshTokens map[uuid.UUID]*models.RefreshToken
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:         make(map[uuid.UUID]*models.User),
		byINS:         make(map[string]uuid.UUID),
		byUsername:    make(map[string]uuid.UUID),
		profiles:      make(map[uuid.UUID]*models.StudentProfile),
		refreshTokens: make(map[uuid.UUID]*models.RefreshToken),
	}
}

func (f *fakeUserRepo) Create(_ context.Context, user *models.User) error {
	f.users[user.ID] = user
	if user.INS != nil {
		f.byINS[*user.INS] = user.ID
	}
	if user.Username != nil {
		f.byUsername[*user.Username] = user.ID
	}
	return nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, errNotFound
	}
	if token, ok := f.refreshTokens[id]; ok {
		user.RefreshToken = token
	}
	return user, nil
}

func (f *fakeUserRepo) GetByINS(_ context.Context, ins string) (*models.User, error) {
	id, ok := f.byINS[ins]
	if !ok {
		return nil, errNotFound
	}
	return f.GetByID(context.Background(), id)
}

func (f *fakeUserRepo) GetByUsername(_ context.Context, username string) (*models.User, error) {
	id, ok := f.byUsername[username]
	if !ok {
		return nil, errNotFound
	}
	return f.GetByID(context.Background(), id)
}

func (f *fakeUserRepo) ListByRole(_ context.Context, role models.UserRole) ([]models.User, error) {
	var result []models.User
	for _, u := range f.users {
		if u.Role == role {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (f *fakeUserRepo) Update(_ context.Context, _ *models.User) error { return nil }

func (f *fakeUserRepo) AttachStudentProfile(_ context.Context, profile *models.StudentProfile) error {
	f.profiles[profile.UserID] = profile
	return nil
}

func (f *fakeUserRepo) AttachTeacherProfile(_ context.Context, _ *models.TeacherProfile) error { return nil }

func (f *fakeUserRepo) AttachStaffProfile(_ context.Context, _ *models.StaffProfile) error { return nil }

func (f *fakeUserRepo) UpsertRefreshToken(_ context.Context, token *models.RefreshToken) error {
	f.refreshTokens[token.UserID] = token
	return nil
}

func TestAuthServiceLoginByINS(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "secret", time.Minute, time.Hour)
	ins := "INS-001"
	password := "Password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:           uuid.New(),
		Role:         models.UserRoleTeacher,
		INS:          &ins,
		PasswordHash: string(hash),
		FirstName:    "Tom",
		LastName:     "Teacher",
	}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	resp, err := svc.LoginByINS(context.Background(), reqdto.INSLoginRequest{INS: ins, Password: password})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestAuthServiceLoginAdminInvalidPassword(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "secret", time.Minute, time.Hour)
	username := "admin"
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	user := &models.User{
		ID:           uuid.New(),
		Role:         models.UserRoleAdmin,
		Username:     &username,
		PasswordHash: string(hash),
		FirstName:    "Alice",
		LastName:     "Admin",
	}
	repo.Create(context.Background(), user)
	_, err := svc.LoginAdmin(context.Background(), reqdto.AdminLoginRequest{Username: username, Password: "wrong"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceRefresh(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, "secret", time.Minute, time.Hour)
	ins := "INS-002"
	hash, _ := bcrypt.GenerateFromPassword([]byte("refreshPass"), bcrypt.DefaultCost)
	user := &models.User{
		ID:           uuid.New(),
		Role:         models.UserRoleStudent,
		INS:          &ins,
		PasswordHash: string(hash),
		FirstName:    "Sam",
		LastName:     "Student",
	}
	repo.Create(context.Background(), user)
	loginResp, err := svc.LoginByINS(context.Background(), reqdto.INSLoginRequest{INS: ins, Password: "refreshPass"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	refreshResp, err := svc.Refresh(context.Background(), reqdto.RefreshTokenRequest{RefreshToken: loginResp.RefreshToken})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshResp.AccessToken == loginResp.AccessToken {
		t.Fatal("expected new access token to differ")
	}
}
