package request

type CreateExamSession struct {
	AcademicSessionID string  `json:"academicSessionId" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	StartsAt          *string `json:"startsAt"`
	EndsAt            *string `json:"endsAt"`
}

type UpdateExamSession struct {
	Name     *string `json:"name"`
	StartsAt *string `json:"startsAt"`
	EndsAt   *string `json:"endsAt"`
}

type ListExamSessionQuery struct {
	PaginationQuery
	AcademicSessionID string `form:"academicSessionId"`
	Q                 string `form:"q"`
}

type CreateExamAttempt struct {
	AssessmentID string  `json:"assessmentId" binding:"required"`
	StudentID    string  `json:"studentId" binding:"required"`
	AttemptNo    int     `json:"attemptNo"`
	DateAt       *string `json:"dateAt"`
	ResultScale  string  `json:"resultScale" binding:"required"`
	ValueNum     *int    `json:"valueNum"`
	ValuePass    *bool   `json:"valuePass"`
	Notes        string  `json:"notes"`
}

type UpdateExamAttempt struct {
	AttemptNo   *int    `json:"attemptNo"`
	DateAt      *string `json:"dateAt"`
	ResultScale *string `json:"resultScale"`
	ValueNum    *int    `json:"valueNum"`
	ValuePass   *bool   `json:"valuePass"`
	Notes       *string `json:"notes"`
}

type ListExamAttemptQuery struct {
	PaginationQuery
	AssessmentID string `form:"assessmentId"`
	StudentID    string `form:"studentId"`
}

type CreateCredit struct {
	CourseID     string  `json:"courseId" binding:"required"`
	StudentID    string  `json:"studentId" binding:"required"`
	Type         string  `json:"type" binding:"required"`
	Reason       string  `json:"reason"`
	ApprovedByID *string `json:"approvedById"`
}

type UpdateCredit struct {
	Type         *string `json:"type"`
	Reason       *string `json:"reason"`
	ApprovedByID *string `json:"approvedById"`
}

type ListCreditQuery struct {
	PaginationQuery
	CourseID  string `form:"courseId"`
	StudentID string `form:"studentId"`
	Type      string `form:"type"`
}
