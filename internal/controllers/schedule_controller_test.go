package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	ctr "gradeflow/internal/controllers"
)

type scheduleTestCtx struct {
	r   *gin.Engine
	jwt string
	db  *gorm.DB
}

func setupScheduleTest(t *testing.T) scheduleTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:scheduledb_%s?mode=memory&cache=shared", t.Name())
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
	must(db.Exec(`CREATE TABLE IF NOT EXISTS courses (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, subject_id TEXT, department_id TEXT, program_id TEXT, academic_session_id TEXT, title TEXT, teacher_id TEXT, room TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS enrollments (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, student_id TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS lessons (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, starts_at DATETIME, ends_at DATETIME, room TEXT, kind TEXT)`).Error)
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','schedule-student@example.com','Student','x','student','active',0)`).Error)
	must(db.Exec(`INSERT INTO courses (id,title) VALUES ('c1','Algorithms 101')`).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	ctrl := ctr.NewScheduleController(db, cfg)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return scheduleTestCtx{r: r, jwt: s, db: db}
}

func TestSchedule_EmptyWhenNoEnrollments(t *testing.T) {
	ctx := setupScheduleTest(t)
	q := url.Values{"studentId": []string{"st1"}}
	req := httptest.NewRequest("GET", "/api/schedule/student?"+q.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	w := httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got=%d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		Items []map[string]any `json:"items"`
		Page  map[string]any   `json:"page"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Items) != 0 {
		t.Fatalf("expected empty items, got %d", len(out.Items))
	}
}

func TestSchedule_WithLessonsAndFilter(t *testing.T) {
	ctx := setupScheduleTest(t)
	// seed enrollment and two lessons
	must := func(err error) {
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	// enroll st1 to c1
	must(ctx.db.Exec(`INSERT INTO enrollments (course_id, student_id) VALUES (?, ?)`, "c1", "st1").Error)
	now := time.Now().UTC()
	l0s := now.Add(-48 * time.Hour)
	l0e := now.Add(-47 * time.Hour)
	l1s := now.Add(1 * time.Hour)
	l1e := now.Add(2 * time.Hour)
	must(ctx.db.Exec(`INSERT INTO lessons (course_id, starts_at, ends_at, room, kind) VALUES (?, ?, ?, ?, ?)`, "c1", l0s, l0e, "R0", "seminar").Error)
	must(ctx.db.Exec(`INSERT INTO lessons (course_id, starts_at, ends_at, room, kind) VALUES (?, ?, ?, ?, ?)`, "c1", l1s, l1e, "R1", "lecture").Error)

	from := now.Format(time.RFC3339)
	to := now.Add(24 * time.Hour).Format(time.RFC3339)
	q := url.Values{"studentId": []string{"st1"}, "from": []string{from}, "to": []string{to}}
	req := httptest.NewRequest("GET", "/api/schedule/student?"+q.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	w := httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got=%d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		Items []map[string]any `json:"items"`
		Page  map[string]any   `json:"page"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("expected 1 item, got %d: %s", len(out.Items), w.Body.String())
	}
	it := out.Items[0]
	if it["courseId"] != "c1" {
		t.Fatalf("unexpected courseId: %v", it["courseId"])
	}
	if it["courseTitle"] != "Algorithms 101" {
		t.Fatalf("unexpected courseTitle: %v", it["courseTitle"])
	}
}

func TestSchedule_MissingStudentIdReturns400(t *testing.T) {
	ctx := setupScheduleTest(t)
	req := httptest.NewRequest("GET", "/api/schedule/student", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	w := httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSchedule_UnauthorizedWithoutJWT(t *testing.T) {
	ctx := setupScheduleTest(t)
	req := httptest.NewRequest("GET", "/api/schedule/student?studentId=st1", nil)
	w := httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got=%d", w.Code)
	}
}
