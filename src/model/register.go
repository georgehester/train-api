package model

type RegisterRequest struct {
	Forename string `json:"forename"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
} // @name RegisterRequest
