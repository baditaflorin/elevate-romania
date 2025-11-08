package main

import (
	"fmt"
	"math"
)

// ElementCluster represents a group of OSM elements that are geographically close
type ElementCluster struct {
	Elements []OSMElement
	BBox     BoundingBox
	Centroid Coordinates
}

// elementWithCoord is a helper struct for clustering
type elementWithCoord struct {
	element OSMElement
	coord   Coordinates
}

// ClusterElements groups OSM elements by geographic proximity to avoid OSM changeset
// bounding box size limits. Uses a grid-based approach for efficiency.
func ClusterElements(elements []OSMElement, maxBBoxDiagonal float64) []ElementCluster {
	if len(elements) == 0 {
		return []ElementCluster{}
	}

	extractor := NewCoordinateExtractor()
	
	// Extract coordinates for all elements
	var elementsWithCoords []elementWithCoord
	for _, elem := range elements {
		if coord, valid := extractor.Extract(elem); valid {
			elementsWithCoords = append(elementsWithCoords, elementWithCoord{elem, coord})
		}
	}
	
	if len(elementsWithCoords) == 0 {
		return []ElementCluster{}
	}
	
	// Calculate grid cell size based on maxBBoxDiagonal
	// Use half the max diagonal to ensure cells can merge if needed
	cellSize := maxBBoxDiagonal / 2.0
	
	// Create grid-based clusters
	gridClusters := make(map[string][]elementWithCoord)
	
	for _, ewc := range elementsWithCoords {
		// Calculate grid cell for this coordinate
		cellLat := math.Floor(ewc.coord.Lat / cellSize)
		cellLon := math.Floor(ewc.coord.Lon / cellSize)
		cellKey := fmt.Sprintf("%d,%d", int(cellLat), int(cellLon))
		
		gridClusters[cellKey] = append(gridClusters[cellKey], ewc)
	}
	
	// Convert grid clusters to ElementCluster objects
	var clusters []ElementCluster
	for _, cellElements := range gridClusters {
		if len(cellElements) == 0 {
			continue
		}
		
		elements := make([]OSMElement, len(cellElements))
		coords := make([]Coordinates, len(cellElements))
		for i, ewc := range cellElements {
			elements[i] = ewc.element
			coords[i] = ewc.coord
		}
		
		bbox := NewBoundingBox(coords)
		centroid := Centroid(coords)
		
		// Check if this cluster's bounding box is acceptable
		if bbox.Diagonal() <= maxBBoxDiagonal {
			clusters = append(clusters, ElementCluster{
				Elements: elements,
				BBox:     bbox,
				Centroid: centroid,
			})
		} else {
			// If a single grid cell is still too large, split it further
			subClusters := splitLargeCluster(cellElements, maxBBoxDiagonal)
			clusters = append(clusters, subClusters...)
		}
	}
	
	return clusters
}

// splitLargeCluster splits a cluster that's still too large into smaller clusters
// using a simple k-means-like approach
func splitLargeCluster(elements []elementWithCoord, maxBBoxDiagonal float64) []ElementCluster {
	// If we have very few elements, just split them individually
	if len(elements) <= 2 {
		var clusters []ElementCluster
		for _, ewc := range elements {
			clusters = append(clusters, ElementCluster{
				Elements: []OSMElement{ewc.element},
				BBox:     NewBoundingBox([]Coordinates{ewc.coord}),
				Centroid: ewc.coord,
			})
		}
		return clusters
	}
	
	// Calculate how many clusters we need based on diagonal
	coords := make([]Coordinates, len(elements))
	for i, ewc := range elements {
		coords[i] = ewc.coord
	}
	bbox := NewBoundingBox(coords)
	currentDiagonal := bbox.Diagonal()
	
	// Estimate number of clusters needed (add safety margin)
	numClusters := int(math.Ceil(currentDiagonal/maxBBoxDiagonal)) + 1
	if numClusters < 2 {
		numClusters = 2
	}
	
	// Simple k-means clustering
	clusters := simpleKMeans(elements, numClusters, maxBBoxDiagonal)
	
	return clusters
}

// simpleKMeans performs a simple k-means clustering on elements
func simpleKMeans(elements []elementWithCoord, k int, maxBBoxDiagonal float64) []ElementCluster {
	if len(elements) <= k {
		// If we have fewer elements than clusters, one per cluster
		var clusters []ElementCluster
		for _, ewc := range elements {
			clusters = append(clusters, ElementCluster{
				Elements: []OSMElement{ewc.element},
				BBox:     NewBoundingBox([]Coordinates{ewc.coord}),
				Centroid: ewc.coord,
			})
		}
		return clusters
	}
	
	// Initialize centroids by spreading them across the space
	coords := make([]Coordinates, len(elements))
	for i, ewc := range elements {
		coords[i] = ewc.coord
	}
	bbox := NewBoundingBox(coords)
	
	centroids := make([]Coordinates, k)
	for i := 0; i < k; i++ {
		// Distribute centroids evenly across the bounding box
		t := float64(i) / float64(k-1)
		if k == 1 {
			t = 0.5
		}
		centroids[i] = Coordinates{
			Lat: bbox.MinLat + t*(bbox.MaxLat-bbox.MinLat),
			Lon: bbox.MinLon + t*(bbox.MaxLon-bbox.MinLon),
		}
	}
	
	// Run k-means iterations (limit to prevent infinite loops)
	maxIterations := 10
	var assignments [][]elementWithCoord
	
	for iter := 0; iter < maxIterations; iter++ {
		// Assign elements to nearest centroid
		assignments = make([][]elementWithCoord, k)
		for _, ewc := range elements {
			nearestIdx := 0
			minDist := HaversineDistance(ewc.coord, centroids[0])
			
			for i := 1; i < k; i++ {
				dist := HaversineDistance(ewc.coord, centroids[i])
				if dist < minDist {
					minDist = dist
					nearestIdx = i
				}
			}
			
			assignments[nearestIdx] = append(assignments[nearestIdx], ewc)
		}
		
		// Update centroids
		converged := true
		for i := 0; i < k; i++ {
			if len(assignments[i]) == 0 {
				continue
			}
			
			clusterCoords := make([]Coordinates, len(assignments[i]))
			for j, ewc := range assignments[i] {
				clusterCoords[j] = ewc.coord
			}
			
			newCentroid := Centroid(clusterCoords)
			if HaversineDistance(centroids[i], newCentroid) > 0.001 {
				converged = false
			}
			centroids[i] = newCentroid
		}
		
		if converged {
			break
		}
	}
	
	// Create final clusters from assignments
	var finalClusters []ElementCluster
	for i := 0; i < k; i++ {
		if len(assignments[i]) == 0 {
			continue
		}
		
		clusterElements := make([]OSMElement, len(assignments[i]))
		clusterCoords := make([]Coordinates, len(assignments[i]))
		for j, ewc := range assignments[i] {
			clusterElements[j] = ewc.element
			clusterCoords[j] = ewc.coord
		}
		
		finalClusters = append(finalClusters, ElementCluster{
			Elements: clusterElements,
			BBox:     NewBoundingBox(clusterCoords),
			Centroid: Centroid(clusterCoords),
		})
	}
	
	return finalClusters
}
