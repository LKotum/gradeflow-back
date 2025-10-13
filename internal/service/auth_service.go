package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gradeflow/internal/config"
	m "gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	Users  *repository.UserRepository
	Tokens *repository.TokenRepository
	RTS    utils.TokenStore
	Cfg    config.Config
}

func NewAuthService(users *repository.UserRepository, tokens *repository.TokenRepository, rts utils.TokenStore, cfg config.Config) *AuthService {
	return &AuthService{Users: users, Tokens: tokens, RTS: rts, Cfg: cfg}
}

func (s *AuthService) HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}
func (s *AuthService) CheckPassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

func (s *AuthService) CreateUser(ctx context.Context, email, fullName, role, password string) (*m.User, error) {
	h, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &m.User{Email: email, FullName: fullName, Role: role, PasswordHash: h, Status: "pending"}
	if err := s.Users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) IssueTokens(ctx context.Context, u *m.User) (access string, refresh string, err error) {
	now := time.Now()
	acc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  u.ID,
		"role": u.Role,
		"exp":  now.Add(s.Cfg.AccessTTL).Unix(),
	})
	accStr, err := acc.SignedString([]byte(s.Cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	rRaw := hex.EncodeToString(buf)
	// Store refresh in Redis via TokenStore
	if s.RTS != nil {
		if err := s.RTS.Set(ctx, "refresh:"+rRaw, u.ID, s.Cfg.RefreshTTL); err != nil {
			return "", "", err
		}
	}
	return accStr, rRaw, nil
}

func (s *AuthService) VerifyPasswordLogin(ctx context.Context, email, password string) (*m.User, error) {
	u, err := s.Users.ByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := s.CheckPassword(u.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	if u.Status != "active" {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *AuthService) GenerateTOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}
func (s *AuthService) VerifyTOTP(secret string, code string, now time.Time) bool {
	slot := now.Unix() / 30
	h := sha256.Sum256([]byte(secret + ":" + string(rune(slot))))
	dec := int64(0)
	for _, ch := range h[len(h)-3:] {
		dec = dec*256 + int64(ch)
	}
	want := dec % 1000000
	got := code
	if len(got) != 6 {
		return false
	}
	w := []byte{'0', '0', '0', '0', '0', '0'}
	for i := 5; i >= 0; i-- {
		w[i] = byte('0' + (want % 10))
		want /= 10
	}
	return string(w) == got
}
