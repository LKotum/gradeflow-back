package migrations

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	m "gradeflow/internal/domain/models"
	"gradeflow/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

func AutoMigrateUp(ctx context.Context, gdb *gorm.DB) error {
	if err := gdb.Exec(`create extension if not exists pgcrypto`).Error; err != nil {
		return fmt.Errorf("enable pgcrypto: %w", err)
	}

	mg := gormigrate.New(gdb, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20250915_init_core",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&m.User{}, &m.RefreshToken{},
					&m.Teacher{}, &m.StaffMember{}, &m.Admin{},
					&m.Department{}, &m.Program{}, &m.Subject{},
					&m.AcademicSession{}, &m.Course{},
					&m.Group{}, &m.Student{}, &m.Enrollment{},
					&m.Lesson{}, &m.Attendance{},
					&m.Assessment{}, &m.AssessmentGrade{},
					&m.Practice{}, &m.PracticeEnrollment{},
					&m.ExamSession{}, &m.ExamAttempt{}, &m.Credit{},
				); err != nil {
					return err
				}

				if err := tx.Exec(`create index if not exists idx_assessment_course_date on assessments (course_id, date_at)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`create index if not exists idx_lesson_course_time on lessons (course_id, starts_at)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`create index if not exists idx_practice_dates on practices (start_date, end_date)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`create index if not exists idx_exam_session_dates on exam_sessions (starts_at, ends_at)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`
                    alter table assessments
                    add constraint chk_assessment_type_scale
                    check (
                        (type='zachet'      and scale='passfail') or
                        (type='diff_zachet' and scale='five')     or
                        (type='exam'        and (scale='five' or scale='hundred'))
                    )
                `).Error; err != nil {
					return err
				}
				if err := tx.Exec(`
                    alter table assessment_grades
                    add constraint chk_grade_value_by_scale
                    check (
                        (scale='passfail' and value_pass is not null and value_num is null) or
                        (scale='five'     and value_num is not null and value_pass is null and value_num in (2,3,4,5)) or
                        (scale='hundred'  and value_num is not null and value_pass is null and value_num >= 0 and value_num <= 100)
                    )
                `).Error; err != nil {
					return err
				}
				if err := tx.Exec(`
                    alter table exam_attempts
                    add constraint chk_exam_attempt_value_by_scale
                    check (
                        (result_scale='passfail' and value_pass is not null and value_num is null) or
                        (result_scale='five'     and value_num is not null and value_pass is null and value_num in (2,3,4,5)) or
                        (result_scale='hundred'  and value_num is not null and value_pass is null and value_num >= 0 and value_num <= 100)
                    )
                `).Error; err != nil {
					return err
				}
				if err := tx.Exec(`
                    alter table practice_enrollments
                    add constraint chk_practice_enrollment_status
                    check (status in ('enrolled','completed','failed'))
                `).Error; err != nil {
					return err
				}
				if err := tx.Exec(`
                    alter table credits
                    add constraint chk_credit_type
                    check (type in ('credit','offset','transfer'))
                `).Error; err != nil {
					return err
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				_ = tx.Exec(`alter table credits drop constraint if exists chk_credit_type`).Error
				_ = tx.Exec(`alter table practice_enrollments drop constraint if exists chk_practice_enrollment_status`).Error
				_ = tx.Exec(`alter table exam_attempts drop constraint if exists chk_exam_attempt_value_by_scale`).Error
				_ = tx.Exec(`alter table assessment_grades drop constraint if exists chk_grade_value_by_scale`).Error
				_ = tx.Exec(`alter table assessments drop constraint if exists chk_assessment_type_scale`).Error
				return tx.Migrator().DropTable(
					&m.Credit{}, &m.ExamAttempt{}, &m.ExamSession{},
					&m.PracticeEnrollment{}, &m.Practice{},
					&m.AssessmentGrade{}, &m.Assessment{},
					&m.Attendance{}, &m.Lesson{},
					&m.Enrollment{}, &m.Student{}, &m.Group{},
					&m.Course{}, &m.AcademicSession{},
					&m.Subject{}, &m.Program{}, &m.Department{},
					&m.Admin{}, &m.StaffMember{}, &m.Teacher{},
					&m.RefreshToken{}, &m.User{},
				)
			},
		},
		{
			ID: "20251013_course_teacher_uuid_and_bootstrap_admin",
			Migrate: func(tx *gorm.DB) error {
				// Convert courses.teacher_id to UUID if column exists and is text; otherwise create if missing
				// Nullify non-UUID values to allow cast
				_ = tx.Exec(`UPDATE courses SET teacher_id = NULL WHERE teacher_id IS NOT NULL AND teacher_id !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'`).Error
				// Try alter type (will fail if column doesn't exist)
				if err := tx.Exec(`ALTER TABLE courses ALTER COLUMN teacher_id TYPE uuid USING teacher_id::uuid`).Error; err != nil {
					// If alter failed because column doesn't exist, add it
					_ = tx.Exec(`ALTER TABLE courses ADD COLUMN IF NOT EXISTS teacher_id uuid NULL`).Error
				}
				// Ensure FK
				_ = tx.Exec(`ALTER TABLE courses ADD CONSTRAINT IF NOT EXISTS fk_courses_teacher FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE SET NULL`).Error

				// Bootstrap default admin user (one-time)
				var cnt int64
				if err := tx.Model(&m.User{}).Where("role = ?", "admin").Count(&cnt).Error; err != nil {
					return err
				}
				if cnt == 0 {
					// Generate random password
					buf := make([]byte, 18)
					if _, err := rand.Read(buf); err != nil {
						return fmt.Errorf("rand: %w", err)
					}
					pw := base64.RawURLEncoding.EncodeToString(buf)
					hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
					if err != nil {
						return err
					}
					email := "admin@local"
					var exists m.User
					if err := tx.Where("email = ?", email).First(&exists).Error; err == nil {
						// collision: generate alternative email admin-<rand>@local
						sfx := make([]byte, 6)
						if _, err := rand.Read(sfx); err == nil {
							email = fmt.Sprintf("admin-%s@local", base64.RawURLEncoding.EncodeToString(sfx))
						}
					}
					u := &m.User{Email: email, FullName: "Default Admin", PasswordHash: string(hash), Role: "admin", Status: "active"}
					if err := tx.Create(u).Error; err != nil {
						return err
					}
					ad := &m.Admin{FullName: "Default Admin", UserID: &u.ID}
					if err := tx.Create(ad).Error; err != nil {
						return err
					}
					// Print once on first creation
					logger.Info("bootstrap admin created", "email", u.Email, "password", pw)
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
	})
	return mg.Migrate()
}
