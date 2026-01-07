package dto

type LoginRequest struct {
	Identifier string `json:"identifier"` // can be username or email
	Password   string `json:"password"`
}