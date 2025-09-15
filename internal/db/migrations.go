// internal/db/migrations.go
package db

import (
	"context"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func AutoMigrateUp(ctx context.Context, gdb *gorm.DB) error {
	if err := gdb.Exec(`create extension if not exists pgcrypto`).Error; err != nil {
		return fmt.Errorf("enable pgcrypto: %w", err)
	}

	m := gormigrate.New(gdb, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20250915_init_core",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&Department{}, &Program{}, &Subject{},
					&AcademicSession{}, &Course{},
					&Group{}, &Student{}, &Enrollment{},
					&Lesson{}, &Attendance{},
					&Assessment{}, &AssessmentGrade{},
				); err != nil {
					return err
				}

				// Индексы, ускоряющие типичные запросы
				if err := tx.Exec(`create index if not exists idx_assessment_course_date on assessments (course_id, date_at)`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`create index if not exists idx_lesson_course_time on lessons (course_id, starts_at)`).Error; err != nil {
					return err
				}

				// CHECK для шкал оценок:
				// 1) Ассессмент: валидные сочетания type↔scale
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

				// 2) Оценка: либо pass для passfail, либо число для five/hundred + границы
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

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				// снимаем чек-ограничения, чтобы не мешали удалению
				_ = tx.Exec(`alter table assessment_grades drop constraint if exists chk_grade_value_by_scale`).Error
				_ = tx.Exec(`alter table assessments drop constraint if exists chk_assessment_type_scale`).Error

				return tx.Migrator().DropTable(
					&AssessmentGrade{}, &Assessment{},
					&Attendance{}, &Lesson{},
					&Enrollment{}, &Student{}, &Group{},
					&Course{}, &AcademicSession{},
					&Subject{}, &Program{}, &Department{},
				)
			},
		},
	})

	return m.Migrate()
}
