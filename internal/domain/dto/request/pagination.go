package request

// PaginationQuery carries limit/offset query parameters
type PaginationQuery struct {
	Limit  int `form:"limit"`  // max items per page
	Offset int `form:"offset"` // items to skip
}
