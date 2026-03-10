package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/authentication/cryptography"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary      Login
// @Description  Returns a JWT (Json Web Token) which will expire in 24 hours.
// @Tags         Administration
// @Accept       json
// @Param        body  body      model.LoginRequest  true "Login Request"
// @Produce      json
// @Success      200  {object} model.LoginResponse
// @Failure      400  {object} model.ErrorResponse "Malformed Request Payload"
// @Failure      401  {object} model.ErrorResponse "Invalid Credentials"
// @Failure      500  {object} model.ErrorResponse "Internal Server Error"
// @Router       /administration/login [post]
func (environment *Environment) AdministrationLoginHandler(context *gin.Context) {
	var request model.LoginRequest

	if bindError := context.ShouldBindJSON(&request); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	request.Email = strings.ToUpper(request.Email)

	var administrator model.Administrator
	administrator.Email = request.Email
	var hash string

	fmt.Println(request)

	queryError := environment.Database.QueryRow(
		context,
		"SELECT id, forename, surname, hash FROM administrators WHERE email = $1",
		request.Email,
	).Scan(&administrator.Id, &administrator.Forename, &administrator.Surname, &hash)
	if queryError != nil {
		fmt.Println(queryError)
		api.SendErrorResponse(context, http.StatusUnauthorized, "Credentials Invalid")
		return
	}

	fmt.Println("HERE")

	passwordValid, passwordError := cryptography.Verify(request.Password, hash)
	if passwordError != nil || passwordValid == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Credentials Invalid")
		return
	}

	token, tokenError := environment.KeyManager.Sign(
		authentication.UserTypeAdministrator,
		administrator.Id,
		administrator.Email,
		administrator.Forename,
		administrator.Surname,
	)
	if tokenError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Generate Token")
		return
	}

	context.JSON(http.StatusOK, model.LoginResponse{Token: token})
}

// @Summary      Login
// @Description  Returns a JWT (Json Web Token) which will expire in 24 hours.
// @Tags         Authentication
// @Accept       json
// @Param        body body model.LoginRequest true "Login Request"
// @Produce      json
// @Success      200  {object} model.LoginResponse
// @Failure      400  {object} model.ErrorResponse "Malformed Request Payload"
// @Failure      401  {object} model.ErrorResponse "Invalid Credentials"
// @Failure      500  {object} model.ErrorResponse "Internal Server Error"
// @Router       /login [post]
func (environment *Environment) LoginHandler(context *gin.Context) {
	var request model.LoginRequest

	if bindError := context.ShouldBindJSON(&request); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	request.Email = strings.ToUpper(request.Email)

	var customer model.Customer
	customer.Email = request.Email
	var hash string

	queryError := environment.Database.QueryRow(
		context,
		"SELECT id, forename, surname, hash FROM customers WHERE email = $1",
		request.Email,
	).Scan(&customer.Id, &customer.Forename, &customer.Surname, &hash)
	if queryError != nil {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Credentials Invalid")
		return
	}

	passwordValid, passwordError := cryptography.Verify(request.Password, hash)
	if passwordError != nil || passwordValid == false {
		api.SendErrorResponse(context, http.StatusUnauthorized, "Credentials Invalid")
		return
	}

	token, tokenError := environment.KeyManager.Sign(
		authentication.UserTypeAdministrator,
		customer.Id,
		customer.Email,
		customer.Forename,
		customer.Surname,
	)
	if tokenError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Generate Token")
		return
	}

	context.JSON(http.StatusOK, model.LoginResponse{Token: token})
}

// @Summary      Register
// @Description  Returns a JWT (Json Web Token) which will expire in 24 hours.
// @Tags         Authentication
// @Accept       json
// @Param        body body model.RegisterRequest true "Login Request"
// @Produce      json
// @Success      200  {object} model.LoginResponse
// @Failure      400  {object} model.ErrorResponse
// @Failure      409  {object} model.ErrorResponse
// @Failure      500  {object} model.ErrorResponse
// @Router       /register [post]
func (environment *Environment) RegisterHandler(context *gin.Context) {
	var request model.RegisterRequest

	if bindError := context.ShouldBindJSON(&request); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	request.Email = strings.ToUpper(request.Email)

	// Validate required fields
	if request.Email == "" || request.Password == "" || request.Forename == "" || request.Surname == "" {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	// Hash the password
	hash, hashError := cryptography.Hash(request.Password)
	if hashError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Hash Password")
		return
	}

	// Check if customer already exists
	var exists bool
	existsError := environment.Database.QueryRow(
		context,
		"SELECT EXISTS(SELECT 1 FROM customers WHERE email = $1)",
		request.Email,
	).Scan(&exists)
	if existsError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Database Error")
		return
	}

	if exists == true {
		api.SendErrorResponse(context, http.StatusConflict, "Email Already Registered")
		return
	}

	// Create new customer with a unique id
	var customer model.Customer
	customer.Id = uuid.New().String()
	customer.Email = request.Email
	customer.Forename = request.Forename
	customer.Surname = request.Surname

	insertError := environment.Database.QueryRow(
		context,
		"INSERT INTO customers (id, email, forename, surname, hash) VALUES ($1, $2, $3, $4, $5) RETURNING id, email, forename, surname",
		customer.Id,
		customer.Email,
		customer.Forename,
		customer.Surname,
		hash,
	).Scan(&customer.Id, &customer.Email, &customer.Forename, &customer.Surname)
	if insertError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Create Customer")
		return
	}

	token, tokenError := environment.KeyManager.Sign(
		authentication.UserTypeAdministrator,
		customer.Id,
		customer.Email,
		customer.Forename,
		customer.Surname,
	)
	if tokenError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Generate Token")
		return
	}

	context.JSON(http.StatusOK, model.LoginResponse{Token: token})
}

func (environment *Environment) CreateHashHandler(context *gin.Context) {
	var request model.HashRequest

	if bindError := context.ShouldBindJSON(&request); bindError != nil {
		api.SendErrorResponse(context, http.StatusBadRequest, "Malformed Request Body")
		return
	}

	// Hash the password
	hash, hashError := cryptography.Hash(request.Password)
	if hashError != nil {
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Hash Password")
		return
	}

	context.JSON(http.StatusOK, model.HashResponse{Hash: hash})
}
