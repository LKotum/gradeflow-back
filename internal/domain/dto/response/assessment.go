package response

import "time"

type Assessment struct {
	ID       string    `json:"id"`
	CourseID string    `json:"courseId"`
	Type     string    `json:"type"`
	DateAt   time.Time `json:"dateAt"`
	Room     string    `json:"room"`
	Scale    string    `json:"scale"`
	MaxPts   *float64  `json:"maxPts"`
}

type AssessmentGrade struct {
	ID           string   `json:"id"`
	AssessmentID string   `json:"assessmentId"`
	StudentID    string   `json:"studentId"`
	Scale        string   `json:"scale"`
	ValueNum     *float64 `json:"valueNum"`
	ValuePass    *bool    `json:"valuePass"`
	GradedBy     string   `json:"gradedBy"`
	GradedAt     string   `json:"gradedAt"`
}
