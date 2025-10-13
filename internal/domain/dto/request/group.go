package request

type CreateGroup struct {
	ProgramID string `json:"programId" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Year      int    `json:"year" binding:"required"`
}

type UpdateGroup struct {
	ProgramID string `json:"programId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Year      int    `json:"year"`
}

type ListGroupQuery struct{ PaginationQuery }
