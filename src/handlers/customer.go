package handlers

import (
	"net/http"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateApplicationRequest struct {
	Name string `json:"name" binding:"required"`
}

// @Summary      Get customer data
// @Description  Returns all customer data including applications
// @Tags         Customer
// @Produce      json
// @Param        id   path      string  true  "Customer ID"
// @Success      200  {object}  model.CustomerWithApplications
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /customer/{id} [get]
func (environment *Environment) GetCustomerByIdHandler(context *gin.Context) {
	customerId := context.Param("id")

	claims, ok := context.MustGet(authentication.ClaimsKey).(*authentication.Claims)
	if ok == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
		return
	}

	// Make sure the current user is the one requesting the data
	if claims.Id != customerId {
		api.SendErrorResponse(context, http.StatusForbidden, "Permission Denied")
		return
	}

	var customer model.CustomerWithApplications
	customerError := environment.Database.QueryRow(
		context,
		"SELECT id, email, forename, surname FROM customers WHERE id = $1",
		customerId,
	).Scan(&customer.Id, &customer.Email, &customer.Forename, &customer.Surname)
	if customerError != nil {
		if customerError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Customer Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Database Error")
		}
		return
	}

	// Fetch all applications for the customer
	rows, applicationsError := environment.Database.Query(
		context,
		"SELECT id, name, key, customer_id, approved FROM applications WHERE customer_id = $1",
		customerId,
	)
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

		// Blank out key if not approved
		if !application.Approved {
			application.Key = ""
		}

		customer.Applications = append(customer.Applications, application)
	}

	context.JSON(http.StatusOK, customer)
}

// // @Summary      Get customer applications
// // @Description  Returns all applications for a customer, blanking out key if not approved
// // @Tags         Customer
// // @Produce      json
// // @Param        id   path      string  true  "Customer ID"
// // @Success      200  {array}   model.Application
// // @Failure      404  {object}  model.ErrorResponse
// // @Failure      500  {object}  model.ErrorResponse
// // @Router       /customer/{id}/application [get]
// func (environment *Environment) GetCustomerApplicationsHandler(context *gin.Context) {
// 	customerId := context.Param("id")

// 	// Verify customer exists
// 	var exists bool
// 	existsError := environment.Database.QueryRow(
// 		context,
// 		"SELECT EXISTS(SELECT 1 FROM customers WHERE id = $1)",
// 		customerId,
// 	).Scan(&exists)
// 	if existsError != nil {
// 		api.SendErrorResponse(context, http.StatusInternalServerError, "Database Error")
// 		return
// 	}

// 	if !exists {
// 		api.SendErrorResponse(context, http.StatusNotFound, "Customer Not Found")
// 		return
// 	}

// 	rows, applicationsError := environment.Database.Query(
// 		context,
// 		"SELECT id, name, key, customer_id, approved FROM applications WHERE customer_id = $1",
// 		customerId,
// 	)
// 	if applicationsError != nil {
// 		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Fetch Applications")
// 		return
// 	}
// 	defer rows.Close()

// 	var applicationList []model.Application

// 	for rows.Next() {
// 		var application model.Application

// 		if scanError := rows.Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved); scanError != nil {
// 			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Applications")
// 			return
// 		}

// 		// Blank out key if not approved
// 		if !application.Approved {
// 			application.Key = ""
// 		}

// 		applicationList = append(applicationList, application)
// 	}

// 	context.JSON(http.StatusOK, applicationList)
// }

// // @Summary      Create application
// // @Description  Creates a new application for the customer
// // @Tags         Customer
// // @Accept       json
// // @Param        id    path      string                       true  "Customer ID"
// // @Param        body  body      CreateApplicationRequest     true  "Application data"
// // @Produce      json
// // @Success      201   {object}  model.Application
// // @Failure      400   {object}  model.ErrorResponse "Malformed Request Body"
// // @Failure      404   {object}  model.ErrorResponse "Customer Not Found"
// // @Failure      500   {object}  model.ErrorResponse "Internal Server Error"
// // @Router       /customer/{id}/application [post]
// func (environment *Environment) CreateApplicationHandler(context *gin.Context) {
// 	customerId := context.Param("id")
// 	var request CreateApplicationRequest

// 	if bindError := context.ShouldBindJSON(&request); bindError != nil {
// 		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
// 		return
// 	}

// 	if request.Name == "" {
// 		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
// 		return
// 	}

// 	// Verify customer exists
// 	var exists bool
// 	existsError := environment.Database.QueryRow(
// 		context,
// 		"SELECT EXISTS(SELECT 1 FROM customers WHERE id = $1)",
// 		customerId,
// 	).Scan(&exists)
// 	if existsError != nil {
// 		api.SendErrorResponse(context, http.StatusInternalServerError, "Database Error")
// 		return
// 	}

// 	if !exists {
// 		api.SendErrorResponse(context, http.StatusNotFound, "Customer Not Found")
// 		return
// 	}

// 	// Create new application
// 	application := model.Application{
// 		Id:         uuid.New().String(),
// 		Name:       request.Name,
// 		CustomerId: customerId,
// 		Approved:   false,
// 		Key:        "", // Will be generated upon approval
// 	}

// 	insertError := environment.Database.QueryRow(
// 		context,
// 		"INSERT INTO applications (id, name, key, customer_id, approved) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, key, customer_id, approved",
// 		application.Id,
// 		application.Name,
// 		application.Key,
// 		application.CustomerId,
// 		application.Approved,
// 	).Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved)
// 	if insertError != nil {
// 		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Create Application")
// 		return
// 	}

// 	context.JSON(http.StatusCreated, application)
// }

// // @Summary      Delete application
// // @Description  Deletes an application for the customer
// // @Tags         Customer
// // @Param        id              path      string  true  "Customer ID"
// // @Param        application_id  query     string  true  "Application ID"
// // @Produce      json
// // @Success      204
// // @Failure      400   {object}  model.ErrorResponse "Missing application_id"
// // @Failure      404   {object}  model.ErrorResponse "Application Not Found"
// // @Failure      500   {object}  model.ErrorResponse "Internal Server Error"
// // @Router       /customer/{id}/application [delete]
// func (environment *Environment) DeleteApplicationHandler(context *gin.Context) {
// 	customerId := context.Param("id")
// 	applicationId := context.Query("application_id")

// 	if applicationId == "" {
// 		api.SendErrorResponse(context, http.StatusBadRequest, "Missing application_id Parameter")
// 		return
// 	}

// 	// Verify application exists and belongs to the customer
// 	var exists bool
// 	existsError := environment.Database.QueryRow(
// 		context,
// 		"SELECT EXISTS(SELECT 1 FROM applications WHERE id = $1 AND customer_id = $2)",
// 		applicationId,
// 		customerId,
// 	).Scan(&exists)
// 	if existsError != nil {
// 		api.SendErrorResponse(context, http.StatusInternalServerError, "Database Error")
// 		return
// 	}

// 	if !exists {
// 		api.SendErrorResponse(context, http.StatusNotFound, "Application Not Found")
// 		return
// 	}

// 	// Delete the application
// 	deleteError := environment.Database.QueryRow(
// 		context,
// 		"DELETE FROM applications WHERE id = $1",
// 		applicationId,
// 	).Scan()

// 	if deleteError != nil && deleteError != pgx.ErrNoRows {
// 		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Delete Application")
// 		return
// 	}

// 	context.Status(http.StatusNoContent)
// }
