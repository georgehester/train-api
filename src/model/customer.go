package model

type Customer struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Forename string `json:"forename"`
	Surname  string `json:"surname"`
} // @name Customer

type CustomerWithApplications struct {
	Id           string        `json:"id"`
	Email        string        `json:"email"`
	Forename     string        `json:"forename"`
	Surname      string        `json:"surname"`
	Applications []Application `json:"applications"`
} // @name CustomerWithApplications

type CreateCustomerRequest struct {
	Email    string `json:"email"`
	Forename string `json:"forename"`
	Surname  string `json:"surname"`
} // @name Customer
