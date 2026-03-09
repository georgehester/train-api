package handlers

import (
	"log"
	"net/http"
	"vulpz/train-api/src/api"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Customer struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Forename string `json:"forename"`
	Surname  string `json:"surname"`
}

type Application struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	CustomerId string `json:"customerId"`
	Approved   bool   `json:"approved"`
}

type CustomerWithApplications struct {
	Id           string        `json:"id"`
	Email        string        `json:"email"`
	Forename     string        `json:"forename"`
	Surname      string        `json:"surname"`
	Applications []Application `json:"applications"`
}

// GetAllCustomers returns the list of all customers.
// @Summary      Get all customers
// @Description  Responds with a list of all customers
// @Tags         administration
// @Produce      json
// @Success      200  {array}  Customer
// @Failure      500  {object} ErrorResponse
// @Router       /administration/customer [get]
func (environment *Environment) GetCustomers(context *gin.Context) {
	rows, databaseError := environment.Database.Query(context, "SELECT id, email, forename, surname FROM customers;")
	if databaseError != nil {
		log.Fatal(databaseError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}
	defer rows.Close()

	var customerList []Customer

	for rows.Next() {
		var customer Customer

		if scanError := rows.Scan(&customer.Id, &customer.Email, &customer.Forename, &customer.Surname); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Customers")
			return
		}

		customerList = append(customerList, customer)
	}

	context.JSON(http.StatusOK, customerList)
}

// GetCustomer returns a specific customer with their unapproved applications.
// @Summary      Get customer with pending applications
// @Description  Responds with a customer's details and their unapproved applications
// @Tags         administration
// @Produce      json
// @Param        id   path      string  true  "Customer ID"
// @Success      200  {object}  CustomerWithApplications
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /administration/customer/{id} [get]
func (environment *Environment) GetCustomer(context *gin.Context) {
	customerId := context.Param("id")

	var customer CustomerWithApplications
	customerError := environment.Database.QueryRow(context, "SELECT id, email, forename, surname FROM customers WHERE id = $1;", customerId).Scan(&customer.Id, &customer.Email, &customer.Forename, &customer.Surname)
	if customerError != nil {
		if customerError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Customer Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		}
		return
	}

	rows, applicationsError := environment.Database.Query(context, "SELECT id, name, key, customer_id, approved FROM applications WHERE customer_id = $1;", customerId)
	if applicationsError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Fetch Applications")
		return
	}
	defer rows.Close()

	customer.Applications = []Application{}

	for rows.Next() {
		var application Application

		if scanError := rows.Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Applications")
			return
		}

		customer.Applications = append(customer.Applications, application)
	}

	context.JSON(http.StatusOK, customer)
}

// GetApplications returns all applications for a specific customer.
// @Summary      Get customer applications
// @Description  Responds with all applications for a given customer
// @Tags         administration
// @Produce      json
// @Param        id   path      string  true  "Customer ID"
// @Success      200  {array}   Application
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /administration/customer/{id}/application [get]
func (environment *Environment) GetApplications(context *gin.Context) {
	customerId := context.Param("id")

	var exists bool
	existsError := environment.Database.QueryRow(context, "SELECT EXISTS(SELECT 1 FROM customers WHERE id = $1);", customerId).Scan(&exists)
	if existsError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}

	if exists == false {
		api.SendErrorResponse(context, http.StatusNotFound, "Customer Not Found")
		return
	}

	rows, applicationsError := environment.Database.Query(context, "SELECT id, name, key, customer_id, approved FROM applications WHERE customer_id = $1;", customerId)
	if applicationsError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Fetch Applications")
		return
	}
	defer rows.Close()

	var applicationList []Application

	for rows.Next() {
		var application Application

		if scanError := rows.Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Applications")
			return
		}

		applicationList = append(applicationList, application)
	}

	context.JSON(http.StatusOK, applicationList)
}

// CreateCustomer creates a new customer.
// @Summary      Create a new customer
// @Description  Creates a new customer with the provided details
// @Tags         administration
// @Accept       json
// @Produce      json
// @Param        customer  body      Customer  true  "Customer data"
// @Success      201  {object}  Customer
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /administration/customer [post]
func (environment *Environment) CreateCustomer(context *gin.Context) {
	var customer Customer

	if bindError := context.ShouldBindJSON(&customer); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	if customer.Email == "" || customer.Forename == "" || customer.Surname == "" {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	customer.Id = uuid.New().String()

	insertError := environment.Database.QueryRow(context,
		"INSERT INTO customers (id, email, forename, surname) VALUES ($1, $2, $3, $4) RETURNING id, email, forename, surname;",
		customer.Id, customer.Email, customer.Forename, customer.Surname).Scan(&customer.Id, &customer.Email, &customer.Forename, &customer.Surname)
	if insertError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Create Customer")
		return
	}

	context.JSON(http.StatusCreated, customer)
}
