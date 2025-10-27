package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
)

type controllerUserRepo struct {
	users         map[uuid.UUID]*models.User
	byINS         map[string]uuid.UUID
	refreshTokens map[uuid.UUID]*models.RefreshToken
}

var errNotFound = errors.New("not found")

func newControllerUserRepo() *controllerUserRepo {
	return &controllerUserRepo{
		users:         make(map[uuid.UUID]*models.User),
		byINS:         make(map[string]uuid.UUID),
		refreshTokens: make(map[uuid.UUID]*models.RefreshToken),
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

func (r *controllerUserRepo) GetByUsername(context.Context, string) (*models.User, error) {
	return nil, errNotFound
}

func (r *controllerUserRepo) ListByRole(_ context.Context, role models.UserRole) ([]models.User, error) {
	var data []models.User
	for _, u := range r.users {
		if u.Role == role {
			data = append(data, *u)
		}
	}
	return data, nil
}

func (r *controllerUserRepo) Update(context.Context, *models.User) error { return nil }

func (r *controllerUserRepo) AttachStudentProfile(context.Context, *models.StudentProfile) error { return nil }

func (r *controllerUserRepo) AttachTeacherProfile(context.Context, *models.TeacherProfile) error { return nil }

func (r *controllerUserRepo) AttachStaffProfile(context.Context, *models.StaffProfile) error { return nil }

func (r *controllerUserRepo) UpsertRefreshToken(_ context.Context, token *models.RefreshToken) error {
	r.refreshTokens[token.UserID] = token
	return nil
}

var _ repository.UserRepository = (*controllerUserRepo)(nil)

func TestAuthControllerLoginByINS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newControllerUserRepo()
	svc := service.NewAuthService(repo, "secret", time.Minute, time.Hour)
	ins := "INS-CTRL-1"
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:           uuid.New(),
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
