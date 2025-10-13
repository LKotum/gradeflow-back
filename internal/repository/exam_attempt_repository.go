package repository

import (
    "fmt"
    "time"

    req "gradeflow/internal/domain/dto/request"
    m "gradeflow/internal/domain/models"

    "gorm.io/gorm"
)

type ExamAttemptRepository interface {
    Create(attempt *m.ExamAttempt) error
    List(query req.ListExamAttemptQuery, limit, offset int, from, to *time.Time) ([]m.ExamAttempt, int64, error)
    GetByID(id string) (*m.ExamAttempt, error)
    Update(attempt *m.ExamAttempt) error
    Delete(id string) error
}

type examAttemptRepository struct {
    db *gorm.DB
}

func NewExamAttemptRepository(db *gorm.DB) ExamAttemptRepository {
    return &examAttemptRepository{db: db}
}

func (r *examAttemptRepository) Create(attempt *m.ExamAttempt) error {
    if err := r.db.Create(attempt).Error; err != nil {
        return fmt.Errorf("create exam attempt: %w", err)
    }
    return nil
}

func (r *examAttemptRepository) List(query req.ListExamAttemptQuery, limit, offset int, from, to *time.Time) ([]m.ExamAttempt, int64, error) {
    base := r.db.Model(&m.ExamAttempt{})
    if v := query.AssessmentID; v != "" {
        base = base.Where("assessment_id = ?", v)
    }
    if v := query.StudentID; v != "" {
        base = base.Where("student_id = ?", v)
    }
    if v := query.CourseID; v != "" {
        base = base.Joins("JOIN assessments ON assessments.id = exam_attempts.assessment_id").Where("assessments.course_id = ?", v)
    }
    if from != nil {
        base = base.Where("date_at >= ?", *from)
    }
    if to != nil {
        base = base.Where("date_at <= ?", *to)
    }
    var total int64
    if err := base.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count exam attempts: %w", err)
    }
    listQuery := base.Order("date_at asc nulls last")
    if limit > 0 {
        listQuery = listQuery.Limit(limit)
    }
    if offset > 0 {
        listQuery = listQuery.Offset(offset)
    }
    var attempts []m.ExamAttempt
    if err := listQuery.Find(&attempts).Error; err != nil {
        return nil, 0, fmt.Errorf("list exam attempts: %w", err)
    }
    return attempts, total, nil
}

func (r *examAttemptRepository) GetByID(id string) (*m.ExamAttempt, error) {
    var attempt m.ExamAttempt
    if err := r.db.First(&attempt, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &attempt, nil
}

func (r *examAttemptRepository) Update(attempt *m.ExamAttempt) error {
    if err := r.db.Save(attempt).Error; err != nil {
        return fmt.Errorf("update exam attempt: %w", err)
    }
    return nil
}

func (r *examAttemptRepository) Delete(id string) error {
    if err := r.db.Delete(&m.ExamAttempt{}, "id = ?", id).Error; err != nil {
        return fmt.Errorf("delete exam attempt: %w", err)
    }
    return nil
}
