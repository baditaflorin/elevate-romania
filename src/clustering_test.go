package main

import (
	"testing"
)

func TestNewBoundingBox(t *testing.T) {
	tests := []struct {
		name     string
		coords   []Coordinates
		expected BoundingBox
	}{
		{
			name: "Single coordinate",
			coords: []Coordinates{
				{Lat: 45.5, Lon: 25.5},
			},
			expected: BoundingBox{MinLat: 45.5, MaxLat: 45.5, MinLon: 25.5, MaxLon: 25.5},
		},
		{
			name: "Multiple coordinates",
			coords: []Coordinates{
				{Lat: 45.0, Lon: 25.0},
				{Lat: 46.0, Lon: 26.0},
				{Lat: 44.5, Lon: 24.5},
			},
			expected: BoundingBox{MinLat: 44.5, MaxLat: 46.0, MinLon: 24.5, MaxLon: 26.0},
		},
		{
			name:     "Empty coordinates",
			coords:   []Coordinates{},
			expected: BoundingBox{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bbox := NewBoundingBox(tt.coords)
			if bbox != tt.expected {
				t.Errorf("NewBoundingBox() = %v, want %v", bbox, tt.expected)
			}
		})
	}
}

func TestBoundingBoxArea(t *testing.T) {
	bbox := BoundingBox{MinLat: 45.0, MaxLat: 46.0, MinLon: 25.0, MaxLon: 26.0}
	area := bbox.Area()
	expected := 1.0 // (46-45) * (26-25) = 1
	if area != expected {
		t.Errorf("BoundingBox.Area() = %f, want %f", area, expected)
	}
}

func TestBoundingBoxDiagonal(t *testing.T) {
	tests := []struct {
		name     string
		bbox     BoundingBox
		expected float64
	}{
		{
			name:     "Square bounding box",
			bbox:     BoundingBox{MinLat: 45.0, MaxLat: 46.0, MinLon: 25.0, MaxLon: 26.0},
			expected: 1.4142135623730951, // sqrt(2)
		},
		{
			name:     "Zero-size bounding box",
			bbox:     BoundingBox{MinLat: 45.0, MaxLat: 45.0, MinLon: 25.0, MaxLon: 25.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagonal := tt.bbox.Diagonal()
			if diagonal != tt.expected {
				t.Errorf("BoundingBox.Diagonal() = %f, want %f", diagonal, tt.expected)
			}
		})
	}
}

func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name     string
		c1       Coordinates
		c2       Coordinates
		minDist  float64 // minimum expected distance
		maxDist  float64 // maximum expected distance
	}{
		{
			name:    "Same point",
			c1:      Coordinates{Lat: 45.0, Lon: 25.0},
			c2:      Coordinates{Lat: 45.0, Lon: 25.0},
			minDist: 0.0,
			maxDist: 0.0,
		},
		{
			name:    "Approximately 111km (1 degree latitude difference)",
			c1:      Coordinates{Lat: 45.0, Lon: 25.0},
			c2:      Coordinates{Lat: 46.0, Lon: 25.0},
			minDist: 110.0,
			maxDist: 112.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := HaversineDistance(tt.c1, tt.c2)
			if dist < tt.minDist || dist > tt.maxDist {
				t.Errorf("HaversineDistance() = %f, want between %f and %f", dist, tt.minDist, tt.maxDist)
			}
		})
	}
}

func TestCentroid(t *testing.T) {
	tests := []struct {
		name     string
		coords   []Coordinates
		expected Coordinates
	}{
		{
			name:     "Single point",
			coords:   []Coordinates{{Lat: 45.0, Lon: 25.0}},
			expected: Coordinates{Lat: 45.0, Lon: 25.0},
		},
		{
			name: "Two points",
			coords: []Coordinates{
				{Lat: 44.0, Lon: 24.0},
				{Lat: 46.0, Lon: 26.0},
			},
			expected: Coordinates{Lat: 45.0, Lon: 25.0},
		},
		{
			name:     "Empty coordinates",
			coords:   []Coordinates{},
			expected: Coordinates{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			centroid := Centroid(tt.coords)
			if centroid != tt.expected {
				t.Errorf("Centroid() = %v, want %v", centroid, tt.expected)
			}
		})
	}
}

func TestClusterElements(t *testing.T) {
	tests := []struct {
		name             string
		elements         []OSMElement
		maxBBoxDiagonal  float64
		expectedClusters int
		checkFunc        func(*testing.T, []ElementCluster)
	}{
		{
			name:             "Empty elements",
			elements:         []OSMElement{},
			maxBBoxDiagonal:  0.5,
			expectedClusters: 0,
		},
		{
			name: "Single element",
			elements: []OSMElement{
				{ID: 1, Type: "node", Lat: 45.0, Lon: 25.0, Tags: map[string]string{"tourism": "alpine_hut"}},
			},
			maxBBoxDiagonal:  0.5,
			expectedClusters: 1,
			checkFunc: func(t *testing.T, clusters []ElementCluster) {
				if len(clusters[0].Elements) != 1 {
					t.Errorf("Expected 1 element in cluster, got %d", len(clusters[0].Elements))
				}
			},
		},
		{
			name: "Two nearby elements - same cluster",
			elements: []OSMElement{
				{ID: 1, Type: "node", Lat: 45.0, Lon: 25.0, Tags: map[string]string{"tourism": "alpine_hut"}},
				{ID: 2, Type: "node", Lat: 45.01, Lon: 25.01, Tags: map[string]string{"railway": "station"}},
			},
			maxBBoxDiagonal:  0.5,
			expectedClusters: 1,
			checkFunc: func(t *testing.T, clusters []ElementCluster) {
				if len(clusters[0].Elements) != 2 {
					t.Errorf("Expected 2 elements in cluster, got %d", len(clusters[0].Elements))
				}
			},
		},
		{
			name: "Two far elements - different clusters",
			elements: []OSMElement{
				{ID: 1, Type: "node", Lat: 45.0, Lon: 25.0, Tags: map[string]string{"tourism": "alpine_hut"}},
				{ID: 2, Type: "node", Lat: 48.0, Lon: 28.0, Tags: map[string]string{"railway": "station"}},
			},
			maxBBoxDiagonal:  0.5,
			expectedClusters: 2,
			checkFunc: func(t *testing.T, clusters []ElementCluster) {
				for _, cluster := range clusters {
					if len(cluster.Elements) != 1 {
						t.Errorf("Expected 1 element per cluster, got %d", len(cluster.Elements))
					}
					// Check that bounding box diagonal is within limits
					if cluster.BBox.Diagonal() > 0.5 {
						t.Errorf("Cluster bounding box diagonal %f exceeds maximum 0.5", cluster.BBox.Diagonal())
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusters := ClusterElements(tt.elements, tt.maxBBoxDiagonal)
			if len(clusters) != tt.expectedClusters {
				t.Errorf("ClusterElements() returned %d clusters, want %d", len(clusters), tt.expectedClusters)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, clusters)
			}
		})
	}
}

func TestCategoryToKey(t *testing.T) {
	tests := []struct {
		name     string
		category ElementCategory
		expected string
	}{
		{
			name:     "Alpine hut",
			category: CategoryAlpineHut,
			expected: "alpine_huts",
		},
		{
			name:     "Train station",
			category: CategoryTrainStation,
			expected: "train_stations",
		},
		{
			name:     "Other accommodation",
			category: CategoryOtherAccommodation,
			expected: "other_accommodations",
		},
		{
			name:     "Unknown category",
			category: CategoryUnknown,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categoryToKey(tt.category)
			if result != tt.expected {
				t.Errorf("categoryToKey() = %s, want %s", result, tt.expected)
			}
		})
	}
}
