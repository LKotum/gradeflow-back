package request

import "time"

type CreatePractice struct {
	Title        string     `json:"title" binding:"required"`
	Description  string     `json:"description"`
	DepartmentID *string    `json:"departmentId"`
	ProgramID    *string    `json:"programId"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	SupervisorID *string    `json:"supervisorId"`
}

type UpdatePractice struct {
	Title        *string    `json:"title"`
	Description  *string    `json:"description"`
	DepartmentID *string    `json:"departmentId"`
	ProgramID    *string    `json:"programId"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	SupervisorID *string    `json:"supervisorId"`
}

type ListPracticeQuery struct {
	PaginationQuery
	DepartmentID string `form:"departmentId"`
	ProgramID    string `form:"programId"`
	From         string `form:"from"`
	To           string `form:"to"`
	Q            string `form:"q"`
}

type PracticeEnrollmentItem struct {
	StudentID string `json:"studentId" binding:"required"`
	Place     string `json:"place"`
	Status    string `json:"status"`
}
