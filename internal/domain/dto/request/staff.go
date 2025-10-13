package request

type CreateStaff struct {
	FullName     string  `json:"fullName" binding:"required"`
	DepartmentID *string `json:"departmentId"`
	Position     string  `json:"position"`
	UserID       *string `json:"userId"`
}

type UpdateStaff struct {
	FullName     *string `json:"fullName"`
	DepartmentID *string `json:"departmentId"`
	Position     *string `json:"position"`
	UserID       *string `json:"userId"`
}

type ListStaffQuery struct {
	PaginationQuery
	DepartmentID string `form:"departmentId"`
	Q            string `form:"q"`
}
