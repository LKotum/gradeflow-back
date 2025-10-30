package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	response "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/middleware"
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
	"gradeflow/pkg/httpx"
)

// ensure swagger picks up response types without unused import warnings
var (
	_ response.AuthResponse
)

// AuthController exposes authentication endpoints.
type AuthController struct {
	auth  *service.AuthService
	users repository.UserRepository
}

// NewAuthController wires service into controller.
func NewAuthController(auth *service.AuthService, users repository.UserRepository) *AuthController {
	return &AuthController{auth: auth, users: users}
}

// RegisterPublicRoutes binds unauthenticated routes.
func (c *AuthController) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/login/ins", c.loginByINS)
	rg.POST("/refresh", c.refresh)
}

// RegisterPrivateRoutes binds authenticated routes.
func (c *AuthController) RegisterPrivateRoutes(rg *gin.RouterGroup) {
	rg.GET("/me", c.me)
	rg.PATCH("/password", c.changePassword)
}

// loginByINS godoc
// @Summary      Login by INS
// @Description  Authenticates student/teacher/dean using individual number.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      request.INSLoginRequest  true  "Credentials"
// @Success      200      {object}  response.AuthResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Router       /auth/login/ins [post]
func (c *AuthController) loginByINS(ctx *gin.Context) {
	var payload reqdto.INSLoginRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.auth.LoginByINS(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_credentials", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// refresh godoc
// @Summary      Refresh access token
// @Description  Exchanges refresh token for a new access token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      request.RefreshTokenRequest  true  "Refresh token"
// @Success      200      {object}  response.AuthResponse
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Router       /auth/refresh [post]
func (c *AuthController) refresh(ctx *gin.Context) {
	var payload reqdto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	resp, err := c.auth.Refresh(ctx.Request.Context(), payload)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_credentials", err.Error(), nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, resp)
}

// changePassword godoc
// @Summary      Change password
// @Security     BearerAuth
// @Tags         Auth
// @Accept       json
// @Param        payload  body      request.ChangePasswordRequest  true  "Password change payload"
// @Success      204
// @Failure      400      {object}  response.ErrorResponse
// @Failure      401      {object}  response.ErrorResponse
// @Router       /auth/password [patch]
func (c *AuthController) changePassword(ctx *gin.Context) {
	userID, ok := c.currentUserID(ctx)
	if !ok {
		return
	}
	var payload reqdto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		httpx.WriteError(ctx, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}
	if err := c.auth.ChangePassword(ctx.Request.Context(), userID, payload); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.WriteError(ctx, http.StatusBadRequest, "invalid_credentials", "current password is incorrect", nil)
			return
		}
		httpx.WriteError(ctx, http.StatusBadRequest, "password_change_failed", err.Error(), nil)
		return
	}
	httpx.WriteNoContent(ctx)
}

// me returns authenticated user profile summary.
// @Summary      Current user
// @Security     BearerAuth
// @Tags         Auth
// @Produce      json
// @Success      200      {object}  response.UserSummary
// @Failure      401      {object}  response.ErrorResponse
// @Router       /auth/me [get]
func (c *AuthController) me(ctx *gin.Context) {
	userID, ok := c.currentUserID(ctx)
	if !ok {
		return
	}
	user, err := c.users.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "user_not_found", "user not found", nil)
		return
	}
	httpx.WriteData(ctx, http.StatusOK, response.UserSummary{
		ID:         user.ID.String(),
		Role:       user.Role,
		INS:        user.INS,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		AvatarURL:  user.AvatarURL,
	})
}

func (c *AuthController) currentUserID(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		httpx.WriteError(ctx, http.StatusUnauthorized, "missing_user", "user context missing", nil)
		return uuid.UUID{}, false
	}
	userIDStr, ok := val.(string)
	if !ok {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_context", "invalid user context", nil)
		return uuid.UUID{}, false
	}
	userID, err := uuidFromString(userIDStr)
	if err != nil {
		httpx.WriteError(ctx, http.StatusUnauthorized, "invalid_context", "invalid user identifier", nil)
		return uuid.UUID{}, false
	}
	return userID, true
}

func uuidFromString(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
