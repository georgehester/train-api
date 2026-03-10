package model

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} // @name LoginRequest

type LoginResponse struct {
	Token string `json:"token"`
} // @name LoginResponse
