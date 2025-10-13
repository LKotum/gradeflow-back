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

type journalTestCtx struct {
	r   *gin.Engine
	jwt string
	db  *gorm.DB
}

func setupJournalTest(t *testing.T) journalTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:journaldb_%s?mode=memory&cache=shared", t.Name())
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
	must(db.Exec(`CREATE TABLE IF NOT EXISTS lessons (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, starts_at DATETIME, ends_at DATETIME, room TEXT, kind TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS attendances (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, lesson_id TEXT, student_id TEXT, status TEXT, marked_by TEXT, marked_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessments (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, type TEXT, date_at DATETIME, room TEXT, scale TEXT, max_pts REAL)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessment_grades (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, assessment_id TEXT, student_id TEXT, scale TEXT, value_num REAL, value_pass INTEGER, graded_by TEXT, graded_at DATETIME)`).Error)
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','journal-staff@example.com','Staff','x','dean','active',0)`).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	ctrl := ctr.NewJournalController(db, cfg)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return journalTestCtx{r: r, jwt: s, db: db}
}

func TestStudentJournal_HappyPathAndCourseFilter(t *testing.T) {
	ctx := setupJournalTest(t)
	// seed data: one course with lesson+attendance, one assessment+grade for student st1
	must := func(err error) {
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	must(ctx.db.Exec(`INSERT INTO lessons (id, course_id, starts_at, ends_at, room, kind) VALUES ('l1','c1', ?, ?, 'R1','lecture')`, time.Now().UTC().Add(-2*time.Hour), time.Now().UTC().Add(-1*time.Hour)).Error)
	must(ctx.db.Exec(`INSERT INTO attendances (lesson_id, student_id, status, marked_by, marked_at) VALUES ('l1','st1','present','u1', ?)`, time.Now().UTC()).Error)
	must(ctx.db.Exec(`INSERT INTO assessments (id, course_id, type, date_at, room, scale, max_pts) VALUES ('a1','c1','exam', ?, 'R2','points',100)`, time.Now().UTC().Add(-24*time.Hour)).Error)
	must(ctx.db.Exec(`INSERT INTO assessment_grades (assessment_id, student_id, scale, value_num, graded_by, graded_at) VALUES ('a1','st1','points',95,'u1', ?)`, time.Now().UTC()).Error)

	// call journal without filter
	req := httptest.NewRequest("GET", "/api/students/st1/journal", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	w := httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got=%d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		StudentID  string        `json:"studentId"`
		CourseID   *string       `json:"courseId"`
		Attendance []interface{} `json:"attendance"`
		Grades     []interface{} `json:"grades"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.StudentID != "st1" || len(out.Attendance) != 1 || len(out.Grades) != 1 {
		t.Fatalf("unexpected payload: %s", w.Body.String())
	}

	// with course filter
	q := url.Values{"courseId": []string{"c1"}}
	req = httptest.NewRequest("GET", "/api/students/st1/journal?"+q.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	w = httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with filter got=%d body=%s", w.Code, w.Body.String())
	}
}

func TestStudentJournal_UnauthorizedWithoutJWT(t *testing.T) {
	ctx := setupJournalTest(t)
	req := httptest.NewRequest("GET", "/api/students/st1/journal", nil)
	w := httptest.NewRecorder()
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got=%d", w.Code)
	}
}
