package main

import (
	"strings"
	"testing"
)

func TestOverpassExtractorCountryParameter(t *testing.T) {
	tests := []struct {
		name            string
		country         string
		expectedInQuery string
	}{
		{
			name:            "Default Romania",
			country:         "România",
			expectedInQuery: `area["name"="România"]["admin_level"="2"]`,
		},
		{
			name:            "Moldova",
			country:         "Moldova",
			expectedInQuery: `area["name"="Moldova"]["admin_level"="2"]`,
		},
		{
			name:            "France",
			country:         "France",
			expectedInQuery: `area["name"="France"]["admin_level"="2"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewOverpassExtractor(tt.country)
			
			// Verify country is set
			if extractor.Country != tt.country {
				t.Errorf("Expected country %s, got %s", tt.country, extractor.Country)
			}
			
			// Verify URL is set
			if extractor.OverpassURL == "" {
				t.Error("Expected OverpassURL to be set")
			}
		})
	}
}

func TestOverpassExtractorGetTrainStationsQuery(t *testing.T) {
	extractor := NewOverpassExtractor("Moldova")
	
	// We can't actually call the API in tests, but we can verify the country is set
	if extractor.Country != "Moldova" {
		t.Errorf("Expected country Moldova, got %s", extractor.Country)
	}
}

func TestOverpassExtractorGetAccommodationsQuery(t *testing.T) {
	extractor := NewOverpassExtractor("France")
	
	// Verify the country is set correctly
	if extractor.Country != "France" {
		t.Errorf("Expected country France, got %s", extractor.Country)
	}
}

func TestCountryInfoStructure(t *testing.T) {
	country := CountryInfo{
		Name:    "România",
		IntName: "Romania",
	}
	
	if country.Name != "România" {
		t.Errorf("Expected name România, got %s", country.Name)
	}
	
	if country.IntName != "Romania" {
		t.Errorf("Expected int_name Romania, got %s", country.IntName)
	}
}

func TestNewOverpassExtractor(t *testing.T) {
	t.Run("Creates extractor with country", func(t *testing.T) {
		country := "TestCountry"
		extractor := NewOverpassExtractor(country)
		
		if extractor == nil {
			t.Fatal("Expected extractor to be created")
		}
		
		if extractor.Country != country {
			t.Errorf("Expected country %s, got %s", country, extractor.Country)
		}
		
		if !strings.HasPrefix(extractor.OverpassURL, "https://") {
			t.Errorf("Expected HTTPS URL, got %s", extractor.OverpassURL)
		}
	})
}
