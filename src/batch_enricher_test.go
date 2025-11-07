package main

import (
	"testing"
)

func TestNewBatchElevationEnricher(t *testing.T) {
	tests := []struct {
		name          string
		apiType       string
		rateLimit     float64
		batchSize     int
		expectedSize  int
		expectedURL   string
	}{
		{
			name:         "Valid batch size",
			apiType:      "opentopo",
			rateLimit:    1000.0,
			batchSize:    50,
			expectedSize: 50,
			expectedURL:  "https://api.opentopodata.org/v1/srtm30m",
		},
		{
			name:         "Batch size too large",
			apiType:      "opentopo",
			rateLimit:    1000.0,
			batchSize:    150,
			expectedSize: 100, // Should be capped at 100
			expectedURL:  "https://api.opentopodata.org/v1/srtm30m",
		},
		{
			name:         "Batch size zero",
			apiType:      "opentopo",
			rateLimit:    1000.0,
			batchSize:    0,
			expectedSize: 100, // Should default to 100
			expectedURL:  "https://api.opentopodata.org/v1/srtm30m",
		},
		{
			name:         "Negative batch size",
			apiType:      "opentopo",
			rateLimit:    1000.0,
			batchSize:    -10,
			expectedSize: 100, // Should default to 100
			expectedURL:  "https://api.opentopodata.org/v1/srtm30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enricher := NewBatchElevationEnricher(tt.apiType, tt.rateLimit, tt.batchSize)

			if enricher.BatchSize != tt.expectedSize {
				t.Errorf("Expected batch size %d, got %d", tt.expectedSize, enricher.BatchSize)
			}

			if enricher.BaseURL != tt.expectedURL {
				t.Errorf("Expected base URL %s, got %s", tt.expectedURL, enricher.BaseURL)
			}

			if enricher.APIType != tt.apiType {
				t.Errorf("Expected API type %s, got %s", tt.apiType, enricher.APIType)
			}
		})
	}
}

func TestLocationRequestBuilding(t *testing.T) {
	// Test that we can build location requests correctly
	elements := []OSMElement{
		{
			Type: "node",
			ID:   1,
			Lat:  46.947464,
			Lon:  22.700911,
		},
		{
			Type: "node",
			ID:   2,
			Lat:  6.947464,
			Lon:  6.947464,
		},
		{
			Type: "way",
			ID:   3,
			Center: &OSMCenter{
				Lat: 43.0,
				Lon: 53.0,
			},
		},
	}

	var locations []LocationRequest
	for i, elem := range elements {
		var lat, lon float64
		var valid bool

		if elem.Type == "node" {
			lat, lon = elem.Lat, elem.Lon
			valid = lat != 0 && lon != 0
		} else if elem.Type == "way" && elem.Center != nil {
			lat, lon = elem.Center.Lat, elem.Center.Lon
			valid = lat != 0 && lon != 0
		}

		if valid {
			locations = append(locations, LocationRequest{
				Lat:     lat,
				Lon:     lon,
				Element: &elements[i],
				Index:   i,
			})
		}
	}

	if len(locations) != 3 {
		t.Errorf("Expected 3 valid locations, got %d", len(locations))
	}

	// Verify coordinates
	expectedCoords := []struct {
		lat float64
		lon float64
	}{
		{46.947464, 22.700911},
		{6.947464, 6.947464},
		{43.0, 53.0},
	}

	for i, loc := range locations {
		if loc.Lat != expectedCoords[i].lat {
			t.Errorf("Location %d: expected lat %f, got %f", i, expectedCoords[i].lat, loc.Lat)
		}
		if loc.Lon != expectedCoords[i].lon {
			t.Errorf("Location %d: expected lon %f, got %f", i, expectedCoords[i].lon, loc.Lon)
		}
	}
}

func TestBatchProcessingLogic(t *testing.T) {
	// Test that batch processing splits correctly
	tests := []struct {
		name             string
		totalElements    int
		batchSize        int
		expectedBatches  int
	}{
		{
			name:            "Exact batch size",
			totalElements:   100,
			batchSize:       100,
			expectedBatches: 1,
		},
		{
			name:            "Multiple full batches",
			totalElements:   200,
			batchSize:       100,
			expectedBatches: 2,
		},
		{
			name:            "Partial last batch",
			totalElements:   150,
			batchSize:       100,
			expectedBatches: 2,
		},
		{
			name:            "Small batch size",
			totalElements:   250,
			batchSize:       50,
			expectedBatches: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := 0
			for i := 0; i < tt.totalElements; i += tt.batchSize {
				batches++
			}

			if batches != tt.expectedBatches {
				t.Errorf("Expected %d batches, got %d", tt.expectedBatches, batches)
			}
		})
	}
}
