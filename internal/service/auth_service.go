package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

// ErrInvalidCredentials is returned when supplied credentials do not match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthService orchestrates authentication and JWT token issuance.
type AuthService struct {
	users      repository.UserRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

// NewAuthService constructs the service with dependencies.
func NewAuthService(users repository.UserRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		users:      users,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

// LoginByINS authenticates using the individual number (students, teachers, staff).
func (s *AuthService) LoginByINS(ctx context.Context, payload reqdto.INSLoginRequest) (*respdto.AuthResponse, error) {
	user, err := s.users.GetByINS(ctx, payload.INS)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user)
}

// Refresh exchanges a refresh token for a new pair.
func (s *AuthService) Refresh(ctx context.Context, payload reqdto.RefreshTokenRequest) (*respdto.AuthResponse, error) {
	token, err := jwt.ParseWithClaims(payload.RefreshToken, &jwtRegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	claims, ok := token.Claims.(*jwtRegisteredClaims)
	if !ok || !token.Valid || claims.Type != "refresh" {
		return nil, ErrInvalidCredentials
	}
	userID, parseErr := uuid.Parse(claims.Subject)
	if parseErr != nil {
		return nil, ErrInvalidCredentials
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.RefreshToken == nil || user.RefreshToken.Token != payload.RefreshToken || user.RefreshToken.ExpiresAt.Before(s.now()) {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User) (*respdto.AuthResponse, error) {
	now := s.now()
	accessExpires := now.Add(s.accessTTL)
	refreshExpires := now.Add(s.refreshTTL)

	accessClaims := jwtRegisteredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(accessExpires),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Type: "access",
		Role: string(user.Role),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwtRegisteredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(refreshExpires),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Type: "refresh",
		Role: string(user.Role),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if err := s.users.UpsertRefreshToken(ctx, &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshExpires,
	}); err != nil {
		return nil, fmt.Errorf("persist refresh token: %w", err)
	}

	return &respdto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpires,
		User: respdto.UserSummary{
			ID:         user.ID.String(),
			Role:       user.Role,
			INS:        user.INS,
			Email:      user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			MiddleName: user.MiddleName,
			AvatarURL:  user.AvatarURL,
		},
	}, nil
}

// ChangePassword updates user credentials after validating current password.
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, payload reqdto.ChangePasswordRequest) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("load user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.CurrentPassword)) != nil {
		return ErrInvalidCredentials
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hashed)
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := s.users.DeleteRefreshToken(ctx, user.ID); err != nil {
		return fmt.Errorf("clear refresh token: %w", err)
	}
	return nil
}

type jwtRegisteredClaims struct {
	jwt.RegisteredClaims
	Type string `json:"type"`
	Role string `json:"role"`
}
