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
)

type studentTestCtx struct {
	r   *gin.Engine
	jwt string
}

func setupStudentTest(t *testing.T) studentTestCtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:studentdb?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	must := func(err error) {
		if err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	must(db.Exec(`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, email TEXT UNIQUE, full_name TEXT, password_hash TEXT, role TEXT, status TEXT, totp_secret TEXT, totp_enabled INTEGER, last_login_at DATETIME)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS groups (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, program_id TEXT, code TEXT UNIQUE, name TEXT, year INTEGER)`).Error)
	must(db.Exec(`CREATE TABLE IF NOT EXISTS students (id TEXT PRIMARY KEY NOT NULL DEFAULT (lower(hex(randomblob(16)))), created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, individual_number TEXT UNIQUE, full_name TEXT, group_id TEXT, user_id TEXT)`).Error)
	must(db.Exec(`INSERT INTO users (id,email,full_name,password_hash,role,status,totp_enabled) VALUES ('u1','student-dean@example.com','Dean','x','dean','active',0)`).Error)
	must(db.Exec(`INSERT INTO groups (id,program_id,code,name,year) VALUES ('g1','p1','G1','Group 1',2025)`).Error)

	cfg := config.Config{JWTSecret: "test", AppURL: "http://localhost"}
	r := gin.New()
	ctrl := ctr.NewStudentController(db, cfg)
	grp := r.Group("/api")
	ctrl.RegisterRoutes(grp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u1", "exp": time.Now().Add(15 * time.Minute).Unix()})
	s, _ := token.SignedString([]byte(cfg.JWTSecret))
	return studentTestCtx{r: r, jwt: s}
}

func TestStudentCRUD(t *testing.T) {
	ctx := setupStudentTest(t)
	// create
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/students", strings.NewReader(`{"individualNumber":"S0001","fullName":"Alice","groupId":"g1"}`))
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
	req = httptest.NewRequest("GET", "/api/students", nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200 got=%d", w.Code)
	}

	// get
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/students/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get expected 200 got=%d", w.Code)
	}

	// update
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/students/"+id, strings.NewReader(`{"fullName":"Alice Smith"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200 got=%d body=%s", w.Code, w.Body.String())
	}

	// delete
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/students/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+ctx.jwt)
	ctx.r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete expected 200 got=%d", w.Code)
	}
}
