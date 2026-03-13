package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"vulpz/train-api/src/api"
	"vulpz/train-api/src/model"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
)

// @Summary      List Stations
// @Description  Responds with all stations, or the first 5 matches when using the optional search query
// @Tags         Stations
// @Produce      json
// @Param        search  query     string         false  "Search term matching station name or TIPLOC"
// @Success      200  {object}  []model.Station
// @Failure      401 {object}  model.ErrorResponse
// @Failure      403 {object}  model.ErrorResponse
// @Failure      429 {object}  model.ErrorResponse
// @Failure      500 {object}  model.ErrorResponse
// @Security     ApplicationKeyBearerAuthorisation
// @Router       /station [get]
func (environment *Environment) GetStationsHandler(context *gin.Context) {
	search := strings.TrimSpace(context.Query("search"))

	var rows pgx.Rows
	var databaseError error

	if search == "" {
		rows, databaseError = environment.Database.Query(context, "SELECT tiploc, nlc, name, crs, latitude, longitude FROM stations;")
	} else {
		searchQuery := "%" + search + "%"
		rows, databaseError = environment.Database.Query(
			context,
			`SELECT tiploc, nlc, name, crs, latitude, longitude
			FROM stations
			WHERE tiploc ILIKE $1 OR name ILIKE $1
			ORDER BY
				CASE
					WHEN tiploc ILIKE $1 THEN 0
					WHEN name ILIKE $1 THEN 1
					ELSE 2
				END,
				name ASC
			LIMIT 5;`,
			searchQuery,
		)
	}

	if databaseError != nil {
		log.Fatal(databaseError)
		api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		return
	}
	defer rows.Close()

	stationList := make([]model.Station, 0)

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
// @Failure      401 {object}  model.ErrorResponse
// @Failure      403 {object}  model.ErrorResponse
// @Failure      404        {object}  model.ErrorResponse
// @Failure      429 {object}  model.ErrorResponse
// @Failure      500        {object}  model.ErrorResponse
// @Security     ApplicationKeyBearerAuthorisation
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

// @Summary      Get Station Analysis
// @Description  Responds with station analysis data for the station matching the provided TIPLOC
// @Tags         Stations
// @Produce      json
// @Param        stationId  path      string                true  "Station Id (TIPLOC)"
// @Success      200        {object}  model.StationAnalysis
// @Failure      401 {object}  model.ErrorResponse
// @Failure      403 {object}  model.ErrorResponse
// @Failure      404        {object}  model.ErrorResponse
// @Failure      429 {object}  model.ErrorResponse
// @Failure      500        {object}  model.ErrorResponse
// @Security     ApplicationKeyBearerAuthorisation
// @Router       /station/{stationId}/analysis [get]
func (environment *Environment) GetStationAnalysisHandler(context *gin.Context) {
	stationId := context.Param("stationId")

	var stationAnalysis model.StationAnalysis
	queryError := environment.Database.QueryRow(
		context,
		"SELECT tiploc, service_count, delay_average, delay_rank, delay_average_commute, delay_rank_commute FROM stations_analysis WHERE tiploc = $1",
		stationId,
	).Scan(
		&stationAnalysis.Tiploc,
		&stationAnalysis.ServiceCount,
		&stationAnalysis.AverageDelay,
		&stationAnalysis.AverageDelayRank,
		&stationAnalysis.AverageDelayCommute,
		&stationAnalysis.AverageDelayCommuteRank,
	)
	if queryError != nil {
		if queryError == pgx.ErrNoRows {
			api.SendErrorResponse(context, http.StatusNotFound, "Station Analysis Not Found")
		} else {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Connect To Database")
		}
		return
	}

	context.JSON(http.StatusOK, stationAnalysis)
}

// @Summary      Get Station GeoJSON
// @Description  Responds with a GeoJSON representing station locations as features
// @Tags         Stations
// @Produce      json
// @Success      200        {object}  model.GeoJSON
// @Failure      401 {object}  model.ErrorResponse
// @Failure      403 {object}  model.ErrorResponse
// @Failure      429 {object}  model.ErrorResponse
// @Failure      500        {object}  model.ErrorResponse
// @Security     ApplicationKeyBearerAuthorisation
// @Router       /stations.geojson [get]
func (environment *Environment) GetStationsGeoJSONHandler(context *gin.Context) {
	const cacheKey = "StationsGeoJSON"

	if cached, found := environment.Cache.Get(cacheKey); found {
		context.Data(http.StatusOK, "application/json", cached.([]byte))
		return
	}

	rows, databaseError := environment.Database.Query(context, `
	SELECT
		stations.tiploc,
		stations.latitude,
		stations.longitude,
		COALESCE(stations_analysis.service_count, 0) AS service_count,
		COALESCE(stations_analysis.delay_average, 0) AS delay_average,
		COALESCE(stations_analysis.delay_average_commute, 0) AS delay_average_commute
	FROM stations
	LEFT JOIN stations_analysis ON stations_analysis.tiploc = stations.tiploc;
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
		var stationId string
		var latitude float64
		var longitude float64
		var serviceCount int
		var averageDelay float64
		var averageDelayCommute float64

		if scanError := rows.Scan(&stationId, &latitude, &longitude, &serviceCount, &averageDelay, &averageDelayCommute); scanError != nil {
			api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Parse Stations")
			return
		}

		output.Features = append(output.Features, model.GeoJSONFeature{
			Type: "Feature",
			Geometry: model.GeoJSONFeatureGeometry{
				Type:        "Point",
				Coordinates: []float64{longitude, latitude},
			},
			Properties: model.GeoJSONFeatureProperties{
				Id:                  stationId,
				ServiceCount:        serviceCount,
				AverageDelay:        averageDelay,
				AverageDelayCommute: averageDelayCommute,
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
