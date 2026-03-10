package handlers

import (
	"log"
	"net/http"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetAllCustomers returns the list of all customers.
// @Summary      Get all customers
// @Description  Responds with a list of all customers
// @Tags         administration
// @Produce      json
// @Success      200  {array}  model.Customer
// @Failure      500  {object} model.ErrorResponse
// @Router       /administration/customer [get]
func (environment *Environment) GetCustomersHandler(context *gin.Context) {
	rows, databaseError := environment.Database.Query(context, "SELECT id, email, forename, surname FROM customers;")
	if databaseError != nil {
		log.Fatal(databaseError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}
	defer rows.Close()

	var customerList []model.Customer

	for rows.Next() {
		var customer model.Customer

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
// @Success      200  {object}  model.CustomerWithApplications
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /administration/customer/{id} [get]
func (environment *Environment) GetCustomerHandler(context *gin.Context) {
	customerId := context.Param("id")

	var customer model.CustomerWithApplications
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

	customer.Applications = []model.Application{}

	for rows.Next() {
		var application model.Application

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
// @Success      200  {array}   model.Application
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /administration/customer/{id}/application [get]
func (environment *Environment) GetCustomerApplicationsHandler(context *gin.Context) {
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

	var applicationList []model.Application

	for rows.Next() {
		var application model.Application

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
// @Param        customer  body      model.Customer  true  "Customer data"
// @Success      201  {object}  model.Customer
// @Failure      400  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /administration/customer [post]
func (environment *Environment) CreateCustomerHandler(context *gin.Context) {
	var customer model.Customer

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

// ApproveApplication approves a specific application and issues a fresh API key.
// @Summary      Approve application
// @Description  Approves an application so an API key can be generated
// @Tags         administration
// @Produce      json
// @Param        id   path      string             true  "Application ID"
// @Success      200  {object}  model.Application
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /administration/application/{id}/approve [post]
func (environment *Environment) ApproveApplicationHandler(context *gin.Context) {
	customerId := context.Param("customerId")
	applicationId := context.Param("applicationId")

	generatedKey, keyError := authentication.GenerateRandomApplicationKey()
	if keyError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Generate Application Key")
		return
	}

	var application model.Application
	updateError := environment.Database.QueryRow(
		context,
		"UPDATE applications SET approved = TRUE, key = $1 WHERE id = $2 AND customer_id = $3 RETURNING id, name, key, customer_id, approved",
		generatedKey,
		applicationId,
		customerId,
	).Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved)
	if updateError != nil {
		if updateError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Application Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Approve Application")
		}
		return
	}

	context.JSON(http.StatusOK, application)
}
