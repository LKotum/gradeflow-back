package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	reqdto "gradeflow/internal/domain/dto/request"
	response "gradeflow/internal/domain/dto/response"
	"gradeflow/internal/middleware"
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
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
	rg.POST("/login/admin", c.loginAdmin)
	rg.POST("/refresh", c.refresh)
}

// RegisterPrivateRoutes binds authenticated routes.
func (c *AuthController) RegisterPrivateRoutes(rg *gin.RouterGroup) {
	rg.GET("/me", c.me)
}

// loginByINS godoc
// @Summary      Login by INS
// @Description  Authenticates student/teacher/dean using individual number.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      request.INSLoginRequest  true  "Credentials"
// @Success      200      {object}  response.AuthResponse
// @Failure      401      {object}  gin.H
// @Router       /auth/login/ins [post]
func (c *AuthController) loginByINS(ctx *gin.Context) {
	var payload reqdto.INSLoginRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.auth.LoginByINS(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// loginAdmin godoc
// @Summary      Admin login
// @Description  Authenticates admin using username/password.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      request.AdminLoginRequest  true  "Credentials"
// @Success      200      {object}  response.AuthResponse
// @Failure      401      {object}  gin.H
// @Router       /auth/login/admin [post]
func (c *AuthController) loginAdmin(ctx *gin.Context) {
	var payload reqdto.AdminLoginRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.auth.LoginAdmin(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// refresh godoc
// @Summary      Refresh access token
// @Description  Exchanges refresh token for a new access token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload  body      request.RefreshTokenRequest  true  "Refresh token"
// @Success      200      {object}  response.AuthResponse
// @Failure      401      {object}  gin.H
// @Router       /auth/refresh [post]
func (c *AuthController) refresh(ctx *gin.Context) {
	var payload reqdto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.auth.Refresh(ctx.Request.Context(), payload)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// me returns authenticated user profile summary.
// @Summary      Current user
// @Security     BearerAuth
// @Tags         Auth
// @Produce      json
// @Success      200      {object}  response.UserSummary
// @Failure      401      {object}  gin.H
// @Router       /auth/me [get]
func (c *AuthController) me(ctx *gin.Context) {
	val, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
		return
	}
	userIDStr, ok := val.(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}
	userID, err := uuidFromString(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}
	user, err := c.users.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	ctx.JSON(http.StatusOK, response.UserSummary{
		ID:         user.ID.String(),
		Role:       user.Role,
		INS:        user.INS,
		Username:   user.Username,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		AvatarURL:  user.AvatarURL,
	})
}

func uuidFromString(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
