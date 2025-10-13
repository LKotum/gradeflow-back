package request

type CreateProgram struct {
	DepartmentID string `json:"departmentId" binding:"required"`
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

type UpdateProgram struct {
	DepartmentID string `json:"departmentId"`
	Code         string `json:"code"`
	Name         string `json:"name"`
}

type ListProgramQuery struct{ PaginationQuery }
