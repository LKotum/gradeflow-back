package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	ctr "gradeflow/internal/controllers"
)

type reportTestCtx struct {
	r   *gin.Engine
	jwt string
}

func setupReportTest(t *testing.T) reportTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:reportdb?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	must := func(err error) {
		if err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	must(db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, email TEXT UNIQUE, full_name TEXT, password_hash TEXT, role TEXT, status TEXT, totp_secret TEXT, totp_enabled INTEGER, last_login_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS groups (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, program_id TEXT, code TEXT, name TEXT, year INTEGER)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS students (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, individual_number TEXT, full_name TEXT, group_id TEXT, user_id TEXT, start_year INTEGER, end_year INTEGER)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS courses (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, subject_id TEXT, department_id TEXT, program_id TEXT, academic_session_id TEXT, title TEXT, teacher_id TEXT, room TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS lessons (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, starts_at DATETIME, ends_at DATETIME, room TEXT, kind TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessments (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, type TEXT, date_at DATETIME, room TEXT, scale TEXT, max_pts REAL)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessment_grades (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, assessment_id TEXT, student_id TEXT, scale TEXT, value_num REAL, value_pass INTEGER, graded_by TEXT, graded_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS attendances (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, lesson_id TEXT, student_id TEXT, status TEXT, marked_by TEXT, marked_at DATETIME)`).Error)

	now := time.Now().UTC()
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','report-dean@example.com','Dean','x','dean','active',0)`).Error)
	must(db.Exec(`INSERT INTO groups (id,program_id,code,name,year) VALUES ('g1','p1','G1','Group 1',2024)`).Error)
	must(db.Exec(`INSERT INTO students (id,individual_number,full_name,group_id,start_year,end_year) VALUES ('st1','S0001','Student One','g1',2023,2027)`).Error)
	must(db.Exec(`INSERT INTO courses (id,academic_session_id,title) VALUES ('c1','sess1','Algorithms')`).Error)
	must(db.Exec(`INSERT INTO lessons (id,course_id,starts_at,ends_at,kind) VALUES ('l1','c1',?,?,'lecture')`, now, now.Add(90*time.Minute)).Error)
	must(db.Exec(`INSERT INTO assessments (id,course_id,type,date_at,scale,max_pts) VALUES ('a1','c1','exam',?, 'hundred',100)`, now).Error)
	must(db.Exec(`INSERT INTO assessment_grades (id,assessment_id,student_id,scale,value_num,value_pass,graded_by,graded_at) VALUES ('g11','a1','st1','hundred',85,NULL,'u1',?)`, now).Error)
	must(db.Exec(`INSERT INTO attendances (id,lesson_id,student_id,status,marked_by,marked_at) VALUES ('att1','l1','st1','present','u1',?),('att2','l1','st1','absent','u1',?)`, now, now.Add(24*time.Hour)).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	ctrl := ctr.NewReportController(db, cfg)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return reportTestCtx{r: r, jwt: s}
}

func TestReportStudentCourseSessionSummary(t *testing.T) {
	ctx := setupReportTest(t)

	doReq := func(path string) map[string]any {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer "+ctx.jwt)
		ctx.r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s got=%d body=%s", path, w.Code, w.Body.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal response for %s: %v", path, err)
		}
		return payload
	}

	student := doReq("/api/reports/students/st1/summary")
	if avg, ok := student["averageGrade"].(float64); !ok || avg < 84.9 || avg > 85.1 {
		t.Fatalf("unexpected student avg grade: %v", student)
	}
	if rate, ok := student["attendanceRate"].(float64); !ok || rate < 0.49 || rate > 0.51 {
		t.Fatalf("unexpected student attendance rate: %v", student)
	}

	course := doReq("/api/reports/courses/c1/summary")
	if count, ok := course["studentCount"].(float64); !ok || int(count) != 1 {
		t.Fatalf("unexpected course student count: %v", course)
	}

	session := doReq("/api/reports/sessions/sess1/summary")
	if courses, ok := session["courseCount"].(float64); !ok || int(courses) != 1 {
		t.Fatalf("unexpected session course count: %v", session)
	}
}
