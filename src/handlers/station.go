package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
)

// @Summary      List Stations
// @Description  Responds with a list of all stations in the database
// @Tags         Stations
// @Produce      json
// @Success      200  {array}  model.Station
// @Failure      500 {object}  model.ErrorResponse
// @Router       /stations [get]
func (environment *Environment) GetStationsHandler(context *gin.Context) {
	rows, databaseError := environment.Database.Query(context, "SELECT tiploc, nlc, name, crs, latitude, longitude FROM stations;")
	if databaseError != nil {
		log.Fatal(databaseError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}
	defer rows.Close()

	var stationList []model.Station

	for rows.Next() {
		var station model.Station

		if scanError := rows.Scan(&station.Tiploc, &station.Nlc, &station.Name, &station.Crs, &station.Latitude, &station.Longitude); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Stations")
			return
		}

		stationList = append(stationList, station)
	}

	context.JSON(http.StatusOK, stationList)
}

// @Summary      Get Station
// @Description  Responds with a single station in the database
// @Tags         Stations
// @Produce      json
// @Param        stationId  path      string         true  "Station Id (TIPLOC)"
// @Success      200        {object}  model.Station
// @Failure      404        {object}  model.ErrorResponse
// @Failure      500        {object}  model.ErrorResponse
// @Router       /station/{stationId} [get]
func (environment *Environment) GetStationHandler(context *gin.Context) {
	stationId := context.Param("stationId")

	var station model.Station
	queryError := environment.Database.QueryRow(
		context,
		"SELECT tiploc, nlc, name, crs, latitude, longitude FROM stations WHERE tiploc = $1",
		stationId,
	).Scan(&station.Tiploc, &station.Nlc, &station.Name, &station.Crs, &station.Latitude, &station.Longitude)
	if queryError != nil {
		if queryError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Station Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		}
		return
	}

	context.JSON(http.StatusOK, station)
}

// @Summary      Get Station GeoJSON
// @Description  Responds with a GeoJSON representing station locations as features
// @Tags         Stations
// @Produce      json
// @Success      200        {object}  model.GeoJSON
// @Failure      500        {object}  model.ErrorResponse
// @Router       /stations.geojson [get]
func (environment *Environment) GetStationsGeoJSONHandler(context *gin.Context) {
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

	output := model.GeoJSON{
		Type:     "FeatureCollection",
		Features: []model.GeoJSONFeature{},
	}

	for rows.Next() {
		var station model.StationServiceCount

		if scanError := rows.Scan(&station.Tiploc, &station.Name, &station.Latitude, &station.Longitude, &station.ServiceCount); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Stations")
			return
		}

		output.Features = append(output.Features, model.GeoJSONFeature{
			Type: "Feature",
			Geometry: model.GeoJSONFeatureGeometry{
				Type:        "Point",
				Coordinates: []float64{station.Longitude, station.Latitude},
			},
			Properties: model.GeoJSONFeatureProperties{
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
