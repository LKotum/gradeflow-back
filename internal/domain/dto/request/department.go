package request

type CreateDepartment struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type UpdateDepartment struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ListDepartmentQuery struct {
	PaginationQuery
}
