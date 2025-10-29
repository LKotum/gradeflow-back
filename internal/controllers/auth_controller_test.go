package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/middleware"
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
)

type controllerUserRepo struct {
	users         map[uuid.UUID]*models.User
	byINS         map[string]uuid.UUID
	refreshTokens map[uuid.UUID]*models.RefreshToken
	nextINS       int
}

var errNotFound = errors.New("not found")

func newControllerUserRepo() *controllerUserRepo {
	return &controllerUserRepo{
		users:         make(map[uuid.UUID]*models.User),
		byINS:         make(map[string]uuid.UUID),
		refreshTokens: make(map[uuid.UUID]*models.RefreshToken),
		nextINS:       1,
	}
}

func (r *controllerUserRepo) Create(_ context.Context, user *models.User) error {
	r.users[user.ID] = user
	if user.INS != nil {
		r.byINS[*user.INS] = user.ID
	}
	return nil
}

func (r *controllerUserRepo) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errNotFound
	}
	if token, ok := r.refreshTokens[id]; ok {
		user.RefreshToken = token
	}
	return user, nil
}

func (r *controllerUserRepo) GetByINS(_ context.Context, ins string) (*models.User, error) {
	id, ok := r.byINS[ins]
	if !ok {
		return nil, errNotFound
	}
	return r.GetByID(context.Background(), id)
}

func (r *controllerUserRepo) ListByRole(_ context.Context, role models.UserRole, _ repository.ListOptions) ([]models.User, int64, error) {
	var data []models.User
	for _, u := range r.users {
		if u.Role == role {
			data = append(data, *u)
		}
	}
	return data, int64(len(data)), nil
}

func (r *controllerUserRepo) ListDeletedByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *controllerUserRepo) Update(context.Context, *models.User) error { return nil }

func (r *controllerUserRepo) AttachStudentProfile(context.Context, *models.StudentProfile) error {
	return nil
}

func (r *controllerUserRepo) AttachTeacherProfile(context.Context, *models.TeacherProfile) error {
	return nil
}

func (r *controllerUserRepo) AttachStaffProfile(context.Context, *models.StaffProfile) error {
	return nil
}

func (r *controllerUserRepo) UpsertRefreshToken(_ context.Context, token *models.RefreshToken) error {
	r.refreshTokens[token.UserID] = token
	return nil
}

func (r *controllerUserRepo) SoftDelete(context.Context, uuid.UUID) error { return nil }

func (r *controllerUserRepo) Restore(context.Context, uuid.UUID) error { return nil }

func (r *controllerUserRepo) DeleteRefreshToken(_ context.Context, userID uuid.UUID) error {
	delete(r.refreshTokens, userID)
	return nil
}

func (r *controllerUserRepo) NextINS(context.Context) (string, error) {
	value := r.nextINS
	r.nextINS++
	return fmt.Sprintf("%08d", value), nil
}

var _ repository.UserRepository = (*controllerUserRepo)(nil)

func TestAuthControllerLoginByINS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newControllerUserRepo()
	svc := service.NewAuthService(repo, "secret", time.Minute, time.Hour)
	ins := "INS-CTRL-1"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		Base:         models.Base{ID: uuid.New()},
		Role:         models.UserRoleTeacher,
		INS:          &ins,
		PasswordHash: string(hash),
		FirstName:    "Lily",
		LastName:     "Login",
	}
	repo.Create(context.Background(), user)

	ctrl := NewAuthController(svc, repo)
	router := gin.New()
	ctrl.RegisterPublicRoutes(router.Group("/auth"))

	body, _ := json.Marshal(reqdto.INSLoginRequest{INS: ins, Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login/ins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["accessToken"] == "" {
		t.Fatal("expected accessToken in response")
	}
}

func TestAuthControllerChangePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newControllerUserRepo()
	svc := service.NewAuthService(repo, "secret", time.Minute, time.Hour)
	ins := "00009999"
	oldPassword := "OldPass123!"
	hash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	user := &models.User{
		Base:         models.Base{ID: uuid.New()},
		Role:         models.UserRoleTeacher,
		INS:          &ins,
		PasswordHash: string(hash),
		FirstName:    "Paula",
		LastName:     "Patch",
	}
	repo.Create(context.Background(), user)

	ctrl := NewAuthController(svc, repo)
	router := gin.New()
	private := router.Group("/auth")
	private.Use(func(ctx *gin.Context) {
		ctx.Set(middleware.ContextUserIDKey, user.ID.String())
	})
	ctrl.RegisterPrivateRoutes(private)

	body, _ := json.Marshal(reqdto.ChangePasswordRequest{CurrentPassword: oldPassword, NewPassword: "NewPass456!"})
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d (%s)", resp.Code, resp.Body.String())
	}
	if _, err := svc.LoginByINS(context.Background(), reqdto.INSLoginRequest{INS: ins, Password: "NewPass456!"}); err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
}

func TestAuthControllerChangePasswordInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newControllerUserRepo()
	svc := service.NewAuthService(repo, "secret", time.Minute, time.Hour)
	ins := "00009998"
	hash, _ := bcrypt.GenerateFromPassword([]byte("ValidPass1!"), bcrypt.DefaultCost)
	user := &models.User{
		Base:         models.Base{ID: uuid.New()},
		Role:         models.UserRoleStudent,
		INS:          &ins,
		PasswordHash: string(hash),
	}
	repo.Create(context.Background(), user)

	ctrl := NewAuthController(svc, repo)
	router := gin.New()
	private := router.Group("/auth")
	private.Use(func(ctx *gin.Context) {
		ctx.Set(middleware.ContextUserIDKey, user.ID.String())
	})
	ctrl.RegisterPrivateRoutes(private)

	body, _ := json.Marshal(reqdto.ChangePasswordRequest{CurrentPassword: "WrongPass", NewPassword: "NewPass123!"})
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}
