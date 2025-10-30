package migrations

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gradeflow/internal/domain/models"
	"gradeflow/pkg/utils"
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
		{
			ID: "20250207_seed_core_data",
			Migrate: func(tx *gorm.DB) error {
				// 0) общий пароль для сидов
				passwordHash, err := bcrypt.GenerateFromPassword([]byte("Test1234"), bcrypt.DefaultCost)
				if err != nil {
					return err
				}

				type userSeed struct {
					FirstName  string
					LastName   string
					MiddleName *string
					Email      string
					Role       models.UserRole
					Position   *string
					Title      *string
					Bio        *string
				}

				staffSeeds := []userSeed{
					{
						FirstName: "Анастасия", LastName: "Кузнецова", MiddleName: utils.StringPtr("Игоревна"),
						Email: "a.kuznetsova@gradeflow.local", Role: models.UserRoleDean,
						Position: utils.StringPtr("Руководитель учебного отдела"),
					},
					{
						FirstName: "Михаил", LastName: "Поляков", MiddleName: utils.StringPtr("Сергеевич"),
						Email: "m.polyakov@gradeflow.local", Role: models.UserRoleDean,
						Position: utils.StringPtr("Специалист по учебным планам"),
					},
				}

				teacherSeeds := []userSeed{
					{
						FirstName: "Екатерина", LastName: "Белова", MiddleName: utils.StringPtr("Андреевна"),
						Email: "e.belova@gradeflow.local", Role: models.UserRoleTeacher,
						Title: utils.StringPtr("Доцент"), Bio: utils.StringPtr("Исследует распределённые вычислительные системы и edge-инфраструктуру."),
					},
					{
						FirstName: "Сергей", LastName: "Орлов", MiddleName: utils.StringPtr("Петрович"),
						Email: "s.orlov@gradeflow.local", Role: models.UserRoleTeacher,
						Title: utils.StringPtr("Профессор"), Bio: utils.StringPtr("Специалист по высокопроизводительным вычислениям и параллельному программированию."),
					},
					{
						FirstName: "Дмитрий", LastName: "Иванов", MiddleName: utils.StringPtr("Алексеевич"),
						Email: "d.ivanov@gradeflow.local", Role: models.UserRoleTeacher,
						Title: utils.StringPtr("Старший преподаватель"), Bio: utils.StringPtr("Занимается автоматизацией DevOps-практик и облачными платформами."),
					},
				}

				studentSeeds := []userSeed{
					{FirstName: "Анна", LastName: "Соколова", MiddleName: utils.StringPtr("Дмитриевна"), Email: "student01@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Илья", LastName: "Морозов", MiddleName: utils.StringPtr("Андреевич"), Email: "student02@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Мария", LastName: "Васильева", MiddleName: utils.StringPtr("Сергеевна"), Email: "student03@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Никита", LastName: "Кузьмин", MiddleName: utils.StringPtr("Олегович"), Email: "student04@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Полина", LastName: "Громова", MiddleName: utils.StringPtr("Артёмовна"), Email: "student05@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Даниил", LastName: "Леонов", MiddleName: utils.StringPtr("Ильич"), Email: "student06@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Валерия", LastName: "Мельникова", MiddleName: utils.StringPtr("Романовна"), Email: "student07@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Егор", LastName: "Андреев", MiddleName: utils.StringPtr("Евгеньевич"), Email: "student08@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Софья", LastName: "Николаева", MiddleName: utils.StringPtr("Михайловна"), Email: "student09@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Максим", LastName: "Жуков", MiddleName: utils.StringPtr("Витальевич"), Email: "student10@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Екатерина", LastName: "Фролова", MiddleName: utils.StringPtr("Григорьевна"), Email: "student11@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Артём", LastName: "Гордеев", MiddleName: utils.StringPtr("Кириллович"), Email: "student12@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Алина", LastName: "Романова", MiddleName: utils.StringPtr("Павловна"), Email: "student13@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Кирилл", LastName: "Корнеев", MiddleName: utils.StringPtr("Степанович"), Email: "student14@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Вероника", LastName: "Лазарева", MiddleName: utils.StringPtr("Игоревна"), Email: "student15@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Иван", LastName: "Чернов", MiddleName: utils.StringPtr("Алексеевич"), Email: "student16@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Дарья", LastName: "Калинина", MiddleName: utils.StringPtr("Фёдоровна"), Email: "student17@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Глеб", LastName: "Семенов", MiddleName: utils.StringPtr("Никитич"), Email: "student18@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Татьяна", LastName: "Блинова", MiddleName: utils.StringPtr("Артёмовна"), Email: "student19@gradeflow.local", Role: models.UserRoleStudent},
					{FirstName: "Алексей", LastName: "Новиков", MiddleName: utils.StringPtr("Максимович"), Email: "student20@gradeflow.local", Role: models.UserRoleStudent},
				}

				quietTx := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})

				// ---------- helpers ----------
				upsertUser := func(seed userSeed) (*models.User, error) {
					var user models.User
					// ищем с учётом soft-delete
					err := quietTx.Unscoped().Where("email = ?", seed.Email).First(&user).Error
					if errors.Is(err, gorm.ErrRecordNotFound) {
						// создаём нового
						ins, err := nextINS(ctx, quietTx)
						if err != nil {
							return nil, err
						}
						user = models.User{
							Base:         models.Base{ID: uuid.New()},
							Role:         seed.Role,
							INS:          &ins,
							Email:        utils.StringPtr(seed.Email),
							FirstName:    seed.FirstName,
							LastName:     seed.LastName,
							MiddleName:   seed.MiddleName,
							PasswordHash: string(passwordHash),
						}
						if err := quietTx.Create(&user).Error; err != nil {
							return nil, err
						}
					} else if err != nil {
						return nil, err
					} else {
						// запись есть — поднимем, если soft-deleted
						if user.DeletedAt.Valid {
							if err := quietTx.Unscoped().
								Model(&models.User{}).
								Where("id = ?", user.ID).
								Update("deleted_at", nil).Error; err != nil {
								return nil, err
							}
						}
						// выровняем поля и пароль
						updates := map[string]any{
							"role":          seed.Role,
							"first_name":    seed.FirstName,
							"last_name":     seed.LastName,
							"middle_name":   seed.MiddleName,
							"password_hash": string(passwordHash),
						}
						if err := quietTx.Model(&user).Updates(updates).Error; err != nil {
							return nil, err
						}
					}
					return &user, nil
				}

				ensureStaffProfile := func(userID uuid.UUID, position *string) error {
					var sp models.StaffProfile
					if err := quietTx.Where("user_id = ?", userID).First(&sp).Error; errors.Is(err, gorm.ErrRecordNotFound) {
						sp = models.StaffProfile{Base: models.Base{ID: uuid.New()}, UserID: userID, Position: position}
						return quietTx.Create(&sp).Error
					} else if err != nil {
						return err
					}
					return nil
				}

				ensureTeacherProfile := func(userID uuid.UUID, title, bio *string) error {
					var tp models.TeacherProfile
					if err := quietTx.Where("user_id = ?", userID).First(&tp).Error; errors.Is(err, gorm.ErrRecordNotFound) {
						tp = models.TeacherProfile{Base: models.Base{ID: uuid.New()}, UserID: userID, Title: title, Bio: bio}
						return quietTx.Create(&tp).Error
					} else if err != nil {
						return err
					}
					return nil
				}

				ensureStudentProfile := func(u *models.User) error {
					var sp models.StudentProfile
					if err := quietTx.Where("user_id = ?", u.ID).First(&sp).Error; errors.Is(err, gorm.ErrRecordNotFound) {
						index := fmt.Sprintf("ST-%s", *u.INS)
						sp = models.StudentProfile{Base: models.Base{ID: uuid.New()}, UserID: u.ID, Index: index}
						return quietTx.Create(&sp).Error
					} else if err != nil {
						return err
					}
					return nil
				}

				ensureGroup := func(g models.Group) error {
					var existing models.Group
					if err := quietTx.Unscoped().Where("name = ?", g.Name).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
						return quietTx.Create(&g).Error
					} else if err != nil {
						return err
					}
					if existing.DeletedAt.Valid {
						return quietTx.Unscoped().Model(&models.Group{}).Where("id = ?", existing.ID).Update("deleted_at", nil).Error
					}
					// можно обновить описание, если поменялось
					return quietTx.Model(&existing).Update("description", g.Description).Error
				}

				ensureSubject := func(s models.Subject) error {
					var existing models.Subject
					if err := quietTx.Unscoped().Where("code = ?", s.Code).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
						return quietTx.Create(&s).Error
					} else if err != nil {
						return err
					}
					if existing.DeletedAt.Valid {
						if err := quietTx.Unscoped().Model(&models.Subject{}).Where("id = ?", existing.ID).Update("deleted_at", nil).Error; err != nil {
							return err
						}
					}
					// выровняем имя/описание
					return quietTx.Model(&existing).Updates(map[string]any{"name": s.Name, "description": s.Description}).Error
				}
				// ---------- /helpers ----------

				// 1) сотрудники (деканат)
				for _, seed := range staffSeeds {
					u, err := upsertUser(seed)
					if err != nil {
						return err
					}
					if err := ensureStaffProfile(u.ID, seed.Position); err != nil {
						return err
					}
				}

				// 2) преподаватели
				for _, seed := range teacherSeeds {
					u, err := upsertUser(seed)
					if err != nil {
						return err
					}
					if err := ensureTeacherProfile(u.ID, seed.Title, seed.Bio); err != nil {
						return err
					}
				}

				// 3) студенты
				for _, seed := range studentSeeds {
					u, err := upsertUser(seed)
					if err != nil {
						return err
					}
					if err := ensureStudentProfile(u); err != nil {
						return err
					}
				}

				// 4) группы
				groupSeeds := []models.Group{
					{Base: models.Base{ID: uuid.New()}, Name: "ИКТ-101", Description: utils.StringPtr("Бакалавриат «Киберфизическая инженерия»")},
					{Base: models.Base{ID: uuid.New()}, Name: "ИКТ-102", Description: utils.StringPtr("Бакалавриат «Интеллектуальные вычислительные системы»")},
				}
				for _, g := range groupSeeds {
					if err := ensureGroup(g); err != nil {
						return err
					}
				}

				// 5) предметы
				subjectSeeds := []models.Subject{
					{Base: models.Base{ID: uuid.New()}, Code: "ICT101", Name: "Архитектура вычислительных систем", Description: utils.StringPtr("Проектирование и анализ аппаратных платформ.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT102", Name: "Алгоритмы и структуры данных", Description: utils.StringPtr("Эффективные методы обработки информации.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT103", Name: "Операционные системы и виртуализация", Description: utils.StringPtr("Механизмы управления ресурсами и виртуальными средами.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT104", Name: "Компьютерные сети и кибербезопасность", Description: utils.StringPtr("Безопасная и отказоустойчивая сетевое взаимодействие.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT105", Name: "Инженерия программного обеспечения", Description: utils.StringPtr("Методы разработки, тестирования и сопровождения ПО.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT106", Name: "Базы данных и информационные хранилища", Description: utils.StringPtr("Проектирование и оптимизация систем хранения данных.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT107", Name: "Машинное обучение и анализ данных", Description: utils.StringPtr("Модели искусственного интеллекта и аналитика больших данных.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT108", Name: "Встроенные и киберфизические системы", Description: utils.StringPtr("Интеграция вычислений в физические объекты.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT109", Name: "Облачные вычисления и DevOps-практики", Description: utils.StringPtr("Оркестрация сервисов и автоматизация поставки изменений.")},
					{Base: models.Base{ID: uuid.New()}, Code: "ICT110", Name: "Человеко-машинные интерфейсы", Description: utils.StringPtr("Проектирование UX и интерактивных интерфейсов.")},
				}
				for _, s := range subjectSeeds {
					if err := ensureSubject(s); err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				subjectCodes := []string{"ICT101", "ICT102", "ICT103", "ICT104", "ICT105", "ICT106", "ICT107", "ICT108", "ICT109", "ICT110"}
				if err := tx.Where("code IN ?", subjectCodes).Delete(&models.Subject{}).Error; err != nil {
					return err
				}
				groupNames := []string{"ИКТ-101", "ИКТ-102"}
				if err := tx.Where("name IN ?", groupNames).Delete(&models.Group{}).Error; err != nil {
					return err
				}
				emails := []string{
					"a.kuznetsova@gradeflow.local", "m.polyakov@gradeflow.local",
					"e.belova@gradeflow.local", "s.orlov@gradeflow.local", "d.ivanov@gradeflow.local",
					"student01@gradeflow.local", "student02@gradeflow.local", "student03@gradeflow.local",
					"student04@gradeflow.local", "student05@gradeflow.local", "student06@gradeflow.local",
					"student07@gradeflow.local", "student08@gradeflow.local", "student09@gradeflow.local",
					"student10@gradeflow.local", "student11@gradeflow.local", "student12@gradeflow.local",
					"student13@gradeflow.local", "student14@gradeflow.local", "student15@gradeflow.local",
					"student16@gradeflow.local", "student17@gradeflow.local", "student18@gradeflow.local",
					"student19@gradeflow.local", "student20@gradeflow.local",
				}
				if err := tx.Where("email IN ?", emails).Delete(&models.User{}).Error; err != nil {
					return err
				}
				return nil
			},
		},
	}

	migrator := gormigrate.New(db.WithContext(ctx), gormigrate.DefaultOptions, migrations)
	// migrator.InitSchema(func(tx *gorm.DB) error {
	// 	return nil
	// })
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

func nextINS(ctx context.Context, db *gorm.DB) (string, error) {
	var ins string
	if err := db.WithContext(ctx).
		Raw(`SELECT lpad(nextval('ins_sequence')::text, 8, '0')`).
		Scan(&ins).Error; err != nil {
		return "", err
	}
	return ins, nil
}
