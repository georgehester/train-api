package handlers

import (
	"net/http"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// @Summary      Get Customer
// @Description  Returns all customer data including applications
// @Tags         Customer
// @Produce      json
// @Param        customerId   path      string  true  "Customer Id"
// @Success      200  {object}  model.CustomerWithApplications
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /customer/{id} [get]
func (environment *Environment) GetCustomerByIdHandler(context *gin.Context) {
	customerId := context.Param("customerId")

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

// @Summary      Create Application
// @Description  Creates a new application for the authenticated customer
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Param        customerId    path      string                    true  "Customer Id"
// @Param        body  body      CreateApplicationRequest  true  "Create Application Request"
// @Success      201   {object}  model.Application
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /customer/{id}/application [post]
func (environment *Environment) CreateApplicationHandler(context *gin.Context) {
	customerId := context.Param("customerId")

	claims, ok := context.MustGet(authentication.ClaimsKey).(*authentication.Claims)
	if ok == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
		return
	}

	if claims.Id != customerId {
		api.SendErrorResponse(context, http.StatusForbidden, "Permission Denied")
		return
	}

	var request model.CreateApplicationRequest
	if bindError := context.ShouldBindJSON(&request); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	var application model.Application
	application.Id = uuid.New().String()
	application.Name = request.Name

	generatedKey, keyError := authentication.GenerateRandomApplicationKey()
	if keyError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Generate Application Key")
		return
	}

	application.Key = generatedKey
	application.CustomerId = customerId

	createError := environment.Database.QueryRow(
		context,
		"INSERT INTO applications (id, name, key, customer_id) VALUES ($1, $2, $3, $4) RETURNING id, name, key, customer_id, approved",
		application.Id,
		application.Name,
		application.Key,
		application.CustomerId,
	).Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved)
	if createError != nil {
		if createError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Customer Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Create Application")
		}
		return
	}

	if !application.Approved {
		application.Key = ""
	}

	context.JSON(http.StatusCreated, application)
}

// @Summary      List Customer Applications
// @Description  Fetches all application details for the authenticated customer
// @Tags         Customer
// @Produce      json
// @Param        customerId   path      string               true  "Customer Id"
// @Success      200  {array}   model.Application
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /customer/{id}/application [get]
func (environment *Environment) GetApplicationsHandler(context *gin.Context) {
	customerId := context.Param("customerId")

	claims, ok := context.MustGet(authentication.ClaimsKey).(*authentication.Claims)
	if ok == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
		return
	}

	if claims.Id != customerId {
		api.SendErrorResponse(context, http.StatusForbidden, "Permission Denied")
		return
	}

	rows, queryError := environment.Database.Query(
		context,
		"SELECT id, name, key, customer_id, approved FROM applications WHERE customer_id = $1",
		customerId,
	)
	if queryError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Fetch Applications")
		return
	}
	defer rows.Close()

	applicationList := []model.Application{}

	for rows.Next() {
		var application model.Application

		if scanError := rows.Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Applications")
			return
		}

		if !application.Approved {
			application.Key = ""
		}

		applicationList = append(applicationList, application)
	}

	context.JSON(http.StatusOK, applicationList)
}

// @Summary      Get Application
// @Description  Fetches application details for the authenticated customer
// @Tags         Customer
// @Produce      json
// @Param        customerId     path      string             true  "Customer Id"
// @Param        applicationId  path      string             true  "Application Id"
// @Success      200    {object}  model.Application
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401    {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404    {object}  model.ErrorResponse
// @Failure      500    {object}  model.ErrorResponse
// @Router       /customer/{customerId}/application/{applicationId} [get]
func (environment *Environment) GetApplicationHandler(context *gin.Context) {
	customerId := context.Param("customerId")
	applicationId := context.Param("applicationId")

	claims, ok := context.MustGet(authentication.ClaimsKey).(*authentication.Claims)
	if ok == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
		return
	}

	if claims.Id != customerId {
		api.SendErrorResponse(context, http.StatusForbidden, "Permission Denied")
		return
	}

	var application model.Application
	queryError := environment.Database.QueryRow(
		context,
		"SELECT id, name, key, customer_id, approved FROM applications WHERE id = $1 AND customer_id = $2",
		applicationId,
		customerId,
	).Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved)
	if queryError != nil {
		if queryError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Application Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Database Error")
		}
		return
	}

	if !application.Approved {
		application.Key = ""
	}

	context.JSON(http.StatusOK, application)
}

// @Summary      Delete Application
// @Description  Deletes an application for the authenticated customer
// @Tags         Customer
// @Param        customerId     path      string             true  "Customer Id"
// @Param        applicationId  path      string             true  "Application Id"
// @Success      204
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /customer/{customerId}/application/{applicationId} [delete]
func (environment *Environment) DeleteApplicationHandler(context *gin.Context) {
	customerId := context.Param("customerId")
	applicationId := context.Param("applicationId")

	claims, ok := context.MustGet(authentication.ClaimsKey).(*authentication.Claims)
	if ok == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
		return
	}

	if claims.Id != customerId {
		api.SendErrorResponse(context, http.StatusForbidden, "Permission Denied")
		return
	}

	deleteCommand, deleteError := environment.Database.Exec(
		context,
		"DELETE FROM applications WHERE id = $1 AND customer_id = $2",
		applicationId,
		customerId,
	)
	if deleteError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Delete Application")
		return
	}

	if deleteCommand.RowsAffected() == 0 {
		api.SendErrorResponse(context, http.StatusNotFound, "Application Not Found")
		return
	}

	context.Status(http.StatusNoContent)
}

// @Summary      Refresh Application Key
// @Description  Refreshes the API key associated with an application
// @Tags         Customer
// @Produce      json
// @Param        customerId     path      string             true  "Customer Id"
// @Param        applicationId  path      string             true  "Application Id"
// @Success      200   {object}  model.Application
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      404   {object}  model.ErrorResponse
// @Failure      500   {object}  model.ErrorResponse
// @Router       /customer/{customerId}/application/{applicationId}/key/refresh [post]
func (environment *Environment) RefreshApplicationKeyHandler(context *gin.Context) {
	customerId := context.Param("customerId")
	applicationId := context.Param("applicationId")

	claims, ok := context.MustGet(authentication.ClaimsKey).(*authentication.Claims)
	if ok == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
		return
	}

	if claims.Id != customerId {
		api.SendErrorResponse(context, http.StatusForbidden, "Permission Denied")
		return
	}

	var application model.Application
	generatedKey, keyError := authentication.GenerateRandomApplicationKey()
	if keyError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Generate Application Key")
		return
	}

	updateError := environment.Database.QueryRow(
		context,
		"UPDATE applications SET key = $1 WHERE id = $2 AND customer_id = $3 RETURNING id, name, key, customer_id, approved",
		generatedKey,
		applicationId,
		customerId,
	).Scan(&application.Id, &application.Name, &application.Key, &application.CustomerId, &application.Approved)
	if updateError != nil {
		if updateError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Application Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Refresh Application Key")
		}
		return
	}

	if !application.Approved {
		application.Key = ""
	}

	context.JSON(http.StatusOK, application)
}
