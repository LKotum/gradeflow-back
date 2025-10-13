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

type assessTestCtx struct {
	r   *gin.Engine
	jwt string
}

func setupAssessTest(t *testing.T) assessTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:assessdb?mode=memory&cache=shared"), &gorm.Config{})
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
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessments (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, course_id TEXT, type TEXT, date_at DATETIME, room TEXT, scale TEXT, max_pts REAL)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS assessment_grades (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, assessment_id TEXT, student_id TEXT, scale TEXT, value_num REAL, value_pass INTEGER, graded_by TEXT, graded_at DATETIME)`).Error)
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','assess-teacher@example.com','Teacher','x','teacher','active',0)`).Error)
	must(db.Exec(`INSERT INTO courses (id,title) VALUES ('c1','Algorithms 101')`).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	assessmentRepo := repository.NewAssessmentRepository(db)
	gradeRepo := repository.NewAssessmentGradeRepository(db)
	assessmentSvc := service.NewAssessmentService(assessmentRepo)
	gradeSvc := service.NewAssessmentGradeService(gradeRepo, assessmentRepo)
	ctrl := ctr.NewAssessmentController(db, cfg, assessmentSvc, gradeSvc)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return assessTestCtx{r: r, jwt: s}
}

func TestAssessmentCRUDAndBulkGrades(t *testing.T) {
	ctx := setupAssessTest(t)
	date := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	// create assessment
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/assessments", strings.NewReader(`{"courseId":"c1","type":"exam","dateAt":"`+date+`","room":"R1","scale":"hundred","maxPts":100}`))
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

	// list assessments
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/assessments", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got=%d", w.Code)
	}

	// bulk grades
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/assessments/"+id+"/grades/bulk", strings.NewReader(`[{"studentId":"st1","valueNum":90},{"studentId":"st2","scale":"passfail","valuePass":true}]`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("bulk expected 200 got=%d body=%s", w.Code, w.Body.String())
	}
}
