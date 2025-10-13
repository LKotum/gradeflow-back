package response

import "time"

type Practice struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	DepartmentID *string    `json:"departmentId"`
	ProgramID    *string    `json:"programId"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	SupervisorID *string    `json:"supervisorId"`
}

type PracticeEnrollment struct {
	ID         string `json:"id"`
	PracticeID string `json:"practiceId"`
	StudentID  string `json:"studentId"`
	Place      string `json:"place"`
	Status     string `json:"status"`
}
