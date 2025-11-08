package main

import (
	"fmt"
	"math"
)

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

// BoundingBox represents a geographic bounding box
type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// NewBoundingBox creates a bounding box from a set of coordinates
func NewBoundingBox(coords []Coordinates) BoundingBox {
	if len(coords) == 0 {
		return BoundingBox{}
	}
	
	bbox := BoundingBox{
		MinLat: coords[0].Lat,
		MaxLat: coords[0].Lat,
		MinLon: coords[0].Lon,
		MaxLon: coords[0].Lon,
	}
	
	for _, coord := range coords[1:] {
		if coord.Lat < bbox.MinLat {
			bbox.MinLat = coord.Lat
		}
		if coord.Lat > bbox.MaxLat {
			bbox.MaxLat = coord.Lat
		}
		if coord.Lon < bbox.MinLon {
			bbox.MinLon = coord.Lon
		}
		if coord.Lon > bbox.MaxLon {
			bbox.MaxLon = coord.Lon
		}
	}
	
	return bbox
}

// Area returns the approximate area of the bounding box in square degrees
func (bb BoundingBox) Area() float64 {
	return (bb.MaxLat - bb.MinLat) * (bb.MaxLon - bb.MinLon)
}

// Diagonal returns the diagonal distance across the bounding box in degrees
func (bb BoundingBox) Diagonal() float64 {
	latDiff := bb.MaxLat - bb.MinLat
	lonDiff := bb.MaxLon - bb.MinLon
	return math.Sqrt(latDiff*latDiff + lonDiff*lonDiff)
}

// HaversineDistance calculates the distance between two coordinates in kilometers
func HaversineDistance(c1, c2 Coordinates) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers
	
	lat1Rad := c1.Lat * math.Pi / 180
	lat2Rad := c2.Lat * math.Pi / 180
	deltaLat := (c2.Lat - c1.Lat) * math.Pi / 180
	deltaLon := (c2.Lon - c1.Lon) * math.Pi / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}

// Centroid calculates the geographic center of a set of coordinates
func Centroid(coords []Coordinates) Coordinates {
	if len(coords) == 0 {
		return Coordinates{}
	}
	
	var sumLat, sumLon float64
	for _, coord := range coords {
		sumLat += coord.Lat
		sumLon += coord.Lon
	}
	
	return Coordinates{
		Lat: sumLat / float64(len(coords)),
		Lon: sumLon / float64(len(coords)),
	}
}
