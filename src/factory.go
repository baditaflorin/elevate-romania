package main

import (
	"net/http"
	"time"
)

// APIClientFactory creates configured API clients
type APIClientFactory struct {
	config *Config
	logger Logger
}

// NewAPIClientFactory creates a new API client factory
func NewAPIClientFactory(config *Config, logger Logger) *APIClientFactory {
	return &APIClientFactory{
		config: config,
		logger: logger,
	}
}

// CreateElevationEnricher creates a configured elevation enricher
func (f *APIClientFactory) CreateElevationEnricher(apiType string) *ElevationEnricher {
	rateLimit := float64(f.config.GetInt("API_RATE_LIMIT_MS"))
	if rateLimit == 0 {
		rateLimit = 1000 // Default 1 second
	}
	
	e := &ElevationEnricher{
		APIType:        apiType,
		RateLimit:      time.Duration(rateLimit * float64(time.Millisecond)),
		coordExtractor: NewCoordinateExtractor(),
	}
	
	// Use configured URL or default
	if apiType == "opentopo" {
		e.BaseURL = f.config.Get("OPENTOPO_URL")
		if e.BaseURL == "" {
			e.BaseURL = "https://api.opentopodata.org/v1/srtm30m"
		}
	} else {
		e.BaseURL = "https://api.open-elevation.com/api/v1/lookup"
	}
	
	return e
}

// CreateBatchElevationEnricher creates a configured batch elevation enricher
func (f *APIClientFactory) CreateBatchElevationEnricher(apiType string) *BatchElevationEnricher {
	rateLimit := float64(f.config.GetInt("API_RATE_LIMIT_MS"))
	if rateLimit == 0 {
		rateLimit = 1000 // Default 1 second
	}
	
	batchSize := f.config.GetInt("BATCH_SIZE")
	if batchSize == 0 {
		batchSize = 100 // Default
	}
	
	timeout := time.Duration(f.config.GetInt("API_TIMEOUT_SEC")) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	e := &BatchElevationEnricher{
		APIType:        apiType,
		RateLimit:      time.Duration(rateLimit * float64(time.Millisecond)),
		BatchSize:      batchSize,
		coordExtractor: NewCoordinateExtractor(),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
	
	// Use configured URL or default
	if apiType == "opentopo" {
		e.BaseURL = f.config.Get("OPENTOPO_URL")
		if e.BaseURL == "" {
			e.BaseURL = "https://api.opentopodata.org/v1/srtm30m"
		}
	} else {
		e.BaseURL = "https://api.open-elevation.com/api/v1/lookup"
	}
	
	return e
}

// CreateOverpassExtractor creates a configured Overpass extractor
func (f *APIClientFactory) CreateOverpassExtractor() *OverpassExtractor {
	url := f.config.Get("OVERPASS_URL")
	if url == "" {
		url = "https://overpass-api.de/api/interpreter"
	}
	
	return &OverpassExtractor{
		OverpassURL: url,
	}
}

// CreateOSMAPIClient creates a configured OSM API client
func (f *APIClientFactory) CreateOSMAPIClient(client *http.Client, dryRun bool) *OSMAPIClient {
	return NewOSMAPIClient(client, dryRun)
}
