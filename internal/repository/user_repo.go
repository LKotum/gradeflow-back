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
    ListByRole(ctx context.Context, role models.UserRole, opts ListOptions) ([]models.User, int64, error)
    ListDeletedByRole(ctx context.Context, role models.UserRole, opts ListOptions) ([]models.User, int64, error)
    Update(ctx context.Context, user *models.User) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    Restore(ctx context.Context, id uuid.UUID) error

    AttachStudentProfile(ctx context.Context, profile *models.StudentProfile) error
    AttachTeacherProfile(ctx context.Context, profile *models.TeacherProfile) error
    AttachStaffProfile(ctx context.Context, profile *models.StaffProfile) error
    UpsertRefreshToken(ctx context.Context, token *models.RefreshToken) error
    DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error
    NextINS(ctx context.Context) (string, error)
}
