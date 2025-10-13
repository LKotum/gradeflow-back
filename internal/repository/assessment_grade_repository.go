package repository

import (
	"fmt"
	"time"

	req "gradeflow/internal/domain/dto/request"
	m "gradeflow/internal/domain/models"

	"gorm.io/gorm"
)

type AssessmentGradeRepository interface {
	Create(g *m.AssessmentGrade) error
	List(query req.ListAssessmentGradeQuery, limit, offset int, from, to *time.Time) ([]m.AssessmentGrade, int64, error)
	ListByAssessment(assessmentID string, limit, offset int) ([]m.AssessmentGrade, int64, error)
	GetByID(id string) (*m.AssessmentGrade, error)
	GetByAssessmentAndStudent(assessmentID, studentID string) (*m.AssessmentGrade, error)
	Update(g *m.AssessmentGrade) error
	Delete(id string) error
}

type assessmentGradeRepository struct {
	db *gorm.DB
}

func NewAssessmentGradeRepository(db *gorm.DB) AssessmentGradeRepository {
	return &assessmentGradeRepository{db: db}
}

func (r *assessmentGradeRepository) Create(g *m.AssessmentGrade) error {
	if err := r.db.Create(g).Error; err != nil {
		return fmt.Errorf("create assessment grade: %w", err)
	}
	return nil
}

func (r *assessmentGradeRepository) List(query req.ListAssessmentGradeQuery, limit, offset int, from, to *time.Time) ([]m.AssessmentGrade, int64, error) {
	base := applyAssessmentGradeFilters(r.db.Model(&m.AssessmentGrade{}), query, from, to)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count assessment grades: %w", err)
	}
	listQuery := applyAssessmentGradeFilters(r.db.Model(&m.AssessmentGrade{}), query, from, to)
	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}
	var grades []m.AssessmentGrade
	if err := listQuery.Find(&grades).Error; err != nil {
		return nil, 0, fmt.Errorf("list assessment grades: %w", err)
	}
	return grades, total, nil
}

func (r *assessmentGradeRepository) ListByAssessment(assessmentID string, limit, offset int) ([]m.AssessmentGrade, int64, error) {
	q := r.db.Model(&m.AssessmentGrade{}).Where("assessment_id = ?", assessmentID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count grades for assessment: %w", err)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	var grades []m.AssessmentGrade
	if err := q.Find(&grades).Error; err != nil {
		return nil, 0, fmt.Errorf("list grades for assessment: %w", err)
	}
	return grades, total, nil
}

func (r *assessmentGradeRepository) GetByID(id string) (*m.AssessmentGrade, error) {
	var g m.AssessmentGrade
	if err := r.db.First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *assessmentGradeRepository) GetByAssessmentAndStudent(assessmentID, studentID string) (*m.AssessmentGrade, error) {
	var g m.AssessmentGrade
	if err := r.db.Where("assessment_id = ? AND student_id = ?", assessmentID, studentID).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *assessmentGradeRepository) Update(g *m.AssessmentGrade) error {
	if err := r.db.Save(g).Error; err != nil {
		return fmt.Errorf("update assessment grade: %w", err)
	}
	return nil
}

func (r *assessmentGradeRepository) Delete(id string) error {
	if err := r.db.Delete(&m.AssessmentGrade{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete assessment grade: %w", err)
	}
	return nil
}

func applyAssessmentGradeFilters(db *gorm.DB, query req.ListAssessmentGradeQuery, from, to *time.Time) *gorm.DB {
	if v := query.AssessmentID; v != "" {
		db = db.Where("assessment_id = ?", v)
	}
	if v := query.StudentID; v != "" {
		db = db.Where("student_id = ?", v)
	}
	if v := query.CourseID; v != "" {
		db = db.Joins("JOIN assessments ON assessments.id = assessment_grades.assessment_id").Where("assessments.course_id = ?", v)
	}
	if from != nil {
		db = db.Where("graded_at >= ?", *from)
	}
	if to != nil {
		db = db.Where("graded_at <= ?", *to)
	}
	return db
}
