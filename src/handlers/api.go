package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"vulpz/train-api/src/api"

	"github.com/gin-gonic/gin"
)

type Environment struct {
	Database *sql.DB
}

func HealthHandler(context *gin.Context) {
	context.Status(http.StatusOK)
}

type Station struct {
	Tiploc    string  `json:"tiploc"`
	Nlc       string  `json:"nlc"`
	Name      string  `json:"name"`
	Crs       string  `json:"crs"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GetStations returns the list of all railway stations.
// @Summary      Get all stations
// @Description  Responds with a list of all stations in the database
// @Tags         stations
// @Produce      json
// @Success      200  {array}  Station
// @Router       /stations [get]
func (environment *Environment) StationsHandler(context *gin.Context) {
	rows, databaseError := environment.Database.Query("SELECT tiploc, nlc, name, crs, latitude, longitude FROM stations")
	if databaseError != nil {
		log.Fatal(databaseError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}
	defer rows.Close()

	var stationList []Station

	for rows.Next() {
		var station Station

		if scanError := rows.Scan(&station.Tiploc, &station.Nlc, &station.Name, &station.Crs, &station.Latitude, &station.Longitude); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Stations")
			return
		}

		stationList = append(stationList, station)
	}

	context.JSON(http.StatusOK, stationList)
}
