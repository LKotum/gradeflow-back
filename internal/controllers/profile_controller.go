package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	response "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/middleware"
	"gradeflow/internal/service"
	"gradeflow/pkg/httpx"
)

// ProfileController exposes profile endpoints for authenticated users.
type ProfileController struct {
	profiles *service.ProfileService
}

// NewProfileController constructs controller.
func NewProfileController(profiles *service.ProfileService) *ProfileController {
	return &ProfileController{profiles: profiles}
}

// RegisterRoutes wires profile endpoints.
func (c *ProfileController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", c.profile)
	rg.PUT("/avatar", c.uploadAvatar)
	rg.GET("/avatar", c.getAvatar)
	rg.DELETE("/avatar", c.deleteAvatar)
}

func (c *ProfileController) currentUserID(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		httpx.WriteError(ctx, http.StatusUnauthorized, "unauthorized", "user context missing", nil)
		return uuid.UUID{}, false
	}
	raw, ok := val.(string)
	if !ok {
		httpx.WriteError(ctx, http.StatusUnauthorized, "unauthorized", "invalid user context", nil)
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "unauthorized", "invalid user identifier", nil)
		return uuid.UUID{}, false
	}
	return id, true
}

// profile godoc
// @Summary      Профиль текущего пользователя
// @Security     BearerAuth
// @Tags         Profile
// @Produce      json
// @Success      200 {object} response.UserProfile
// @Failure      401 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /profile [get]
func (c *ProfileController) profile(ctx *gin.Context) {
	userID, ok := c.currentUserID(ctx)
	if !ok {
		return
	}
	profile, err := c.profiles.Profile(ctx.Request.Context(), userID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "profile_fetch_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, profile)
}

// uploadAvatar godoc
// @Summary      Загрузить или обновить аватар
// @Security     BearerAuth
// @Tags         Profile
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar  formData  file  true  "Файл изображения (png, jpg, gif)"
// @Success      200 {object} response.UserProfile
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Failure      413 {object} response.ErrorResponse
// @Failure      503 {object} response.ErrorResponse
// @Router       /profile/avatar [put]
func (c *ProfileController) uploadAvatar(ctx *gin.Context) {
	userID, ok := c.currentUserID(ctx)
	if !ok {
		return
	}
	file, _, err := ctx.Request.FormFile("avatar")
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_avatar", "avatar file is required", nil)
		return
	}
	defer file.Close()

	profile, err := c.profiles.UploadAvatar(ctx.Request.Context(), userID, file)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAvatarNotConfigured):
			httpx.WriteError(ctx, http.StatusServiceUnavailable, "avatar_unavailable", "avatar storage is not configured", nil)
		case errors.Is(err, service.ErrAvatarTooLarge):
			httpx.WriteError(ctx, http.StatusRequestEntityTooLarge, "avatar_too_large", "avatar file is too large", nil)
		case errors.Is(err, service.ErrInvalidAvatar):
			httpx.WriteError(ctx, http.StatusBadRequest, "invalid_avatar", "unsupported image format", nil)
		default:
			httpx.WriteError(ctx, http.StatusInternalServerError, "avatar_upload_failed", err.Error(), nil)
		}
		return
	}
	httpx.WriteData(ctx, http.StatusOK, profile)
}

// getAvatar godoc
// @Summary      Получить аватар текущего пользователя
// @Security     BearerAuth
// @Tags         Profile
// @Produce      octet-stream
// @Success      200
// @Failure      401 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      503 {object} response.ErrorResponse
// @Router       /profile/avatar [get]
func (c *ProfileController) getAvatar(ctx *gin.Context) {
	userID, ok := c.currentUserID(ctx)
	if !ok {
		return
	}
	stream, err := c.profiles.GetAvatar(ctx.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAvatarNotConfigured):
			httpx.WriteError(ctx, http.StatusServiceUnavailable, "avatar_unavailable", "avatar storage is not configured", nil)
		case errors.Is(err, service.ErrAvatarNotFound):
			httpx.WriteError(ctx, http.StatusNotFound, "avatar_not_found", "avatar not found", nil)
		default:
			httpx.WriteError(ctx, http.StatusInternalServerError, "avatar_fetch_failed", err.Error(), nil)
		}
		return
	}
	defer stream.Reader.Close()

	headers := map[string]string{
		"Cache-Control": "no-store, no-cache, must-revalidate, max-age=0",
	}
	ctx.DataFromReader(http.StatusOK, stream.Size, stream.ContentType, stream.Reader, headers)
}

// deleteAvatar godoc
// @Summary      Удалить аватар
// @Security     BearerAuth
// @Tags         Profile
// @Produce      json
// @Success      200 {object} response.UserProfile
// @Failure      401 {object} response.ErrorResponse
// @Failure      503 {object} response.ErrorResponse
// @Router       /profile/avatar [delete]
func (c *ProfileController) deleteAvatar(ctx *gin.Context) {
	userID, ok := c.currentUserID(ctx)
	if !ok {
		return
	}
	profile, err := c.profiles.DeleteAvatar(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrAvatarNotConfigured) {
			httpx.WriteError(ctx, http.StatusServiceUnavailable, "avatar_unavailable", "avatar storage is not configured", nil)
			return
		}
		httpx.WriteError(ctx, http.StatusInternalServerError, "avatar_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, profile)
}

var (
	_ response.UserProfile
)
