package model

type Station struct {
	Tiploc    string  `json:"tiploc"`
	Nlc       string  `json:"nlc"`
	Name      string  `json:"name"`
	Crs       string  `json:"crs"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
} // @name Station

type StationServiceCount struct {
	Tiploc       string  `json:"tiploc"`
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ServiceCount int     `json:"serviceCount"`
} // @name StationServiceCount
