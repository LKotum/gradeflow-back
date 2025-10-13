package response

type Subject struct {
	ID           string `json:"id"`
	DepartmentID string `json:"departmentId"`
	Code         string `json:"code"`
	Title        string `json:"title"`
	Credits      int    `json:"credits"`
}
