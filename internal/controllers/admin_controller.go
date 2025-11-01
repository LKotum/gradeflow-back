package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	"gradeflow/internal/domain/models"
	"gradeflow/internal/service"
	"gradeflow/pkg/httpx"
)

// AdminController manages administrator endpoints.
type AdminController struct {
	admins *service.AdminService
}

// NewAdminController creates the controller.
func NewAdminController(admins *service.AdminService) *AdminController {
	return &AdminController{admins: admins}
}

// RegisterRoutes registers admin endpoints.
// @Summary Admin operations
// @Tags Admin
// @Security BearerAuth
// @BasePath /admin
func (c *AdminController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/deans", c.createDean)
	rg.GET("/deans", c.listDeans)
	rg.PATCH("/deans/:deanId", c.updateDean)
	rg.DELETE("/deans/:deanId", c.deleteDean)
	rg.POST("/deans/:deanId/restore", c.restoreDean)

	rg.GET("/users", c.listUsers)
	rg.GET("/users/deleted", c.listDeletedUsers)
	rg.PATCH("/users/:userId", c.updateUser)
	rg.PATCH("/users/:userId/password", c.resetPassword)
	rg.DELETE("/users/:userId", c.deleteUser)
	rg.POST("/users/:userId/restore", c.restoreUser)
	rg.PUT("/users/:userId/avatar", c.uploadUserAvatar)
	rg.DELETE("/users/:userId/avatar", c.deleteUserAvatar)

	rg.GET("/groups/deleted", c.listDeletedGroups)
	rg.POST("/groups/:groupId/restore", c.restoreGroup)

	rg.GET("/subjects/deleted", c.listDeletedSubjects)
	rg.POST("/subjects/:subjectId/restore", c.restoreSubject)
}

// createDean godoc
// @Summary Create dean staff
// @Security BearerAuth
// @Tags Admin
// @Accept json
// @Produce json
// @Param payload body request.CreateDeanRequest true "Dean payload"
// @Success 201 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/deans [post]
func (c *AdminController) createDean(ctx *gin.Context) {
	var payload reqdto.CreateDeanRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.admins.CreateDean(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "dean_create_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusCreated, resp)
}

// listDeans godoc
// @Summary List dean staff
// @Security BearerAuth
// @Tags Admin
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedUserProfiles
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/deans [get]
func (c *AdminController) listDeans(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.admins.ListDeans(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "dean_list_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// updateDean godoc
// @Summary Update dean profile
// @Security BearerAuth
// @Tags Admin
// @Accept json
// @Produce json
// @Param deanId path string true "Dean ID"
// @Param payload body request.UpdateDeanRequest true "Update payload"
// @Success 200 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /admin/deans/{deanId} [patch]
func (c *AdminController) updateDean(ctx *gin.Context) {
	deanID, err := uuid.Parse(ctx.Param("deanId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_dean", "invalid dean identifier", nil)
		return
	}
	var payload reqdto.UpdateDeanRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.admins.UpdateDean(ctx.Request.Context(), deanID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "dean_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// deleteDean godoc
// @Summary Soft delete dean
// @Security BearerAuth
// @Tags Admin
// @Param deanId path string true "Dean ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/deans/{deanId} [delete]
func (c *AdminController) deleteDean(ctx *gin.Context) {
	deanID, err := uuid.Parse(ctx.Param("deanId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_dean", "invalid dean identifier", nil)
		return
	}
	if err := c.admins.DeleteDean(ctx.Request.Context(), deanID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "dean_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// restoreDean godoc
// @Summary Restore dean
// @Security BearerAuth
// @Tags Admin
// @Param deanId path string true "Dean ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/deans/{deanId}/restore [post]
func (c *AdminController) restoreDean(ctx *gin.Context) {
	deanID, err := uuid.Parse(ctx.Param("deanId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_dean", "invalid dean identifier", nil)
		return
	}
	if err := c.admins.RestoreDean(ctx.Request.Context(), deanID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "dean_restore_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// listUsers godoc
// @Summary List users by role
// @Security BearerAuth
// @Tags Admin
// @Produce json
// @Param role query string true "Role (student|teacher|dean)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedUserProfiles
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users [get]
func (c *AdminController) listUsers(ctx *gin.Context) {
	roleStr := ctx.Query("role")
	var role models.UserRole
	switch roleStr {
	case string(models.UserRoleStudent):
		role = models.UserRoleStudent
	case string(models.UserRoleTeacher):
		role = models.UserRoleTeacher
	case string(models.UserRoleDean):
		role = models.UserRoleDean
	default:
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_role", "role must be student, teacher or dean", nil)
		return
	}
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.admins.ListUsers(ctx.Request.Context(), role, query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "users_list_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// updateUser godoc
// @Summary Update user profile
// @Security BearerAuth
// @Tags Admin
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param payload body request.UpdateUserRequest true "Update payload"
// @Success 200 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/users/{userId} [patch]
func (c *AdminController) updateUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_user", "invalid user id", nil)
		return
	}
	var payload reqdto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.admins.UpdateUser(ctx.Request.Context(), userID, payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "user_update_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// listDeletedUsers godoc
// @Summary List deleted users by role
// @Security BearerAuth
// @Tags Admin
// @Produce json
// @Param role query string true "Role (student|teacher|dean)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedUserProfiles
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/deleted [get]
func (c *AdminController) listDeletedUsers(ctx *gin.Context) {
	roleStr := ctx.Query("role")
	var role models.UserRole
	switch roleStr {
	case string(models.UserRoleStudent):
		role = models.UserRoleStudent
	case string(models.UserRoleTeacher):
		role = models.UserRoleTeacher
	case string(models.UserRoleDean):
		role = models.UserRoleDean
	default:
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_role", "role must be student, teacher or dean", nil)
		return
	}
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.admins.ListDeletedUsers(ctx.Request.Context(), role, query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "deleted_users_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// deleteUser godoc
// @Summary Soft delete user by ID
// @Security BearerAuth
// @Tags Admin
// @Param userId path string true "User ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/users/{userId} [delete]
func (c *AdminController) deleteUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_user", "invalid user identifier", nil)
		return
	}
	if err := c.admins.DeleteUser(ctx.Request.Context(), userID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "user_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// uploadUserAvatar godoc
// @Summary Загрузить аватар пользователя
// @Security BearerAuth
// @Tags Admin
// @Accept mpfd
// @Produce json
// @Param userId path string true "ID пользователя"
// @Param avatar formData file true "Файл изображения"
// @Success 200 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/users/{userId}/avatar [put]
func (c *AdminController) uploadUserAvatar(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_user", "invalid user identifier", nil)
		return
	}
	file, _, err := ctx.Request.FormFile("avatar")
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_avatar", "avatar file is required", nil)
		return
	}
	defer file.Close()
	profile, err := c.admins.UpdateUserAvatar(ctx.Request.Context(), userID, file)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAvatarNotConfigured):
			httpx.WriteError(ctx, http.StatusServiceUnavailable, "avatar_unavailable", err.Error(), nil)
		case errors.Is(err, service.ErrInvalidAvatar):
			httpx.WriteError(ctx, http.StatusBadRequest, "invalid_avatar", err.Error(), nil)
		case errors.Is(err, service.ErrAvatarTooLarge):
			httpx.WriteError(ctx, http.StatusRequestEntityTooLarge, "avatar_too_large", err.Error(), nil)
		default:
			httpx.WriteError(ctx, http.StatusInternalServerError, "avatar_upload_failed", err.Error(), nil)
		}
		return
	}
	httpx.WriteData(ctx, http.StatusOK, profile)
}

// deleteUserAvatar godoc
// @Summary Удалить аватар пользователя
// @Security BearerAuth
// @Tags Admin
// @Param userId path string true "ID пользователя"
// @Success 200 {object} response.UserProfile
// @Failure 400 {object} response.ErrorResponse
// @Failure 503 {object} response.ErrorResponse
// @Router /admin/users/{userId}/avatar [delete]
func (c *AdminController) deleteUserAvatar(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_user", "invalid user identifier", nil)
		return
	}
	profile, err := c.admins.DeleteUserAvatar(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrAvatarNotConfigured) {
			httpx.WriteError(ctx, http.StatusServiceUnavailable, "avatar_unavailable", err.Error(), nil)
			return
		}
		if errors.Is(err, service.ErrAvatarNotFound) {
			httpx.WriteError(ctx, http.StatusNotFound, "avatar_not_found", err.Error(), nil)
			return
		}
		httpx.WriteError(ctx, http.StatusInternalServerError, "avatar_delete_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, profile)
}

// resetPassword godoc
// @Summary Reset user password
// @Security BearerAuth
// @Tags Admin
// @Accept json
// @Param userId path string true "User ID"
// @Param payload body request.ResetPasswordRequest true "New password"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/users/{userId}/password [patch]
func (c *AdminController) resetPassword(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_user", "invalid user identifier", nil)
		return
	}
	var payload reqdto.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	if err := c.admins.ResetPassword(ctx.Request.Context(), userID, payload.Password); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "password_reset_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// restoreUser godoc
// @Summary Restore user by ID
// @Security BearerAuth
// @Tags Admin
// @Param userId path string true "User ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/users/{userId}/restore [post]
func (c *AdminController) restoreUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_user", "invalid user identifier", nil)
		return
	}
	if err := c.admins.RestoreUser(ctx.Request.Context(), userID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "user_restore_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// listDeletedGroups godoc
// @Summary List deleted groups
// @Security BearerAuth
// @Tags Admin
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedGroupSummaries
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/groups/deleted [get]
func (c *AdminController) listDeletedGroups(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.admins.ListDeletedGroups(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "deleted_groups_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// restoreGroup godoc
// @Summary Restore group
// @Security BearerAuth
// @Tags Admin
// @Param groupId path string true "Group ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/groups/{groupId}/restore [post]
func (c *AdminController) restoreGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_group", "invalid group identifier", nil)
		return
	}
	if err := c.admins.RestoreGroup(ctx.Request.Context(), groupID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "group_restore_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// listDeletedSubjects godoc
// @Summary List deleted subjects
// @Security BearerAuth
// @Tags Admin
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search phrase"
// @Success 200 {object} response.PaginatedSubjectSummaries
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/subjects/deleted [get]
func (c *AdminController) listDeletedSubjects(ctx *gin.Context) {
	var query reqdto.PaginationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_query", err.Error(), nil)
		return
	}
	resp, err := c.admins.ListDeletedSubjects(ctx.Request.Context(), query)
	if err != nil {
		httpx.WriteError(ctx, http.StatusInternalServerError, "deleted_subjects_failed", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// restoreSubject godoc
// @Summary Restore subject
// @Security BearerAuth
// @Tags Admin
// @Param subjectId path string true "Subject ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Router /admin/subjects/{subjectId}/restore [post]
func (c *AdminController) restoreSubject(ctx *gin.Context) {
	subjectID, err := uuid.Parse(ctx.Param("subjectId"))
	if err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_subject", "invalid subject identifier", nil)
		return
	}
	if err := c.admins.RestoreSubject(ctx.Request.Context(), subjectID); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "subject_restore_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}
