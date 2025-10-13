package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	ctr "gradeflow/internal/controllers"
)

type deptTestCtx struct {
	r   *gin.Engine
	jwt string
	db  *gorm.DB
}

func setupDeptTest(t *testing.T) deptTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:deptdb_%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
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
		t.Fatalf("create users: %v", err)
	}
	if err := db.Exec(`
        CREATE TABLE IF NOT EXISTS departments (
            id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))),
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            deleted_at DATETIME NULL,
            code TEXT NOT NULL UNIQUE,
            name TEXT NOT NULL
        );
    `).Error; err != nil {
		t.Fatalf("create departments: %v", err)
	}

	// seed user for JWT middleware
	if err := db.Exec(`
		INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled)
		VALUES ('u1','dean@example.com','Dean','x','dean','active',0)
	`).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	dept := ctr.NewDepartmentController(db, cfg)
	grp := r.Group("/api")
	dept.RegisterRoutes(grp)

	// create JWT (dean)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "u1",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})
	s, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}
	return deptTestCtx{r: r, jwt: s, db: db}
}

func TestDepartmentCRUD(t *testing.T) {
	ctx := setupDeptTest(t)
	// create
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/departments", strings.NewReader(`{"code":"CS","name":"Computer Science"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expected 201 got=%d body=%s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	idVal, ok := created["id"]
	if !ok {
		t.Fatalf("create response missing id field: %v", created)
	}
	id, ok := idVal.(string)
	if !ok || id == "" {
		t.Fatalf("create response id not a non-empty string: %v (%T)", idVal, idVal)
	}

	// list
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/departments", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got=%d", w.Code)
	}

	// get
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/departments/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get expected 200 got=%d", w.Code)
	}

	// update
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/departments/"+id, strings.NewReader(`{"name":"CS Dept"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200 got=%d body=%s", w.Code, w.Body.String())
	}

	// delete
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/departments/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got=%d", w.Code)
	}
}

func TestDepartment_UnauthorizedWithoutJWT(t *testing.T) {
	ctx := setupDeptTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/departments", nil)
	// no Authorization header
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDepartment_ForbiddenForStudentOnCreate(t *testing.T) {
	ctx := setupDeptTest(t)
	// seed a student user and sign JWT for them
	if err := ctx.db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u2','stud@example.com','Stud','x','student','active',0)`).Error; err != nil {
		t.Fatalf("seed student: %v", err)
	}
	cfg := config.Config{JWTSecret: "test"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u2", "exp": time.Now().Add(15 * time.Minute).Unix()})
	studJWT, _ := token.SignedString([]byte(cfg.JWTSecret))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/departments", strings.NewReader(`{"code":"XX","name":"Name"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+studJWT)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDepartment_CreateInvalidJSONReturns400(t *testing.T) {
	ctx := setupDeptTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/departments", strings.NewReader(`{"code":"broken"`)) // invalid JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
}
