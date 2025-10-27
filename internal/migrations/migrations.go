package migrations

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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

    var count int64
    if err := db.WithContext(ctx).Model(&models.User{}).Where("role = ?", models.UserRoleAdmin).Count(&count).Error; err != nil {
        return nil, err
    }
    if count > 0 {
        return nil, nil
    }

    password, err := generatePassword()
    if err != nil {
        return nil, err
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    var rawINS string
    if err := db.WithContext(ctx).Raw(`SELECT lpad(nextval('ins_sequence')::text, 8, '0')`).Scan(&rawINS).Error; err != nil {
        return nil, err
    }
    user := models.User{
        Base: models.Base{ID: uuid.New()},
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
