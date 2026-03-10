package model

type Application struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	CustomerId string `json:"customerId"`
	Approved   bool   `json:"approved"`
} // @name Application
