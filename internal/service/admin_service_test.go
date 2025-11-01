package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/cache"
)

type adminUserRepoStub struct {
	users       map[uuid.UUID]*models.User
	softDeleted []uuid.UUID
}

func newAdminUserRepoStub() *adminUserRepoStub {
	return &adminUserRepoStub{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (r *adminUserRepoStub) Create(context.Context, *models.User) error { return nil }

func (r *adminUserRepoStub) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return user, nil
}

func (r *adminUserRepoStub) GetByINS(context.Context, string) (*models.User, error) { return nil, nil }

func (r *adminUserRepoStub) ListByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *adminUserRepoStub) ListDeletedByRole(context.Context, models.UserRole, repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *adminUserRepoStub) Update(context.Context, *models.User) error { return nil }

func (r *adminUserRepoStub) SoftDelete(_ context.Context, id uuid.UUID) error {
	r.softDeleted = append(r.softDeleted, id)
	return nil
}

func (r *adminUserRepoStub) Restore(context.Context, uuid.UUID) error { return nil }

func (r *adminUserRepoStub) AttachStudentProfile(context.Context, *models.StudentProfile) error { return nil }

func (r *adminUserRepoStub) AttachTeacherProfile(context.Context, *models.TeacherProfile) error { return nil }

func (r *adminUserRepoStub) AttachStaffProfile(context.Context, *models.StaffProfile) error { return nil }

func (r *adminUserRepoStub) UpsertRefreshToken(context.Context, *models.RefreshToken) error { return nil }
func (r *adminUserRepoStub) DeleteRefreshToken(context.Context, uuid.UUID) error       { return nil }

func (r *adminUserRepoStub) NextINS(context.Context) (string, error) { return "00000001", nil }

func TestAdminServiceDeleteUser_PreventsAdminRemoval(t *testing.T) {
	repo := newAdminUserRepoStub()
	adminID := uuid.New()
	repo.users[adminID] = &models.User{
		Base: models.Base{ID: adminID},
		Role: models.UserRoleAdmin,
	}
	service := NewAdminService(repo, nil, nil, nil, nil, cache.NewNoop(), nil)

	err := service.DeleteUser(context.Background(), adminID)
	if err == nil {
		t.Fatalf("expected error when deleting admin account")
	}
	if len(repo.softDeleted) != 0 {
		t.Fatalf("admin user must not be soft deleted")
	}
}
