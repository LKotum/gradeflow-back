package request

type CreateStudent struct {
	IndividualNumber string `json:"individualNumber" binding:"required"`
	FullName         string `json:"fullName" binding:"required"`
	GroupID          string `json:"groupId"`
}

type UpdateStudent struct {
	IndividualNumber string `json:"individualNumber"`
	FullName         string `json:"fullName"`
	GroupID          string `json:"groupId"`
}

type ListStudentQuery struct{ PaginationQuery }
