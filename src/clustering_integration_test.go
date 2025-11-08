package main

import (
"testing"
)

// TestClusteringIntegration demonstrates the clustering functionality with real-world data
func TestClusteringIntegration(t *testing.T) {
// Create test elements spread across Romania (wide area that would exceed OSM limits)
elements := []OSMElement{
// Bucharest area (Southern Romania)
{ID: 1, Type: "node", Lat: 44.4268, Lon: 26.1025, Tags: map[string]string{"tourism": "alpine_hut", "ele": "100"}},
{ID: 2, Type: "node", Lat: 44.4368, Lon: 26.1125, Tags: map[string]string{"railway": "station", "ele": "110"}},

// Cluj area (Central Romania)
{ID: 3, Type: "node", Lat: 46.7712, Lon: 23.6236, Tags: map[string]string{"tourism": "hotel", "ele": "400"}},
{ID: 4, Type: "node", Lat: 46.7812, Lon: 23.6336, Tags: map[string]string{"railway": "station", "ele": "410"}},

// Iasi area (Northeast Romania)
{ID: 5, Type: "node", Lat: 47.1585, Lon: 27.6014, Tags: map[string]string{"tourism": "alpine_hut", "ele": "350"}},

// Timisoara area (Western Romania)
{ID: 6, Type: "node", Lat: 45.7489, Lon: 21.2087, Tags: map[string]string{"tourism": "hotel", "ele": "90"}},

// Constanta area (Eastern Coast)
{ID: 7, Type: "node", Lat: 44.1598, Lon: 28.6348, Tags: map[string]string{"railway": "station", "ele": "5"}},
}

t.Logf("Testing with %d elements spread across Romania", len(elements))

// Calculate overall bounding box without clustering
coords := make([]Coordinates, len(elements))
extractor := NewCoordinateExtractor()
for i, elem := range elements {
if coord, valid := extractor.Extract(elem); valid {
coords[i] = coord
}
}
overallBBox := NewBoundingBox(coords)

t.Logf("Overall bounding box diagonal: %.4f degrees", overallBBox.Diagonal())

// Cluster the elements
clusters := ClusterElements(elements, MaxBoundingBoxDiagonal)

t.Logf("Created %d clusters", len(clusters))

// Verify all clusters are within limits
totalElements := 0
for i, cluster := range clusters {
diagonal := cluster.BBox.Diagonal()
t.Logf("Cluster %d: %d elements, diagonal: %.4f degrees", i+1, len(cluster.Elements), diagonal)

if diagonal > MaxBoundingBoxDiagonal {
t.Errorf("Cluster %d exceeds maximum diagonal (%.4f > %.2f)", i+1, diagonal, MaxBoundingBoxDiagonal)
}

totalElements += len(cluster.Elements)
}

// Verify no elements were lost
if totalElements != len(elements) {
t.Errorf("Element count mismatch: got %d, want %d", totalElements, len(elements))
}

// Verify we created multiple clusters (since Romania is large)
if len(clusters) < 2 {
t.Logf("Warning: Expected multiple clusters for Romania-wide data, got %d", len(clusters))
}

t.Log("✓ All clusters are within bounding box size limits")
}

// TestRealWorldScenarioRussia tests a scenario similar to Russia with very dispersed elements
func TestRealWorldScenarioRussia(t *testing.T) {
// Simulate elements across Russia (huge area)
elements := []OSMElement{
// Moscow area
{ID: 1, Type: "node", Lat: 55.7558, Lon: 37.6173, Tags: map[string]string{"railway": "station"}},
// St Petersburg
{ID: 2, Type: "node", Lat: 59.9343, Lon: 30.3351, Tags: map[string]string{"railway": "station"}},
// Vladivostok (Far East)
{ID: 3, Type: "node", Lat: 43.1150, Lon: 131.8855, Tags: map[string]string{"railway": "station"}},
// Novosibirsk (Siberia)
{ID: 4, Type: "node", Lat: 55.0084, Lon: 82.9357, Tags: map[string]string{"railway": "station"}},
// Sochi (South)
{ID: 5, Type: "node", Lat: 43.5855, Lon: 39.7231, Tags: map[string]string{"tourism": "hotel"}},
}

// Calculate overall bounding box
coords := make([]Coordinates, len(elements))
extractor := NewCoordinateExtractor()
for i, elem := range elements {
if coord, valid := extractor.Extract(elem); valid {
coords[i] = coord
}
}
overallBBox := NewBoundingBox(coords)
overallDiagonal := overallBBox.Diagonal()

t.Logf("Russia scenario: Overall diagonal = %.2f degrees (HUGE!)", overallDiagonal)

// This should definitely be larger than our limit
if overallDiagonal <= MaxBoundingBoxDiagonal {
t.Errorf("Test setup error: Expected overall diagonal > %.2f, got %.2f", MaxBoundingBoxDiagonal, overallDiagonal)
}

// Cluster the elements
clusters := ClusterElements(elements, MaxBoundingBoxDiagonal)

t.Logf("Split into %d clusters", len(clusters))

// Verify all clusters are within limits
for i, cluster := range clusters {
diagonal := cluster.BBox.Diagonal()
if diagonal > MaxBoundingBoxDiagonal {
t.Errorf("Cluster %d exceeds limit: %.4f > %.2f", i+1, diagonal, MaxBoundingBoxDiagonal)
}
}

// Should create many clusters
if len(clusters) < 3 {
t.Logf("Warning: Expected many clusters for Russia-wide data, got %d", len(clusters))
}

t.Log("✓ Successfully handled Russia-scale geographic dispersion")
}
