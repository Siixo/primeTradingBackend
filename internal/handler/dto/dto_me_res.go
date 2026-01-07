package dto

type MeResponse struct{
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   uint   `json:"user_id"`
}