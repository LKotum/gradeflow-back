package request

type CreateSubject struct {
	DepartmentID string `json:"departmentId" binding:"required"`
	Code         string `json:"code" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Credits      int    `json:"credits"`
}

type UpdateSubject struct {
	DepartmentID string `json:"departmentId"`
	Code         string `json:"code"`
	Title        string `json:"title"`
	Credits      int    `json:"credits"`
}

type ListSubjectQuery struct{ PaginationQuery }
