package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	ctr "gradeflow/internal/controllers"
	m "gradeflow/internal/domain/models"
	"gradeflow/pkg/utils"
)

func setupAuthTest(t *testing.T) (*gin.Engine, *gorm.DB, *utils.MemoryTokenStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// SQLite schema compatible with GORM Base fields
	if err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            deleted_at DATETIME NULL,
            email TEXT NOT NULL UNIQUE,
            full_name TEXT NOT NULL,
            password_hash TEXT NOT NULL,
            role TEXT NOT NULL,
            status TEXT NOT NULL,
            totp_secret TEXT NULL,
            totp_enabled INTEGER NOT NULL DEFAULT 0,
            last_login_at DATETIME NULL
        );
    `).Error; err != nil {
		t.Fatalf("create users table: %v", err)
	}
	if err := db.Exec(`
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            deleted_at DATETIME NULL,
            user_id TEXT NOT NULL,
            token TEXT NOT NULL UNIQUE,
            expires_at DATETIME NOT NULL,
            revoked INTEGER NOT NULL DEFAULT 0
        );
    `).Error; err != nil {
		t.Fatalf("create refresh_tokens table: %v", err)
	}

	cfg := config.Config{AppURL: "http://localhost:8080", JWTSecret: "testsecret", AccessTTL: 15 * time.Minute, RefreshTTL: 24 * time.Hour}
	mem := utils.NewMemoryTokenStore()
	c := ctr.NewAuthController(db, cfg, mem)
	r := gin.New()
	api := r.Group("/api")
	c.RegisterRoutes(api.Group("/auth"))
	return r, db, mem
}

func performReq(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterVerifyLoginFlow(t *testing.T) {
	r, db, mem := setupAuthTest(t)

	regBody := `{"email":"test@example.com","fullName":"Test User","role":"student","password":"Password123"}`
	w := performReq(r, "POST", "/api/auth/register", regBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("register code=%d body=%s", w.Code, w.Body.String())
	}

	var verifyToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "verify:") {
			verifyToken = strings.TrimPrefix(k, "verify:")
			break
		}
	}
	if verifyToken == "" {
		t.Fatalf("expected verify token")
	}

	w = performReq(r, "GET", "/api/auth/verify?token="+verifyToken, "")
	if w.Code != http.StatusOK {
		t.Fatalf("verify code=%d body=%s", w.Code, w.Body.String())
	}

	loginBody := `{"email":"test@example.com","password":"Password123"}`
	w = performReq(r, "POST", "/api/auth/login", loginBody)
	if w.Code != http.StatusOK {
		t.Fatalf("login code=%d body=%s", w.Code, w.Body.String())
	}

	var pair struct{ Access, Refresh string }
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil {
		t.Fatalf("unmarshal token pair: %v", err)
	}
	if pair.Access == "" || pair.Refresh == "" {
		t.Fatalf("missing tokens in response")
	}

	var u m.User
	if err := db.First(&u, "email = ?", "test@example.com").Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if u.Status != "active" {
		t.Fatalf("expected active status, got %s", u.Status)
	}
}

func TestResetPasswordFlow(t *testing.T) {
	r, _, mem := setupAuthTest(t)

	_ = performReq(r, "POST", "/api/auth/register", `{"email":"r@example.com","fullName":"R","role":"student","password":"OldPass123"}`)
	var verifyToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "verify:") {
			verifyToken = strings.TrimPrefix(k, "verify:")
			break
		}
	}
	_ = performReq(r, "GET", "/api/auth/verify?token="+verifyToken, "")

	w := performReq(r, "POST", "/api/auth/request-reset", `{"email":"r@example.com"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("request-reset code=%d body=%s", w.Code, w.Body.String())
	}

	var resetToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "reset:") {
			resetToken = strings.TrimPrefix(k, "reset:")
			break
		}
	}
	if resetToken == "" {
		t.Fatalf("expected reset token")
	}

	w = performReq(r, "POST", "/api/auth/reset", `{"token":"`+resetToken+`","password":"NewPass456"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("reset code=%d body=%s", w.Code, w.Body.String())
	}

	w = performReq(r, "POST", "/api/auth/login", `{"email":"r@example.com","password":"NewPass456"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login after reset code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyEmail_InvalidAndExpiredToken(t *testing.T) {
	r, _, mem := setupAuthTest(t)
	w := performReq(r, "GET", "/api/auth/verify?token=doesnotexist", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("verify invalid token expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
	tok := "expiredTok"
	_ = mem.Set(context.Background(), "verify:"+tok, "some-user-id", -1*time.Second)
	w = performReq(r, "GET", "/api/auth/verify?token="+tok, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("verify expired token expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestResetPassword_InvalidAndExpiredToken(t *testing.T) {
	r, _, mem := setupAuthTest(t)
	w := performReq(r, "POST", "/api/auth/reset", `{"token":"nope","password":"SomePass123"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("reset invalid token expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
	tok := "expiredReset"
	_ = mem.Set(context.Background(), "reset:"+tok, "some-user-id", -1*time.Second)
	w = performReq(r, "POST", "/api/auth/reset", `{"token":"`+tok+`","password":"SomePass123"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("reset expired token expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	r, _, mem := setupAuthTest(t)
	_ = performReq(r, "POST", "/api/auth/register", `{"email":"wp@example.com","fullName":"WP","role":"student","password":"CorrectPass1"}`)
	var verifyToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "verify:") {
			verifyToken = strings.TrimPrefix(k, "verify:")
			break
		}
	}
	_ = performReq(r, "GET", "/api/auth/verify?token="+verifyToken, "")
	w := performReq(r, "POST", "/api/auth/login", `{"email":"wp@example.com","password":"WrongPass"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("login wrong password expected 401 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestLogin_UnverifiedUser(t *testing.T) {
	r, _, _ := setupAuthTest(t)
	_ = performReq(r, "POST", "/api/auth/register", `{"email":"uv@example.com","fullName":"UV","role":"student","password":"SomePass123"}`)
	w := performReq(r, "POST", "/api/auth/login", `{"email":"uv@example.com","password":"SomePass123"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("login unverified expected 401 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	r, _, _ := setupAuthTest(t)
	w := performReq(r, "POST", "/api/auth/refresh", `{"refresh":"bogus"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("refresh invalid expected 401 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRefresh_IssueNewTokens(t *testing.T) {
	r, _, mem := setupAuthTest(t)
	// Register, verify, login
	_ = performReq(r, "POST", "/api/auth/register", `{"email":"rf@example.com","fullName":"RF","role":"student","password":"P455word!!"}`)
	var verifyToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "verify:") {
			verifyToken = strings.TrimPrefix(k, "verify:")
			break
		}
	}
	_ = performReq(r, "GET", "/api/auth/verify?token="+verifyToken, "")
	w := performReq(r, "POST", "/api/auth/login", `{"email":"rf@example.com","password":"P455word!!"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login code=%d body=%s", w.Code, w.Body.String())
	}
	var pair struct{ Access, Refresh string }
	_ = json.Unmarshal(w.Body.Bytes(), &pair)
	// Refresh
	w = performReq(r, "POST", "/api/auth/refresh", `{"refresh":"`+pair.Refresh+`"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("refresh expected 200 got=%d body=%s", w.Code, w.Body.String())
	}
	var pair2 struct{ Access, Refresh string }
	_ = json.Unmarshal(w.Body.Bytes(), &pair2)
	if pair2.Access == "" || pair2.Refresh == "" {
		t.Fatalf("missing tokens after refresh")
	}
}

func TestMe_WithoutJWT_Unauthorized(t *testing.T) {
	r, _, _ := setupAuthTest(t)
	w := performReq(r, "GET", "/api/auth/me", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("me without jwt expected 401 got=%d", w.Code)
	}
}

func TestMe_WithJWT_Succeeds(t *testing.T) {
	r, _, mem := setupAuthTest(t)
	// Register, verify, login
	_ = performReq(r, "POST", "/api/auth/register", `{"email":"me@example.com","fullName":"ME","role":"student","password":"MeMeMe123"}`)
	var verifyToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "verify:") {
			verifyToken = strings.TrimPrefix(k, "verify:")
			break
		}
	}
	_ = performReq(r, "GET", "/api/auth/verify?token="+verifyToken, "")
	w := performReq(r, "POST", "/api/auth/login", `{"email":"me@example.com","password":"MeMeMe123"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login code=%d body=%s", w.Code, w.Body.String())
	}
	var pair struct{ Access, Refresh string }
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// /me with bearer
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+pair.Access)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("me expected 200 got=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestLogout_RevokesRefresh(t *testing.T) {
	r, _, mem := setupAuthTest(t)
	_ = performReq(r, "POST", "/api/auth/register", `{"email":"lo@example.com","fullName":"LO","role":"student","password":"LogOut123"}`)
	var verifyToken string
	for _, k := range mem.Keys() {
		if strings.HasPrefix(k, "verify:") {
			verifyToken = strings.TrimPrefix(k, "verify:")
			break
		}
	}
	_ = performReq(r, "GET", "/api/auth/verify?token="+verifyToken, "")
	w := performReq(r, "POST", "/api/auth/login", `{"email":"lo@example.com","password":"LogOut123"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login code=%d body=%s", w.Code, w.Body.String())
	}
	var pair struct{ Access, Refresh string }
	_ = json.Unmarshal(w.Body.Bytes(), &pair)
	// logout (protected endpoint, include bearer)
	req := httptest.NewRequest("POST", "/api/auth/logout", strings.NewReader(`{"refresh":"`+pair.Refresh+`"}`))
	req.Header.Set("Authorization", "Bearer "+pair.Access)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("logout expected 200 got=%d body=%s", w.Code, w.Body.String())
	}
	// try refresh again should fail
	w = performReq(r, "POST", "/api/auth/refresh", `{"refresh":"`+pair.Refresh+`"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout expected 401 got=%d body=%s", w.Code, w.Body.String())
	}
}
