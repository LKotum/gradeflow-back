package request

type CreateAdmin struct {
	FullName string  `json:"fullName" binding:"required"`
	UserID   *string `json:"userId"`
}

type UpdateAdmin struct {
	FullName *string `json:"fullName"`
	UserID   *string `json:"userId"`
}

type ListAdminQuery struct {
	PaginationQuery
	Q string `form:"q"`
}
