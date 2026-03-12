package main

import (
	"context"
	"log"
	"time"
	"vulpz/train-api/src/authentication"
	"vulpz/train-api/src/configuration"
	"vulpz/train-api/src/email"
	"vulpz/train-api/src/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patrickmn/go-cache"
)

// @title           Train API
// @version         1.0
// @description     This is simple REST API for a UK train data.
// @host            api.train.vulpz.com
// @BasePath        /
func main() {
	// Load the environment
	executionEnvironment := configuration.LoadEnvironment()

	// Connect to database
	var database *pgxpool.Pool
	var databaseError error
	context := context.Background()

	database, databaseError = pgxpool.New(
		context,
		executionEnvironment.DatabaseConnectionString,
	)
	if databaseError != nil {
		log.Fatal(databaseError)
	}
	defer database.Close()

	// Load public and private keys
	keyManager, keyError := authentication.NewKeyManager()
	if keyError != nil {
		log.Fatal(keyError)
	}

	// Create email client
	emailClient := email.EmailClient{
		From:        "Train API <train@nightfoxdev.com>",
		Username:    "george@nightfoxdev.com",
		AppPassword: executionEnvironment.EmailAppPassword,
		SMTPHost:    "smtppro.zoho.com",
		SMTPPort:    "587",
	}

	environment := &handlers.Environment{
		Database:         database,
		Cache:            cache.New(5*time.Minute, 10*time.Minute),
		KeyManager:       keyManager,
		EmailClient:      &emailClient,
		OpenRouterAPIKey: executionEnvironment.OpenRouterAPIKey,
	}

	router := gin.Default()

	// Add CORS headers dependent on current environment
	if executionEnvironment.IsDevelopment() {
		router.Use(DevelopmentCORSMiddleware())
		router.GET("/hash", environment.CreateHashHandler)
	} else {
		router.Use(CORSMiddleware())
	}

	// Provide documentation endpoint
	router.GET("/documentation", handlers.DocumentationHandler)

	// Provide unprotected endpoints
	router.GET("/health", handlers.HealthHandler)
	router.POST("/administration/login", environment.AdministrationLoginHandler)
	router.POST("/login", environment.LoginHandler)
	router.POST("/register", environment.RegisterHandler)
	router.DELETE("/password", environment.ResetPasswordHandler)
	router.PUT("/password", environment.UpdatePasswordHandler)

	// Create a protected router group behind user authentication layer
	protectedRouterGroup := router.Group("/")
	protectedRouterGroup.Use(keyManager.Middleware())
	protectedRouterGroup.GET("/customer/:customerId", environment.GetCustomerByIdHandler)
	protectedRouterGroup.POST("/customer/:customerId/application", environment.CreateApplicationHandler)
	protectedRouterGroup.GET("/customer/:customerId/application", environment.GetApplicationsHandler)
	protectedRouterGroup.GET("/customer/:customerId/application/:applicationId", environment.GetApplicationHandler)
	protectedRouterGroup.DELETE("/customer/:customerId/application/:applicationId", environment.DeleteApplicationHandler)
	protectedRouterGroup.POST("/customer/:customerId/application/:applicationId/key/refresh", environment.RefreshApplicationKeyHandler)

	// Create an administration group to protect behind authentication layer
	administrationRouterGroup := router.Group("/administration")
	administrationRouterGroup.Use(keyManager.Middleware())
	administrationRouterGroup.Use(keyManager.AdministrationMiddleware())
	administrationRouterGroup.GET("/customer", environment.GetCustomersHandler)
	administrationRouterGroup.POST("/customer", environment.CreateCustomerHandler)
	administrationRouterGroup.GET("/customer/:customerId", environment.GetCustomerHandler)
	administrationRouterGroup.GET("/customer/:customerId/application", environment.GetCustomerApplicationsHandler)
	administrationRouterGroup.POST("/customer/:customerId/application/:applicationId/approve", environment.ApproveApplicationHandler)

	// Create a router group for product endpoints that require an API key
	const productRequestsPerMinute = 120
	productRouterGroup := router.Group("/")
	productRouterGroup.Use(keyManager.ApplicationKeyMiddleware(database))
	// productRouterGroup.Use(keyManager.ApplicationRateLimitMiddleware(productRequestsPerMinute))
	productRouterGroup.GET("/station", environment.GetStationsHandler)
	productRouterGroup.GET("/station/:stationId", environment.GetStationHandler)
	productRouterGroup.GET("/station/:stationId/analysis", environment.GetStationAnalysisHandler)
	productRouterGroup.GET("/stations.geojson", environment.GetStationsGeoJSONHandler)
	productRouterGroup.POST("/prompt", environment.PromptHandler)

	router.Run(":" + executionEnvironment.Port)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "train.vulpz.com")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}

		context.Next()
	}
}

func CORSDocumentationMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "documentation.train.vulpz.com")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}

		context.Next()
	}
}

func DevelopmentCORSMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}

		context.Next()
	}
}
