package response

type Program struct {
	ID           string `json:"id"`
	DepartmentID string `json:"departmentId"`
	Code         string `json:"code"`
	Name         string `json:"name"`
}
