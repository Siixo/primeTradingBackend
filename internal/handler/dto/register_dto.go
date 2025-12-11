package dto

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Password2 string `json:"password2"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // can be username or email
	Password   string `json:"password"`
}