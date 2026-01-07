package dto

type LoginResponse struct{
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   uint   `json:"user_id"`
}
