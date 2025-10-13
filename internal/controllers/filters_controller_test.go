package controllers_test

import (
    "encoding/json"
    "fmt"
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
    "gradeflow/internal/repository"
    "gradeflow/internal/service"
)

// quick helper to sign jwt
func signJWT(t *testing.T, secret string, sub string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": sub, "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(secret))
	return s
}

func TestFilters_Lessons_Attendance_Assessments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:filtersdb_%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	must := func(err error) {
		if err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	// schema
	must(db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, email TEXT UNIQUE, full_name TEXT, password_hash TEXT, role TEXT, status TEXT, totp_secret TEXT, totp_enabled INTEGER, last_login_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS courses (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, subject_id TEXT, department_id TEXT, program_id TEXT, academic_session_id TEXT, title TEXT, teacher_id TEXT, room TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS lessons (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, starts_at DATETIME, ends_at DATETIME, room TEXT, kind TEXT)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS attendances (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, lesson_id TEXT, student_id TEXT, status TEXT, marked_by TEXT, marked_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessments (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, type TEXT, date_at DATETIME, room TEXT, scale TEXT, max_pts REAL)`).Error)
	// seed
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','filters@example.com','User','x','teacher','active',0)`).Error)
	must(db.Exec(`INSERT INTO courses (id,academic_session_id,title) VALUES ('c1','s1','Course 1'),('c2','s2','Course 2')`).Error)
	now := time.Now().UTC()
	// lessons: l1 in s1 now+1h..+2h for c1, l2 past for c1, l3 for c2 in s2 within window
	must(db.Exec(`INSERT INTO lessons (id,course_id,starts_at,ends_at,kind) VALUES ('l1','c1',?,?,'lecture')`, now.Add(1*time.Hour), now.Add(2*time.Hour)).Error)
	must(db.Exec(`INSERT INTO lessons (id,course_id,starts_at,ends_at,kind) VALUES ('l2','c1',?,?,'lecture')`, now.Add(-48*time.Hour), now.Add(-47*time.Hour)).Error)
	must(db.Exec(`INSERT INTO lessons (id,course_id,starts_at,ends_at,kind) VALUES ('l3','c2',?,?,'seminar')`, now.Add(1*time.Hour), now.Add(90*time.Minute)).Error)
	// attendance: for l1 (st1 present now), for l3 (st2 absent now)
	must(db.Exec(`INSERT INTO attendances (id,lesson_id,student_id,status,marked_at) VALUES ('a1','l1','st1','present',?),('a2','l3','st2','absent',?)`, now, now).Error)
	// assessments: a1 for c1 yesterday, a2 for c2 tomorrow
	must(db.Exec(`INSERT INTO assessments (id,course_id,type,date_at,scale) VALUES ('a1','c1','exam',?, 'points'),('a2','c2','test',?, 'points')`, now.Add(-24*time.Hour), now.Add(24*time.Hour)).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	jwt := signJWT(t, cfg.JWTSecret, "u1")

	r := gin.New()
	lessonRepo := repository.NewLessonRepository(db)
	lessonSvc := service.NewLessonService(lessonRepo)
	attendanceRepo := repository.NewAttendanceRepository(db)
	attendanceSvc := service.NewAttendanceService(attendanceRepo)
	assessmentRepo := repository.NewAssessmentRepository(db)
	assessmentSvc := service.NewAssessmentService(assessmentRepo)
	gradeRepo := repository.NewAssessmentGradeRepository(db)
	gradeSvc := service.NewAssessmentGradeService(gradeRepo, assessmentRepo)
	ctr.NewLessonController(db, cfg, lessonSvc, attendanceSvc).RegisterRoutes(r.Group("/api"))
	ctr.NewAttendanceController(db, cfg, attendanceSvc).RegisterRoutes(r.Group("/api"))
	ctr.NewAssessmentController(db, cfg, assessmentSvc, gradeSvc).RegisterRoutes(r.Group("/api"))

	// 1) lessons by course+date window -> expect only l1 (c1 in future window)
	v := url.Values{"courseId": {"c1"}, "from": {now.Format(time.RFC3339)}, "to": {now.Add(3 * time.Hour).Format(time.RFC3339)}}
	req := httptest.NewRequest("GET", "/api/lessons?"+v.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || !itemsCountIs(w.Body.Bytes(), 1) {
		t.Fatalf("lessons filter failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 2) lessons by session (s2) -> expect l3 only
	v = url.Values{"sessionId": {"s2"}}
	req = httptest.NewRequest("GET", "/api/lessons?"+v.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || !itemsCountIs(w.Body.Bytes(), 1) {
		t.Fatalf("lessons session filter failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 3) attendance by course and date window -> expect 1 (a1 for c1)
	v = url.Values{"courseId": {"c1"}, "from": {now.Add(-1 * time.Hour).Format(time.RFC3339)}, "to": {now.Add(1 * time.Hour).Format(time.RFC3339)}}
	req = httptest.NewRequest("GET", "/api/attendance?"+v.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || !itemsCountIs(w.Body.Bytes(), 1) {
		t.Fatalf("attendance filter failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 4) assessments by date window (future only) -> expect 1 (a2)
	v = url.Values{"from": {now.Format(time.RFC3339)}}
	req = httptest.NewRequest("GET", "/api/assessments?"+v.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 || !itemsCountIs(w.Body.Bytes(), 1) {
		t.Fatalf("assessments date filter failed: code=%d body=%s", w.Code, w.Body.String())
	}
}

// itemsCountIs unmarshals list response and checks len(items)
func itemsCountIs(body []byte, exp int) bool {
	var out struct {
		Items []any `json:"items"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return false
	}
	return len(out.Items) == exp
}
