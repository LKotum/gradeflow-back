package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	reqdto "gradeflow/internal/domain/dto/request"
	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

// AdminService handles administrator operations such as provisioning dean staff.
type AdminService struct {
	users repository.UserRepository
}

// NewAdminService creates the service.
func NewAdminService(users repository.UserRepository) *AdminService {
	return &AdminService{users: users}
}

// CreateDean provisions a dean office staff user.
func (s *AdminService) CreateDean(ctx context.Context, payload reqdto.CreateDeanRequest) (*respdto.UserProfile, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
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
        Base: models.Base{ID: uuid.New()},
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

// ListDeans returns dean staff profiles.
func (s *AdminService) ListDeans(ctx context.Context) ([]respdto.UserProfile, error) {
	users, err := s.users.ListByRole(ctx, models.UserRoleDean)
	if err != nil {
		return nil, fmt.Errorf("list deans: %w", err)
	}
	resp := make([]respdto.UserProfile, 0, len(users))
	for _, u := range users {
		resp = append(resp, respdto.UserProfile{
			ID:         u.ID.String(),
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			MiddleName: u.MiddleName,
			Email:      u.Email,
			INS:        u.INS,
			AvatarURL:  u.AvatarURL,
			Role:       string(u.Role),
		})
	}
	return resp, nil
}
