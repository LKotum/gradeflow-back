package gormrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

// Ensure implementation conforms to repository interface.
var _ repository.UserRepository = (*UserRepositoryGorm)(nil)

// UserRepositoryGorm persists users using a GORM connection.
type UserRepositoryGorm struct {
	db *gorm.DB
}

// NewUserRepository constructs a UserRepository backed by GORM.
func NewUserRepository(db *gorm.DB) *UserRepositoryGorm {
	return &UserRepositoryGorm{db: db}
}

// Create inserts a new user row with related aggregates.
func (r *UserRepositoryGorm) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func preloadUser(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Student").
		Preload("Student.Group").
		Preload("Teacher").
		Preload("Staff").
		Preload("RefreshToken")
}

// GetByID retrieves a user by identifier.
func (r *UserRepositoryGorm) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := preloadUser(r.db.WithContext(ctx)).
		First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByINS fetches user by individual number.
func (r *UserRepositoryGorm) GetByINS(ctx context.Context, ins string) (*models.User, error) {
	var user models.User
	if err := preloadUser(r.db.WithContext(ctx)).
		Where("ins = ?", ins).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves user by username (for admin logins).
func (r *UserRepositoryGorm) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := preloadUser(r.db.WithContext(ctx)).
		Where("username = ?", username).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListByRole returns users by role.
func (r *UserRepositoryGorm) ListByRole(ctx context.Context, role models.UserRole) ([]models.User, error) {
	var users []models.User
	if err := preloadUser(r.db.WithContext(ctx)).
		Where("role = ?", role).
		Order("last_name asc, first_name asc").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Update persists updated user fields.
func (r *UserRepositoryGorm) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// AttachStudentProfile creates or updates the student profile.
func (r *UserRepositoryGorm) AttachStudentProfile(ctx context.Context, profile *models.StudentProfile) error {
	if profile == nil {
		return errors.New("nil student profile")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_id", "index", "updated_at"}),
	}).Create(profile).Error
}

// AttachTeacherProfile creates or updates the teacher profile.
func (r *UserRepositoryGorm) AttachTeacherProfile(ctx context.Context, profile *models.TeacherProfile) error {
	if profile == nil {
		return errors.New("nil teacher profile")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "bio", "updated_at"}),
	}).Create(profile).Error
}

// AttachStaffProfile creates or updates the staff profile.
func (r *UserRepositoryGorm) AttachStaffProfile(ctx context.Context, profile *models.StaffProfile) error {
	if profile == nil {
		return errors.New("nil staff profile")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"position", "updated_at"}),
	}).Create(profile).Error
}

// UpsertRefreshToken stores or updates a refresh token for a user.
func (r *UserRepositoryGorm) UpsertRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	if token == nil {
		return errors.New("nil refresh token provided")
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"token", "expires_at", "updated_at"}),
		}).
		Create(token).
		Error
}
