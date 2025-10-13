package response

type Teacher struct {
	ID           string  `json:"id"`
	FullName     string  `json:"fullName"`
	DepartmentID *string `json:"departmentId"`
	Title        string  `json:"title"`
	Rank         string  `json:"rank"`
	UserID       *string `json:"userId"`
}
