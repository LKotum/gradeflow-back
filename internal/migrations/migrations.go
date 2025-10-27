package migrations

import (
	"context"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"gradeflow/internal/domain/models"
)

// Run executes database migrations using gormigrate.
func Run(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return err
	}
	migrations := []*gormigrate.Migration{
		{
			ID: "20241019_init",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&models.User{},
					&models.StudentProfile{},
					&models.TeacherProfile{},
					&models.StaffProfile{},
					&models.Group{},
					&models.Subject{},
					&models.SubjectGroup{},
					&models.TeachingAssignment{},
					&models.ClassSession{},
					&models.Grade{},
					&models.RefreshToken{},
				); err != nil {
					return err
				}
				if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_grades_session_student ON grades (session_id, student_id)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_teaching_assignments_teacher_subject ON teaching_assignments (teacher_id, subject_id)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_subject_groups_subject_group ON subject_groups (subject_id, group_id)`).Error; err != nil {
					return err
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					&models.Grade{},
					&models.ClassSession{},
					&models.SubjectGroup{},
					&models.TeachingAssignment{},
					&models.Subject{},
					&models.Group{},
					&models.StaffProfile{},
					&models.TeacherProfile{},
					&models.StudentProfile{},
					&models.RefreshToken{},
					&models.User{},
				)
			},
		},
	}
	migrator := gormigrate.New(db.WithContext(ctx), gormigrate.DefaultOptions, migrations)
	migrator.InitSchema(func(tx *gorm.DB) error {
		return nil
	})
	return migrator.Migrate()
}
