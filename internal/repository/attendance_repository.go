package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	m "gradeflow/internal/domain/models"
)

type AttendanceFilter struct {
	LessonID  string
	StudentID string
	CourseID  string
	SessionID string
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}

type AttendanceUpsert struct {
	StudentID string
	Status    string
}

type AttendanceRepository interface {
	Create(att *m.Attendance) error
	List(filter AttendanceFilter) ([]m.Attendance, int64, error)
	GetByID(id string) (*m.Attendance, error)
	Update(att *m.Attendance) error
	Delete(id string) error
	BulkUpsert(lessonID string, entries []AttendanceUpsert, markedBy *string, markedAt time.Time) error
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) Create(att *m.Attendance) error {
	if err := r.db.Create(att).Error; err != nil {
		return fmt.Errorf("create attendance: %w", err)
	}
	return nil
}

func (r *attendanceRepository) List(filter AttendanceFilter) ([]m.Attendance, int64, error) {
	base := r.db.Model(&m.Attendance{})
	base = applyAttendanceFilters(base, filter)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count attendances: %w", err)
	}

	var items []m.Attendance
	if err := applyAttendanceFilters(r.db.Model(&m.Attendance{}), filter).
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list attendances: %w", err)
	}
	return items, total, nil
}

func (r *attendanceRepository) GetByID(id string) (*m.Attendance, error) {
	var att m.Attendance
	if err := r.db.First(&att, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &att, nil
}

func (r *attendanceRepository) Update(att *m.Attendance) error {
	if err := r.db.Save(att).Error; err != nil {
		return fmt.Errorf("update attendance: %w", err)
	}
	return nil
}

func (r *attendanceRepository) Delete(id string) error {
	if err := r.db.Delete(&m.Attendance{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete attendance: %w", err)
	}
	return nil
}

func (r *attendanceRepository) BulkUpsert(lessonID string, entries []AttendanceUpsert, markedBy *string, markedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			var current m.Attendance
			err := tx.Where("lesson_id = ? AND student_id = ?", lessonID, entry.StudentID).First(&current).Error
			switch {
			case err == nil:
				current.Status = entry.Status
				if markedBy != nil {
					current.MarkedBy = *markedBy
				}
				current.MarkedAt = markedAt
				if err := tx.Save(&current).Error; err != nil {
					return fmt.Errorf("update attendance for student %s: %w", entry.StudentID, err)
				}
			case err == gorm.ErrRecordNotFound:
				rec := m.Attendance{
					LessonID:  lessonID,
					StudentID: entry.StudentID,
					Status:    entry.Status,
					MarkedAt:  markedAt,
				}
				if markedBy != nil {
					rec.MarkedBy = *markedBy
				}
				if err := tx.Create(&rec).Error; err != nil {
					return fmt.Errorf("create attendance for student %s: %w", entry.StudentID, err)
				}
			default:
				return fmt.Errorf("load attendance for student %s: %w", entry.StudentID, err)
			}
		}
		return nil
	})
}

func applyAttendanceFilters(db *gorm.DB, filter AttendanceFilter) *gorm.DB {
	if filter.LessonID != "" {
		db = db.Where("lesson_id = ?", filter.LessonID)
	}
	if filter.StudentID != "" {
		db = db.Where("student_id = ?", filter.StudentID)
	}
	if filter.CourseID != "" {
		db = db.Joins("JOIN lessons ON lessons.id = attendances.lesson_id").Where("lessons.course_id = ?", filter.CourseID)
	}
	if filter.SessionID != "" {
		db = db.Joins("JOIN lessons l ON l.id = attendances.lesson_id").
			Joins("JOIN courses c ON c.id = l.course_id").
			Where("c.academic_session_id = ?", filter.SessionID)
	}
	if filter.From != nil {
		db = db.Where("marked_at >= ?", *filter.From)
	}
	if filter.To != nil {
		db = db.Where("marked_at <= ?", *filter.To)
	}
	return db
}
