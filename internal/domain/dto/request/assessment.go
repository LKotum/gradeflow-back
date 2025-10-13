package request

import "time"

type CreateAssessment struct {
	CourseID string    `json:"courseId" binding:"required"`
	Type     string    `json:"type" binding:"required"`
	DateAt   time.Time `json:"dateAt" binding:"required"`
	Room     string    `json:"room"`
	Scale    string    `json:"scale" binding:"required"`
	MaxPts   *float64  `json:"maxPts"`
}

type UpdateAssessment struct {
	CourseID string     `json:"courseId"`
	Type     string     `json:"type"`
	DateAt   *time.Time `json:"dateAt"`
	Room     string     `json:"room"`
	Scale    string     `json:"scale"`
	MaxPts   *float64   `json:"maxPts"`
}

type ListAssessmentQuery struct {
	PaginationQuery
	CourseID string `form:"courseId"`
}

type GradeItem struct {
	StudentID string   `json:"studentId" binding:"required"`
	Scale     string   `json:"scale"`
	ValueNum  *float64 `json:"valueNum"`
	ValuePass *bool    `json:"valuePass"`
}

// Standalone grade DTOs (for /grades endpoints)
type CreateAssessmentGrade struct {
	AssessmentID string   `json:"assessmentId" binding:"required"`
	StudentID    string   `json:"studentId" binding:"required"`
	Scale        string   `json:"scale"`
	ValueNum     *float64 `json:"valueNum"`
	ValuePass    *bool    `json:"valuePass"`
}

type UpdateAssessmentGrade struct {
	Scale     string   `json:"scale"`
	ValueNum  *float64 `json:"valueNum"`
	ValuePass *bool    `json:"valuePass"`
}

type ListAssessmentGradeQuery struct {
	PaginationQuery
	AssessmentID string `form:"assessmentId"`
	StudentID    string `form:"studentId"`
}
