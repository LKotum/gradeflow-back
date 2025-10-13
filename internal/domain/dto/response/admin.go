package response

type Admin struct {
	ID       string  `json:"id"`
	FullName string  `json:"fullName"`
	UserID   *string `json:"userId"`
}
