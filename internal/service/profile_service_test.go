package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
)

type memoryObject struct {
	data        []byte
	contentType string
}

type memoryStorage struct {
	objects map[string]map[string]memoryObject
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		objects: make(map[string]map[string]memoryObject),
	}
}

func (m *memoryStorage) bucket(name string) map[string]memoryObject {
	if _, ok := m.objects[name]; !ok {
		m.objects[name] = make(map[string]memoryObject)
	}
	return m.objects[name]
}

func (m *memoryStorage) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	m.bucket(bucketName)[objectName] = memoryObject{
		data:        buf,
		contentType: opts.ContentType,
	}
	return minio.UploadInfo{Size: int64(len(buf))}, nil
}

func (m *memoryStorage) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	obj, ok := m.bucket(bucketName)[objectName]
	if !ok {
		return nil, minio.ErrorResponse{StatusCode: http.StatusNotFound}
	}
	return io.NopCloser(bytes.NewReader(obj.data)), nil
}

func (m *memoryStorage) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	obj, ok := m.bucket(bucketName)[objectName]
	if !ok {
		return minio.ObjectInfo{}, minio.ErrorResponse{StatusCode: http.StatusNotFound}
	}
	return minio.ObjectInfo{
		Size:        int64(len(obj.data)),
		ContentType: obj.contentType,
	}, nil
}

func (m *memoryStorage) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	delete(m.bucket(bucketName), objectName)
	return nil
}

func (m *memoryStorage) has(bucket, object string) bool {
	_, ok := m.bucket(bucket)[object]
	return ok
}

type profileUserRepo struct {
	users map[uuid.UUID]*models.User
}

func newProfileUserRepo() *profileUserRepo {
	return &profileUserRepo{users: make(map[uuid.UUID]*models.User)}
}

func (r *profileUserRepo) Create(ctx context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *profileUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("not found")
}

func (r *profileUserRepo) GetByINS(ctx context.Context, ins string) (*models.User, error) {
	for _, user := range r.users {
		if user.INS != nil && *user.INS == ins {
			return user, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *profileUserRepo) ListByRole(ctx context.Context, role models.UserRole, opts repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *profileUserRepo) ListDeletedByRole(ctx context.Context, role models.UserRole, opts repository.ListOptions) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (r *profileUserRepo) Update(ctx context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *profileUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *profileUserRepo) Restore(ctx context.Context, id uuid.UUID) error    { return nil }

func (r *profileUserRepo) AttachStudentProfile(ctx context.Context, profile *models.StudentProfile) error {
	return nil
}

func (r *profileUserRepo) AttachTeacherProfile(ctx context.Context, profile *models.TeacherProfile) error {
	return nil
}

func (r *profileUserRepo) AttachStaffProfile(ctx context.Context, profile *models.StaffProfile) error {
	return nil
}

func (r *profileUserRepo) UpsertRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return nil
}

func (r *profileUserRepo) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *profileUserRepo) NextINS(ctx context.Context) (string, error) {
	return "00000001", nil
}

func TestProfileServiceUploadAvatar(t *testing.T) {
	repo := newProfileUserRepo()
	userID := uuid.New()
	repo.users[userID] = &models.User{
		Base:      models.Base{ID: userID},
		Role:      models.UserRoleTeacher,
		FirstName: "Ivan",
		LastName:  "Petrov",
	}

	storage := newMemoryStorage()
	service := NewProfileService(repo, nil, "avatars", "/api")
	service.storage = storage
	service.bucket = "avatars"
	service.timeSource = func() time.Time { return time.Unix(1700000000, 0) }

	img := image.NewRGBA(image.Rect(0, 0, 320, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	profile, err := service.UploadAvatar(context.Background(), userID, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("upload avatar: %v", err)
	}
	if profile.AvatarURL == nil || !strings.HasPrefix(*profile.AvatarURL, "/api/profile/avatar") {
		t.Fatalf("expected avatar url with api prefix, got %v", profile.AvatarURL)
	}
	objectName := avatarObjectPrefix + userID.String() + ".png"
	if !storage.has("avatars", objectName) {
		t.Fatalf("expected object %s to exist in storage", objectName)
	}

	stream, err := service.GetAvatar(context.Background(), userID)
	if err != nil {
		t.Fatalf("get avatar: %v", err)
	}
	defer stream.Reader.Close()
	payload, err := io.ReadAll(stream.Reader)
	if err != nil {
		t.Fatalf("read avatar: %v", err)
	}
	if len(payload) == 0 {
		t.Fatalf("avatar payload empty")
	}
}

func TestProfileServiceDeleteAndGet(t *testing.T) {
	repo := newProfileUserRepo()
	userID := uuid.New()
	repo.users[userID] = &models.User{
		Base:      models.Base{ID: userID},
		Role:      models.UserRoleStudent,
		FirstName: "Anna",
		LastName:  "Sidorova",
	}
	storage := newMemoryStorage()
	objectName := avatarObjectPrefix + userID.String() + ".png"
	storage.bucket("avatars")[objectName] = memoryObject{data: []byte("avatar"), contentType: "image/png"}

	service := NewProfileService(repo, nil, "avatars", "/api")
	service.storage = storage
	service.bucket = "avatars"

	if _, err := service.DeleteAvatar(context.Background(), userID); err != nil {
		t.Fatalf("delete avatar: %v", err)
	}
	if storage.has("avatars", objectName) {
		t.Fatalf("avatar should be removed from storage")
	}
	if _, err := service.GetAvatar(context.Background(), userID); !errors.Is(err, ErrAvatarNotFound) {
		t.Fatalf("expected ErrAvatarNotFound, got %v", err)
	}
}

func TestProfileServiceInvalidAvatar(t *testing.T) {
	repo := newProfileUserRepo()
	userID := uuid.New()
	repo.users[userID] = &models.User{Base: models.Base{ID: userID}, Role: models.UserRoleDean, FirstName: "Oleg", LastName: "Ivanov"}
	service := NewProfileService(repo, nil, "avatars", "/api")

	_, err := service.UploadAvatar(context.Background(), userID, bytes.NewReader([]byte("not an image")))
	if !errors.Is(err, ErrAvatarNotConfigured) {
		t.Fatalf("expected ErrAvatarNotConfigured due to missing storage, got %v", err)
	}

	storage := newMemoryStorage()
	service.storage = storage
	service.bucket = "avatars"

	_, err = service.UploadAvatar(context.Background(), userID, bytes.NewReader([]byte("not an image")))
	if !errors.Is(err, ErrInvalidAvatar) {
		t.Fatalf("expected ErrInvalidAvatar, got %v", err)
	}

	large := bytes.Repeat([]byte("a"), maxAvatarBytes+10)
	_, err = service.UploadAvatar(context.Background(), userID, bytes.NewReader(large))
	if !errors.Is(err, ErrAvatarTooLarge) {
		t.Fatalf("expected ErrAvatarTooLarge, got %v", err)
	}
}
