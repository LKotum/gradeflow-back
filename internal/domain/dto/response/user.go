package response

type TokenPair struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FullName    string `json:"fullName"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	TOTPEnabled bool   `json:"totpEnabled"`
}
