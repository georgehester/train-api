package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"vulpz/train-api/src/api"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
)

type Environment struct {
	Database *pgx.Conn
	Cache    *cache.Cache
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

type StationServiceCount struct {
	Tiploc       string  `json:"tiploc"`
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ServiceCount int     `json:"serviceCount"`
}

type GeoJSON struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                   `json:"type"`
	Geometry   GeoJSONFeatureGeometry   `json:"geometry"`
	Properties GeoJSONFeatureProperties `json:"properties"`
}

type GeoJSONFeatureGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type GeoJSONFeatureProperties struct {
	Id    string `json:"id"`
	Value int    `json:"value"`
}

// GetStations returns the list of all railway stations.
// @Summary      Get all stations
// @Description  Responds with a list of all stations in the database
// @Tags         stations
// @Produce      json
// @Success      200  {array}  Station
// @Router       /stations [get]
func (environment *Environment) StationsHandler(context *gin.Context) {
	rows, databaseError := environment.Database.Query(context, "SELECT tiploc, nlc, name, crs, latitude, longitude FROM stations;")
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

func (environment *Environment) StationsGeoJSONHandler(context *gin.Context) {
	const cacheKey = "StationsGeoJSON"

	if cached, found := environment.Cache.Get(cacheKey); found {
		context.Data(http.StatusOK, "application/json", cached.([]byte))
		return
	}

	rows, databaseError := environment.Database.Query(context, `
	SELECT stations.tiploc, stations.name, stations.latitude, stations.longitude, COUNT(stops.id) as count FROM stations
	LEFT JOIN stops ON stops.station_tiploc = stations.tiploc
	GROUP BY stations.tiploc;
	`)
	if databaseError != nil {
		log.Fatal(databaseError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}
	defer rows.Close()

	output := GeoJSON{
		Type:     "FeatureCollection",
		Features: []GeoJSONFeature{},
	}

	for rows.Next() {
		var station StationServiceCount

		if scanError := rows.Scan(&station.Tiploc, &station.Name, &station.Latitude, &station.Longitude, &station.ServiceCount); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Stations")
			return
		}

		output.Features = append(output.Features, GeoJSONFeature{
			Type: "Feature",
			Geometry: GeoJSONFeatureGeometry{
				Type:        "Point",
				Coordinates: []float64{station.Longitude, station.Latitude},
			},
			Properties: GeoJSONFeatureProperties{
				Id:    station.Tiploc,
				Value: station.ServiceCount,
			},
		})
	}

	data, jsonError := json.Marshal(output)
	if jsonError != nil {
		log.Fatal(jsonError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Encode GeoJSON")
		return
	}

	environment.Cache.Set(cacheKey, data, cache.DefaultExpiration)

	context.Data(http.StatusOK, "application/json", data)
}
