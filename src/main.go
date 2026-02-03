package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"vulpz/train-api/src/handlers"
)

// @title           Train API
// @version         1.0
// @description     This is simple REST API for a UK train data.
// @host            api.train.vulpz.com
// @BasePath        /
func main() {
	var database *sql.DB
	var databaseError error

	database, databaseError = sql.Open("sqlite3", "train.db")
	if databaseError != nil {
		log.Fatal(databaseError)
	}

	defer database.Close()

	environment := &handlers.Environment{
		Database: database,
	}

	router := gin.Default()

	router.Use(CORSMiddleware())

	// Provide public endpoints
	router.StaticFile("/documentation", "documentation/swagger.json")

	// Create a protected router group behind authentication
	protected := router.Group("/")
	protected.Use(CORSMiddleware())
	protected.GET("/health", handlers.HealthHandler)
	protected.GET("/stations", environment.StationsHandler)

	router.Run()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}

		context.Next()
	}
}
