package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	req "gradeflow/internal/domain/dto/request"
	resp "gradeflow/internal/domain/dto/response"
	m "gradeflow/internal/domain/models"
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
	"gradeflow/pkg/middleware"
	"gradeflow/pkg/utils"
)

type AuthController struct {
	S      *service.AuthService
	DB     *gorm.DB
	Cfg    config.Config
	M      *utils.Mailer
	Tokens utils.TokenStore
}

func NewAuthController(db *gorm.DB, cfg config.Config, tokens utils.TokenStore) *AuthController {
	users := repository.NewUserRepository(db)
	rtRepo := repository.NewTokenRepository(db)
	svc := service.NewAuthService(users, rtRepo, tokens, cfg)
	var mler *utils.Mailer
	if cfg.SMTP.Host != "" {
		mler = utils.NewMailer(utils.SMTPConfig{
			Host:      cfg.SMTP.Host,
			Port:      cfg.SMTP.Port,
			User:      cfg.SMTP.User,
			Pass:      cfg.SMTP.Pass,
			FromEmail: cfg.SMTP.FromEmail,
			FromName:  cfg.SMTP.FromName,
		})
	}
	return &AuthController{S: svc, DB: db, Cfg: cfg, M: mler, Tokens: tokens}
}

func (h *AuthController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.register)
	rg.POST("/login", h.login)
	rg.POST("/refresh", h.refresh)
	rg.GET("/verify", h.verifyEmail)
	rg.POST("/request-reset", h.requestReset)
	rg.POST("/reset", h.resetPassword)

	pr := rg.Group("")
	pr.Use(middleware.JWT(h.Cfg, h.DB))
	pr.GET("/me", h.me)
	pr.POST("/logout", h.logout)
}

// @Summary      Register new user
// @Description  Creates a new user and sends a verification email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  request.Register  true  "registration"
// @Success      201    {object}  response.CreatedID
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /auth/register [post]
func (h *AuthController) register(c *gin.Context) {
	var r req.Register
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	u, err := h.S.CreateUser(c.Request.Context(), r.Email, r.FullName, r.Role, r.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	tok, err := secureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if err := h.Tokens.Set(c.Request.Context(), "verify:"+tok, u.ID, 24*time.Hour); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if h.M != nil && h.Cfg.AppURL != "" && h.Cfg.SMTP.FromEmail != "" {
		_ = h.M.Send(u.Email, "Verify your email", h.Cfg.AppURL+"/api/auth/verify?token="+tok)
	}
	c.JSON(http.StatusCreated, resp.CreatedID{ID: u.ID})
}

// @Summary      Login with email/password (+TOTP if enabled)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  request.Login  true  "login"
// @Success      200    {object}  response.TokenPair
// @Failure      400    {object}  response.Error
// @Failure      401    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /auth/login [post]
func (h *AuthController) login(c *gin.Context) {
	var r req.Login
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	u, err := h.S.VerifyPasswordLogin(c.Request.Context(), r.Email, r.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp.Error{Error: "invalid login"})
		return
	}
	if u.TOTPEnabled {
		if u.TOTPSecret == nil || !h.S.VerifyTOTP(*u.TOTPSecret, r.TOTP, time.Now()) {
			c.JSON(http.StatusUnauthorized, resp.Error{Error: "invalid totp"})
			return
		}
	}
	acc, ref, err := h.S.IssueTokens(c.Request.Context(), u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.TokenPair{Access: acc, Refresh: ref})
}

// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  request.Refresh  true  "refresh"
// @Success      200    {object}  response.TokenPair
// @Failure      400    {object}  response.Error
// @Failure      401    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /auth/refresh [post]
func (h *AuthController) refresh(c *gin.Context) {
	var r req.Refresh
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	// Validate refresh via Redis token store
	userID, err := h.Tokens.Get(c.Request.Context(), "refresh:"+r.Refresh)
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp.Error{Error: "invalid refresh"})
		return
	}
	u, err := repository.NewUserRepository(h.DB).ByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp.Error{Error: "invalid refresh"})
		return
	}
	acc, ref, err := h.S.IssueTokens(c.Request.Context(), u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	_ = h.Tokens.Del(c.Request.Context(), "refresh:"+r.Refresh)
	c.JSON(http.StatusOK, resp.TokenPair{Access: acc, Refresh: ref})
}

// @Summary      Verify email address
// @Tags         auth
// @Produce      json
// @Param        token  query  string  true  "verification token"
// @Success      200    {object}  response.OK
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /auth/verify [get]
func (h *AuthController) verifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "missing token"})
		return
	}
	userID, err := h.Tokens.Get(c.Request.Context(), "verify:"+token)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid or expired token"})
		return
	}
	if err := h.DB.Model(&m.User{}).Where("id = ?", userID).Update("status", "active").Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	_ = h.Tokens.Del(c.Request.Context(), "verify:"+token)
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

// @Summary      Request password reset
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  request.RequestReset  true  "email"
// @Success      200    {object}  response.OK
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /auth/request-reset [post]
func (h *AuthController) requestReset(c *gin.Context) {
	var r req.RequestReset
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	var u m.User
	if err := h.DB.Where("email = ?", r.Email).First(&u).Error; err != nil {
		// don't leak existence
		c.JSON(http.StatusOK, resp.OK{OK: true})
		return
	}
	tok, err := secureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if err := h.Tokens.Set(c.Request.Context(), "reset:"+tok, u.ID, 2*time.Hour); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if h.M != nil && h.Cfg.AppURL != "" && h.Cfg.SMTP.FromEmail != "" {
		_ = h.M.Send(u.Email, "Reset password", h.Cfg.AppURL+"/api/auth/reset?token="+tok)
	}
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

// @Summary      Reset password using token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  request.ResetPassword  true  "reset"
// @Success      200    {object}  response.OK
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /auth/reset [post]
func (h *AuthController) resetPassword(c *gin.Context) {
	var r req.ResetPassword
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: err.Error()})
		return
	}
	userID, err := h.Tokens.Get(c.Request.Context(), "reset:"+r.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid or expired token"})
		return
	}
	var u m.User
	if err := h.DB.First(&u, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusBadRequest, resp.Error{Error: "invalid token"})
		return
	}
	hash, err := h.S.HashPassword(r.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	if err := h.DB.Model(&u).Updates(map[string]any{"password_hash": hash}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error{Error: err.Error()})
		return
	}
	// Revoke all refresh tokens for user by scanning keys is out of scope here; rely on client to refresh after reset
	_ = h.Tokens.Del(c.Request.Context(), "reset:"+r.Token)
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

// @Summary      Get current user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200    {object}  response.User
// @Failure      401    {object}  response.Error
// @Router       /auth/me [get]
func (h *AuthController) me(c *gin.Context) {
	if v, ok := c.Get("user"); ok {
		u := v.(*m.User)
		c.JSON(http.StatusOK, resp.User{ID: u.ID, Email: u.Email, FullName: u.FullName, Role: u.Role, Status: u.Status, TOTPEnabled: u.TOTPEnabled})
		return
	}
	c.JSON(http.StatusUnauthorized, resp.Error{Error: "unauthorized"})
}

// @Summary      Logout (revoke refresh token)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body  request.Logout  true  "logout"
// @Success      200    {object}  response.OK
// @Failure      401    {object}  response.Error
// @Router       /auth/logout [post]
func (h *AuthController) logout(c *gin.Context) {
	var r req.Logout
	_ = c.ShouldBindJSON(&r)
	if r.Refresh != "" {
		_ = h.Tokens.Del(c.Request.Context(), "refresh:"+r.Refresh)
	}
	c.JSON(http.StatusOK, resp.OK{OK: true})
}

func secureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
