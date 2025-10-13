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
	"gradeflow/internal/repository"
	"gradeflow/internal/service"
)

type enrollTestCtx struct {
	r   *gin.Engine
	jwt string
	db  *gorm.DB
}

func setupEnrollTest(t *testing.T) enrollTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:enrolldb_%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	must := func(err error) {
		if err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	must(db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, email TEXT UNIQUE, full_name TEXT, password_hash TEXT, role TEXT, status TEXT, totp_secret TEXT, totp_enabled INTEGER, last_login_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS students (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, individual_number TEXT UNIQUE, full_name TEXT, group_id TEXT, user_id TEXT, start_year INTEGER, end_year INTEGER)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS courses (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, subject_id TEXT, department_id TEXT, program_id TEXT, academic_session_id TEXT, title TEXT, teacher_id TEXT, room TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS enrollments (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, student_id TEXT)`).Error)
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','enroll-teacher@example.com','Teacher','x','teacher','active',0)`).Error)
	must(db.Exec(`INSERT INTO students (id,individual_number,full_name) VALUES ('st1','S0001','Alice')`).Error)
	must(db.Exec(`INSERT INTO courses (id,title) VALUES ('c1','Algorithms 101')`).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	enrollmentRepo := repository.NewEnrollmentRepository(db)
	enrollmentSvc := service.NewEnrollmentService(enrollmentRepo)
	ctrl := ctr.NewEnrollmentController(db, cfg, enrollmentSvc)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return enrollTestCtx{r: r, jwt: s, db: db}
}

func TestEnrollmentCRUD(t *testing.T) {
	ctx := setupEnrollTest(t)
	// create
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enrollments", strings.NewReader(`{"courseId":"c1","studentId":"st1"}`))
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
	req = httptest.NewRequest("GET", "/api/enrollments", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got=%d", w.Code)
	}

	// get
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/enrollments/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get expected 200 got=%d", w.Code)
	}

	// update
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/enrollments/"+id, strings.NewReader(`{"studentId":"st1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200 got=%d body=%s", w.Code, w.Body.String())
	}

	// delete
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/enrollments/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got=%d", w.Code)
	}
}

func TestEnrollment_UnauthorizedWithoutJWT(t *testing.T) {
	ctx := setupEnrollTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/enrollments", nil)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestEnrollment_ForbiddenForStudentOnCreate(t *testing.T) {
	ctx := setupEnrollTest(t)
	if err := ctx.db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u2','stud@example.com','Stud','x','student','active',0)`).Error; err != nil {
		t.Fatalf("seed student: %v", err)
	}
	cfg := config.Config{JWTSecret: "test"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u2", "exp": time.Now().Add(15 * time.Minute).Unix()})
	studJWT, _ := token.SignedString([]byte(cfg.JWTSecret))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enrollments", strings.NewReader(`{"courseId":"c1","studentId":"st1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+studJWT)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestEnrollment_CreateInvalidJSONReturns400(t *testing.T) {
	ctx := setupEnrollTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/enrollments", strings.NewReader(`{"courseId":"c1"`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestEnrollment_GetNotFound(t *testing.T) {
	ctx := setupEnrollTest(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/enrollments/does-not-exist", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got=%d body=%s", w.Code, w.Body.String())
	}
}
