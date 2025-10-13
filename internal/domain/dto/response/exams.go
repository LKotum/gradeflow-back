package response

type ExamSession struct {
	ID                string  `json:"id"`
	AcademicSessionID string  `json:"academicSessionId"`
	Name              string  `json:"name"`
	StartsAt          *string `json:"startsAt"`
	EndsAt            *string `json:"endsAt"`
}

type ExamAttempt struct {
	ID           string  `json:"id"`
	AssessmentID string  `json:"assessmentId"`
	StudentID    string  `json:"studentId"`
	AttemptNo    int     `json:"attemptNo"`
	DateAt       *string `json:"dateAt"`
	ResultScale  string  `json:"resultScale"`
	ValueNum     *int    `json:"valueNum"`
	ValuePass    *bool   `json:"valuePass"`
	Notes        string  `json:"notes"`
}

type Credit struct {
	ID           string  `json:"id"`
	CourseID     string  `json:"courseId"`
	StudentID    string  `json:"studentId"`
	Type         string  `json:"type"`
	Reason       string  `json:"reason"`
	ApprovedByID *string `json:"approvedById"`
}
