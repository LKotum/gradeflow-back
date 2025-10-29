package migrations

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gradeflow/internal/domain/models"
)

// BootstrapAdmin contains credentials for the generated administrator.
type BootstrapAdmin struct {
	INS      string
	Password string
}

// Run executes database migrations using gormigrate.
func Run(ctx context.Context, db *gorm.DB) (*BootstrapAdmin, error) {
	if err := db.WithContext(ctx).Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).Exec(`CREATE SEQUENCE IF NOT EXISTS ins_sequence AS BIGINT START 1`).Error; err != nil {
		return nil, err
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
				if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_grades_session_student ON grades (session_id, student_id) WHERE deleted_at IS NULL`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_teaching_assignments_teacher_subject ON teaching_assignments (teacher_id, subject_id) WHERE deleted_at IS NULL`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_subject_groups_subject_group ON subject_groups (subject_id, group_id) WHERE deleted_at IS NULL`).Error; err != nil {
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
	if err := migrator.Migrate(); err != nil {
		return nil, err
	}

	// Guard: sometimes gormigrate thinks a migration has already run (its
	// migrations table records the ID) while the actual tables are missing
	// (for example, when a dump/restore skipped tables but kept the
	// migrations table). Ensure core tables exist before we run queries that
	// depend on them (like creating the bootstrap admin).
	// Check for the users table and run AutoMigrate for the core models if
	// it's missing.
	if !db.Migrator().HasTable(&models.User{}) {
		if err := db.AutoMigrate(
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
			return nil, err
		}
	}

	admin, err := ensureSystemAdmin(ctx, db)
	if err != nil {
		return nil, err
	}
	if admin != nil {
		log.Printf("bootstrap admin created ins=%s password=%s", admin.INS, admin.Password)
	}
	return admin, nil
}

func ensureSystemAdmin(ctx context.Context, db *gorm.DB) (*BootstrapAdmin, error) {
	var existing models.User
	err := db.WithContext(ctx).Unscoped().
		Where("role = ?", models.UserRoleAdmin).
		First(&existing).Error
	if err == nil {
		// Ensure previously soft-deleted admin is restored.
		if existing.DeletedAt.Valid {
			if err := db.WithContext(ctx).Unscoped().
				Model(&models.User{}).
				Where("id = ?", existing.ID).
				Update("deleted_at", nil).Error; err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	password, err := generatePassword()
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	rawINS := "00000000"
	user := models.User{
		Base:         models.Base{ID: uuid.New()},
		Role:         models.UserRoleAdmin,
		INS:          &rawINS,
		PasswordHash: string(hash),
		FirstName:    "System",
		LastName:     "Administrator",
	}

	if err := db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &BootstrapAdmin{INS: rawINS, Password: password}, nil
}

func generatePassword() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
