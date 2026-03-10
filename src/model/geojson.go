package model

type GeoJSON struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
} // @name GeoJSON

type GeoJSONFeature struct {
	Type       string                   `json:"type"`
	Geometry   GeoJSONFeatureGeometry   `json:"geometry"`
	Properties GeoJSONFeatureProperties `json:"properties"`
} // @name GeoJSON Feature

type GeoJSONFeatureGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
} // @name GeoJSON Feature Geometry

type GeoJSONFeatureProperties struct {
	Id    string `json:"id"`
	Value int    `json:"value"`
} // @name GeoJSON Feature Properties
