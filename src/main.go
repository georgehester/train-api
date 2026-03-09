package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
	"log"
	"time"
	"vulpz/train-api/src/handlers"
)

// @title           Train API
// @version         1.0
// @description     This is simple REST API for a UK train data.
// @host            api.train.vulpz.com
// @BasePath        /
func main() {
	var database *pgx.Conn
	var databaseError error
	context := context.Background()

	database, databaseError = pgx.Connect(
		context,
		"user=application password=password host=localhost port=5432 dbname=train sslmode=disable",
	)
	if databaseError != nil {
		log.Fatal(databaseError)
	}

	defer database.Close(context)

	environment := &handlers.Environment{
		Database: database,
		Cache:    cache.New(5*time.Minute, 10*time.Minute),
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
	protected.GET("/stations.geojson", environment.StationsGeoJSONHandler)

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
