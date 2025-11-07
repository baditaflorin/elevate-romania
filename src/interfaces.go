package main

import "net/http"

// ElevationProvider defines the interface for fetching elevation data
type ElevationProvider interface {
	GetElevation(lat, lon float64) (*float64, error)
}

// BatchElevationProvider defines the interface for batch elevation fetching
type BatchElevationProvider interface {
	BatchGetElevations(locations []LocationRequest) ([]BatchElevationResult, error)
}

// DataExtractor defines the interface for extracting OSM data
type DataExtractor interface {
	GetAllData() (*OSMData, error)
}

// ElementFilter defines the interface for filtering OSM elements
type ElementFilter interface {
	FilterData(data *OSMData) *FilteredData
}

// ElementValidator defines the interface for validating elements
type ElementValidator interface {
	Validate(element OSMElement) (bool, string)
}

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// ConfigProvider defines the interface for configuration management
type ConfigProvider interface {
	Get(key string) string
	GetInt(key string) int
	GetFloat(key string) float64
	GetBool(key string) bool
}
