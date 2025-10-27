package repository

import (
	"context"

	"github.com/google/uuid"

	"gradeflow/internal/domain/models"
)

// UserRepository abstracts persistence for users and related aggregates.
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    GetByINS(ctx context.Context, ins string) (*models.User, error)
    ListByRole(ctx context.Context, role models.UserRole) ([]models.User, error)
    Update(ctx context.Context, user *models.User) error

    AttachStudentProfile(ctx context.Context, profile *models.StudentProfile) error
    AttachTeacherProfile(ctx context.Context, profile *models.TeacherProfile) error
    AttachStaffProfile(ctx context.Context, profile *models.StaffProfile) error
    UpsertRefreshToken(ctx context.Context, token *models.RefreshToken) error
    NextINS(ctx context.Context) (string, error)
}
