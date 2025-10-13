package request

type Register struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type Login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TOTP     string `json:"totp"`
}

type Refresh struct {
	Refresh string `json:"refresh" binding:"required"`
}

type RequestReset struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPassword struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type Logout struct {
	Refresh string `json:"refresh"`
}
