package main

import "testing"

func TestCoordinatesIsValid(t *testing.T) {
	tests := []struct {
		name     string
		coords   Coordinates
		expected bool
	}{
		{"Valid coordinates", Coordinates{Lat: 45.5, Lon: 25.5}, true},
		{"Zero lat", Coordinates{Lat: 0, Lon: 25.5}, false},
		{"Zero lon", Coordinates{Lat: 45.5, Lon: 0}, false},
		{"Both zero", Coordinates{Lat: 0, Lon: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.coords.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCoordinateExtractorExtract(t *testing.T) {
	extractor := NewCoordinateExtractor()

	tests := []struct {
		name        string
		element     OSMElement
		expectValid bool
		expectLat   float64
		expectLon   float64
	}{
		{
			name: "Valid node",
			element: OSMElement{
				Type: "node",
				Lat:  45.5,
				Lon:  25.5,
			},
			expectValid: true,
			expectLat:   45.5,
			expectLon:   25.5,
		},
		{
			name: "Valid way with center",
			element: OSMElement{
				Type:   "way",
				Center: &OSMCenter{Lat: 46.0, Lon: 26.0},
			},
			expectValid: true,
			expectLat:   46.0,
			expectLon:   26.0,
		},
		{
			name: "Node with zero coordinates",
			element: OSMElement{
				Type: "node",
				Lat:  0,
				Lon:  0,
			},
			expectValid: false,
		},
		{
			name: "Way without center",
			element: OSMElement{
				Type: "way",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coords, valid := extractor.Extract(tt.element)
			if valid != tt.expectValid {
				t.Errorf("Extract() valid = %v, want %v", valid, tt.expectValid)
			}
			if tt.expectValid {
				if coords.Lat != tt.expectLat || coords.Lon != tt.expectLon {
					t.Errorf("Extract() coords = (%.1f, %.1f), want (%.1f, %.1f)",
						coords.Lat, coords.Lon, tt.expectLat, tt.expectLon)
				}
			}
		})
	}
}

func TestCoordinateExtractorHasValidCoordinates(t *testing.T) {
	extractor := NewCoordinateExtractor()

	tests := []struct {
		name     string
		element  OSMElement
		expected bool
	}{
		{
			name: "Valid node",
			element: OSMElement{
				Type: "node",
				Lat:  45.5,
				Lon:  25.5,
			},
			expected: true,
		},
		{
			name: "Invalid node",
			element: OSMElement{
				Type: "node",
				Lat:  0,
				Lon:  0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractor.HasValidCoordinates(tt.element); got != tt.expected {
				t.Errorf("HasValidCoordinates() = %v, want %v", got, tt.expected)
			}
		})
	}
}
