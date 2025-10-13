package request

type CreateTeacher struct {
	FullName     string  `json:"fullName" binding:"required"`
	DepartmentID *string `json:"departmentId"`
	Title        string  `json:"title"`
	Rank         string  `json:"rank"`
	UserID       *string `json:"userId"`
}

type UpdateTeacher struct {
	FullName     *string `json:"fullName"`
	DepartmentID *string `json:"departmentId"`
	Title        *string `json:"title"`
	Rank         *string `json:"rank"`
	UserID       *string `json:"userId"`
}

type ListTeacherQuery struct {
	PaginationQuery
	DepartmentID string `form:"departmentId"`
	Q            string `form:"q"`
}
