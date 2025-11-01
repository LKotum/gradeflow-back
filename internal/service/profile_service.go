package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	respdto "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/pkg/cache"
	"gradeflow/pkg/logger"
	"gradeflow/pkg/utils"
)

const (
	avatarObjectPrefix = "avatars/"
	maxAvatarBytes     = 5 * 1024 * 1024
	targetAvatarSize   = 128
)

var (
	// ErrAvatarNotConfigured is returned when avatar storage is not available.
	ErrAvatarNotConfigured = errors.New("avatar storage not configured")

	// ErrAvatarNotFound is returned when user avatar is missing.
	ErrAvatarNotFound = errors.New("avatar not found")

	// ErrInvalidAvatar is returned when provided avatar has unsupported format.
	ErrInvalidAvatar = errors.New("invalid avatar file")

	// ErrAvatarTooLarge is returned when provided avatar exceeds size limit.
	ErrAvatarTooLarge = errors.New("avatar file too large")
)

type avatarStorage interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}

type minioStorage struct {
	client *minio.Client
}

func (m *minioStorage) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (m *minioStorage) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (m *minioStorage) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return m.client.StatObject(ctx, bucketName, objectName, opts)
}

func (m *minioStorage) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return m.client.RemoveObject(ctx, bucketName, objectName, opts)
}

// AvatarStream wraps avatar payload with metadata.
type AvatarStream struct {
	Reader      io.ReadCloser
	Size        int64
	ContentType string
}

// ProfileService exposes profile-related operations (avatars, preferences).
type ProfileService struct {
	users         repository.UserRepository
	storage       avatarStorage
	bucket        string
	avatarAPIPath string
	timeSource    func() time.Time
	cache         cache.Store
}

// NewProfileService constructs profile service.
func NewProfileService(users repository.UserRepository, minioClient *minio.Client, cacheStore cache.Store, bucket, apiBasePath string) *ProfileService {
	var storage avatarStorage
	if minioClient != nil && bucket != "" {
		storage = &minioStorage{client: minioClient}
	}
	trimmed := strings.TrimSpace(apiBasePath)
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" || trimmed == "/" {
		trimmed = "/profile/avatar"
	} else {
		if !strings.HasPrefix(trimmed, "/") {
			trimmed = "/" + trimmed
		}
		trimmed = trimmed + "/profile/avatar"
	}
	if cacheStore == nil {
		cacheStore = cache.NewNoop()
	}
	return &ProfileService{
		users:         users,
		storage:       storage,
		bucket:        bucket,
		avatarAPIPath: trimmed,
		timeSource:    time.Now,
		cache:         cacheStore,
	}
}

// Profile returns user profile summary for authenticated user.
func (s *ProfileService) Profile(ctx context.Context, userID uuid.UUID) (*respdto.UserProfile, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	profile := buildUserProfile(user)
	return &profile, nil
}

// UploadAvatar stores user avatar (resized to 128x128) in MinIO and updates profile.
func (s *ProfileService) UploadAvatar(ctx context.Context, userID uuid.UUID, data io.Reader) (*respdto.UserProfile, error) {
    if s.storage == nil || s.bucket == "" {
        return nil, ErrAvatarNotConfigured
    }
    return s.uploadAvatarFor(ctx, userID, data)
}

// DeleteAvatar removes avatar object and clears profile reference.
func (s *ProfileService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*respdto.UserProfile, error) {
    if s.storage == nil || s.bucket == "" {
        return nil, ErrAvatarNotConfigured
    }
    return s.deleteAvatarFor(ctx, userID)
}

// UploadAvatarFor allows privileged users to update someone else's avatar.
func (s *ProfileService) UploadAvatarFor(ctx context.Context, userID uuid.UUID, data io.Reader) (*respdto.UserProfile, error) {
    if s.storage == nil || s.bucket == "" {
        return nil, ErrAvatarNotConfigured
    }
    return s.uploadAvatarFor(ctx, userID, data)
}

// DeleteAvatarFor allows privileged users to remove someone else's avatar.
func (s *ProfileService) DeleteAvatarFor(ctx context.Context, userID uuid.UUID) (*respdto.UserProfile, error) {
    if s.storage == nil || s.bucket == "" {
        return nil, ErrAvatarNotConfigured
    }
    return s.deleteAvatarFor(ctx, userID)
}

// GetAvatar streams avatar object from storage.
func (s *ProfileService) GetAvatar(ctx context.Context, userID uuid.UUID) (*AvatarStream, error) {
	if s.storage == nil || s.bucket == "" {
		return nil, ErrAvatarNotConfigured
	}
	objectName := avatarObjectPrefix + userID.String() + ".png"
	info, err := s.storage.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrAvatarNotFound
		}
		return nil, fmt.Errorf("stat avatar: %w", err)
	}
	reader, err := s.storage.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrAvatarNotFound
		}
		return nil, fmt.Errorf("load avatar: %w", err)
	}
	contentType := info.ContentType
	if strings.TrimSpace(contentType) == "" {
		contentType = "image/png"
	}
	return &AvatarStream{
		Reader:      reader,
		Size:        info.Size,
		ContentType: contentType,
	}, nil
}

func (s *ProfileService) uploadAvatarFor(ctx context.Context, userID uuid.UUID, data io.Reader) (*respdto.UserProfile, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	payload, err := prepareAvatarPayload(data)
	if err != nil {
		return nil, err
	}
	return s.storeAvatar(ctx, user, payload)
}

func (s *ProfileService) deleteAvatarFor(ctx context.Context, userID uuid.UUID) (*respdto.UserProfile, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	if err := s.removeAvatarObject(ctx, user.ID); err != nil {
		logger.Warn("remove avatar failed", "error", err, "user", user.ID)
	}
	user.AvatarURL = nil
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user avatar: %w", err)
	}
	s.invalidateUserCaches(ctx, user)
	profile := buildUserProfile(user)
	return &profile, nil
}

func (s *ProfileService) storeAvatar(ctx context.Context, user *models.User, payload []byte) (*respdto.UserProfile, error) {
	objectName := avatarObjectPrefix + user.ID.String() + ".png"
	reader := bytes.NewReader(payload)
	opts := minio.PutObjectOptions{ContentType: "image/png"}
	if _, err := s.storage.PutObject(ctx, s.bucket, objectName, reader, int64(len(payload)), opts); err != nil {
		return nil, fmt.Errorf("store avatar: %w", err)
	}
	versionedURL := fmt.Sprintf("%s?v=%d", s.avatarAPIPath, s.timeSource().Unix())
	user.AvatarURL = utils.StringPtr(versionedURL)
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user avatar: %w", err)
	}
	s.invalidateUserCaches(ctx, user)
	profile := buildUserProfile(user)
	return &profile, nil
}

func (s *ProfileService) removeAvatarObject(ctx context.Context, userID uuid.UUID) error {
	objectName := avatarObjectPrefix + userID.String() + ".png"
	return s.storage.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func prepareAvatarPayload(data io.Reader) ([]byte, error) {
	payload, err := readAvatarPayload(data)
	if err != nil {
		return nil, err
	}
	img, format, decodeErr := image.Decode(bytes.NewReader(payload))
	if decodeErr != nil {
		return nil, ErrInvalidAvatar
	}
	if !isSupportedAvatarFormat(format) {
		return nil, ErrInvalidAvatar
	}
	resized := resizeToSquare(img, targetAvatarSize)
	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("encode avatar: %w", err)
	}
	return buf.Bytes(), nil
}

func readAvatarPayload(r io.Reader) ([]byte, error) {
	limited := io.LimitReader(r, maxAvatarBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read avatar: %w", err)
	}
	if len(data) == 0 {
		return nil, ErrInvalidAvatar
	}
	if len(data) > maxAvatarBytes {
		return nil, ErrAvatarTooLarge
	}
	return data, nil
}

func isSupportedAvatarFormat(format string) bool {
	switch strings.ToLower(format) {
	case "jpeg", "jpg", "png", "gif":
		return true
	default:
		return false
	}
}

func resizeToSquare(img image.Image, size int) *image.NRGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return image.NewNRGBA(image.Rect(0, 0, size, size))
	}
	side := width
	if height < width {
		side = height
	}
	srcStartX := bounds.Min.X + (width-side)/2
	srcStartY := bounds.Min.Y + (height-side)/2
	cropped := image.NewNRGBA(image.Rect(0, 0, side, side))
	draw.Draw(cropped, cropped.Bounds(), img, image.Point{X: srcStartX, Y: srcStartY}, draw.Src)

	scaled := image.NewNRGBA(image.Rect(0, 0, size, size))
	if side == size {
		draw.Draw(scaled, scaled.Bounds(), cropped, cropped.Bounds().Min, draw.Src)
		return scaled
	}
	for y := 0; y < size; y++ {
		srcY := int(math.Min(float64(side-1), math.Floor(float64(y*side)/float64(size))))
		for x := 0; x < size; x++ {
			srcX := int(math.Min(float64(side-1), math.Floor(float64(x*side)/float64(size))))
			scaled.Set(x, y, cropped.NRGBAAt(srcX, srcY))
		}
	}
	return scaled
}

func (s *ProfileService) invalidateUserCaches(ctx context.Context, user *models.User) {
	if s.cache == nil || user == nil {
		return
	}
	prefixes := []string{"admin:users"}
	switch user.Role {
	case models.UserRoleDean:
		prefixes = append(prefixes, "admin:deans")
	case models.UserRoleTeacher:
		prefixes = append(prefixes, "dean:teachers", "dean:subjects", "dean:groups")
	case models.UserRoleStudent:
		prefixes = append(prefixes, "dean:students", "dean:groups", "dean:subjects")
	}
	for _, prefix := range prefixes {
		if err := s.cache.InvalidatePrefix(ctx, prefix); err != nil {
			logger.Warn("profile cache invalidate failed", "prefix", prefix, "error", err)
		}
	}
}

func buildUserProfile(user *models.User) respdto.UserProfile {
	if user == nil {
		return respdto.UserProfile{}
	}
	return respdto.UserProfile{
		ID:         user.ID.String(),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Email:      user.Email,
		INS:        user.INS,
		AvatarURL:  user.AvatarURL,
		Role:       string(user.Role),
	}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrAvatarNotFound) {
		return true
	}
	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		if minioErr.StatusCode == http.StatusNotFound || minioErr.Code == "NoSuchKey" || minioErr.Code == "NotFound" {
			return true
		}
	}
	resp := minio.ToErrorResponse(err)
	if resp.Code == "NoSuchKey" || resp.Code == "NotFound" || resp.StatusCode == http.StatusNotFound {
		return true
	}
	if resp.StatusCode == 0 && resp.Code == "" {
		if direct, ok := err.(minio.ErrorResponse); ok {
			if direct.StatusCode == http.StatusNotFound {
				return true
			}
		}
		if directPtr, ok := err.(*minio.ErrorResponse); ok && directPtr != nil {
			if directPtr.StatusCode == http.StatusNotFound {
				return true
			}
		}
		var apiErr interface{ StatusCode() int }
		if errors.As(err, &apiErr) {
			return apiErr.StatusCode() == http.StatusNotFound
		}
	}
	return false
}
