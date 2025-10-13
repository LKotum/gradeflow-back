package response

type Course struct {
	ID                string  `json:"id"`
	SubjectID         string  `json:"subjectId"`
	DepartmentID      string  `json:"departmentId"`
	ProgramID         *string `json:"programId"`
	AcademicSessionID string  `json:"academicSessionId"`
	Title             string  `json:"title"`
	TeacherID         *string `json:"teacherId"`
	Room              string  `json:"room"`
}
