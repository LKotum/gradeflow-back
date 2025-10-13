package request

type CreateCourse struct {
	SubjectID         string  `json:"subjectId" binding:"required"`
	DepartmentID      string  `json:"departmentId" binding:"required"`
	ProgramID         *string `json:"programId"`
	AcademicSessionID string  `json:"academicSessionId" binding:"required"`
	Title             string  `json:"title" binding:"required"`
	TeacherID         *string `json:"teacherId"`
	Room              string  `json:"room"`
}

type UpdateCourse struct {
	SubjectID         string  `json:"subjectId"`
	DepartmentID      string  `json:"departmentId"`
	ProgramID         *string `json:"programId"`
	AcademicSessionID string  `json:"academicSessionId"`
	Title             string  `json:"title"`
	TeacherID         *string `json:"teacherId"`
	Room              string  `json:"room"`
}

type ListCourseQuery struct {
	PaginationQuery
	DepartmentID      string `form:"departmentId"`
	ProgramID         string `form:"programId"`
	SubjectID         string `form:"subjectId"`
	AcademicSessionID string `form:"academicSessionId"`
}
