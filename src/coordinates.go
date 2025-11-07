package main

import "fmt"

// Coordinates represents a geographic coordinate pair
type Coordinates struct {
	Lat float64
	Lon float64
}

// IsValid checks if the coordinates are valid (non-zero)
func (c Coordinates) IsValid() bool {
	return c.Lat != 0 && c.Lon != 0
}

// String returns a string representation of the coordinates
func (c Coordinates) String() string {
	return fmt.Sprintf("%.6f,%.6f", c.Lat, c.Lon)
}

// CoordinateExtractor provides utilities for extracting coordinates from OSM elements
type CoordinateExtractor struct{}

// NewCoordinateExtractor creates a new coordinate extractor
func NewCoordinateExtractor() *CoordinateExtractor {
	return &CoordinateExtractor{}
}

// Extract extracts coordinates from an OSM element
// Returns coordinates and a boolean indicating if extraction was successful
func (ce *CoordinateExtractor) Extract(element OSMElement) (Coordinates, bool) {
	if element.Type == "node" {
		coords := Coordinates{Lat: element.Lat, Lon: element.Lon}
		return coords, coords.IsValid()
	}
	
	if element.Type == "way" && element.Center != nil {
		coords := Coordinates{Lat: element.Center.Lat, Lon: element.Center.Lon}
		return coords, coords.IsValid()
	}
	
	return Coordinates{}, false
}

// ExtractMultiple extracts coordinates from multiple elements
func (ce *CoordinateExtractor) ExtractMultiple(elements []OSMElement) []Coordinates {
	coords := make([]Coordinates, 0, len(elements))
	for _, element := range elements {
		if coord, valid := ce.Extract(element); valid {
			coords = append(coords, coord)
		}
	}
	return coords
}

// HasValidCoordinates checks if an element has valid coordinates
func (ce *CoordinateExtractor) HasValidCoordinates(element OSMElement) bool {
	_, valid := ce.Extract(element)
	return valid
}
