package response

type Staff struct {
	ID           string  `json:"id"`
	FullName     string  `json:"fullName"`
	DepartmentID *string `json:"departmentId"`
	Position     string  `json:"position"`
	UserID       *string `json:"userId"`
}
