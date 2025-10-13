package controllers_test

import (
	"encoding/json"
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

type courseTestCtx struct {
	r   *gin.Engine
	jwt string
}

func setupCourseTest(t *testing.T) courseTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:coursedb?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	must := func(err error) {
		if err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	must(db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, email TEXT UNIQUE, full_name TEXT, password_hash TEXT, role TEXT, status TEXT, totp_secret TEXT, totp_enabled INTEGER, last_login_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS departments (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, code TEXT UNIQUE, name TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS programs (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, department_id TEXT, code TEXT UNIQUE, name TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS subjects (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, department_id TEXT, code TEXT UNIQUE, name TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS academic_sessions (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, code TEXT UNIQUE, name TEXT, starts_at DATETIME, ends_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS courses (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, subject_id TEXT, department_id TEXT, program_id TEXT, academic_session_id TEXT, title TEXT, teacher_id TEXT, room TEXT)`).Error)
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','course-dean@example.com','Dean','x','dean','active',0)`).Error)
	must(db.Exec(`INSERT INTO departments (id,code,name) VALUES ('d1','CS','Computer Science')`).Error)
	must(db.Exec(`INSERT INTO subjects (id,department_id,code,name) VALUES ('s1','d1','ALG','Algorithms')`).Error)
	must(db.Exec(`INSERT INTO academic_sessions (id,code,name) VALUES ('as1','2025F','Fall 2025')`).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	courseRepo := repository.NewCourseRepository(db)
	courseSvc := service.NewCourseService(courseRepo)
	ctrl := ctr.NewCourseController(db, cfg, courseSvc)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return courseTestCtx{r: r, jwt: s}
}

func TestCourseCRUD(t *testing.T) {
	ctx := setupCourseTest(t)
	// create
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/courses", strings.NewReader(`{"subjectId":"s1","departmentId":"d1","academicSessionId":"as1","title":"Algorithms 101","teacherId":"t1","room":"R1"}`))
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
	req = httptest.NewRequest("GET", "/api/courses", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got=%d", w.Code)
	}

	// get
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/courses/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get expected 200 got=%d", w.Code)
	}

	// update
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/courses/"+id, strings.NewReader(`{"title":"Algorithms I","room":"R2"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200 got=%d body=%s", w.Code, w.Body.String())
	}

	// delete
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/courses/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got=%d", w.Code)
	}
}
