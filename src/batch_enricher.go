package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BatchElevationEnricher handles batch elevation requests
type BatchElevationEnricher struct {
	APIType    string
	RateLimit  time.Duration
	BaseURL    string
	BatchSize  int
	httpClient *http.Client
}

// LocationRequest represents a location to fetch elevation for
type LocationRequest struct {
	Lat     float64
	Lon     float64
	Element *OSMElement
}

// BatchElevationResult represents the result of a batch elevation request
type BatchElevationResult struct {
	Elevation *float64
	Error     error
	Element   *OSMElement
}

// OpenTopoDataBatchResponse represents the response from OpenTopoData API
type OpenTopoDataBatchResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Elevation float64 `json:"elevation"`
		Location  struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"results"`
}

// NewBatchElevationEnricher creates a new batch enricher
func NewBatchElevationEnricher(apiType string, rateLimit float64, batchSize int) *BatchElevationEnricher {
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 100 // Default to max supported by OpenTopoData
	}

	e := &BatchElevationEnricher{
		APIType:   apiType,
		RateLimit: time.Duration(rateLimit * float64(time.Millisecond)),
		BatchSize: batchSize,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Note: Using direct API endpoint instead of proxy for better reliability
	// The proxy URL (go.proxy.okssh.com) was causing DNS resolution issues
	if apiType == "opentopo" {
		e.BaseURL = "https://api.opentopodata.org/v1/srtm30m"
	} else {
		e.BaseURL = "https://api.open-elevation.com/api/v1/lookup"
	}

	return e
}

// BatchGetElevations fetches elevations for multiple locations in a single API call
func (e *BatchElevationEnricher) BatchGetElevations(locations []LocationRequest) ([]BatchElevationResult, error) {
	if len(locations) == 0 {
		return []BatchElevationResult{}, nil
	}

	if e.APIType != "opentopo" {
		return nil, fmt.Errorf("batch mode only supported for opentopo API")
	}

	// Build the locations parameter: "lat1,lon1|lat2,lon2|..."
	var locationParts []string
	for _, loc := range locations {
		locationParts = append(locationParts, fmt.Sprintf("%.6f,%.6f", loc.Lat, loc.Lon))
	}
	locationsParam := strings.Join(locationParts, "|")

	// Make the API request with properly encoded query parameter
	requestURL := fmt.Sprintf("%s?locations=%s", e.BaseURL, url.QueryEscape(locationsParam))
	resp, err := e.httpClient.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch elevations: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("elevation API returned status %d", resp.StatusCode)
	}

	var result OpenTopoDataBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %v", err)
	}

	if result.Status != "OK" {
		return nil, fmt.Errorf("API returned non-OK status: %s", result.Status)
	}

	// Match results back to input locations
	results := make([]BatchElevationResult, len(locations))
	for i, loc := range locations {
		if i < len(result.Results) {
			elevation := result.Results[i].Elevation
			results[i] = BatchElevationResult{
				Elevation: &elevation,
				Error:     nil,
				Element:   loc.Element,
			}
		} else {
			results[i] = BatchElevationResult{
				Elevation: nil,
				Error:     fmt.Errorf("no elevation data returned for location %d", i),
				Element:   loc.Element,
			}
		}
	}

	return results, nil
}

// EnrichElementsBatch enriches multiple elements using batch API calls
func (e *BatchElevationEnricher) EnrichElementsBatch(elements []OSMElement, maxCount int) []OSMElement {
	var enriched []OSMElement
	var locationsToFetch []LocationRequest

	// Prepare locations for batch processing
	for i := range elements {
		if maxCount > 0 && i >= maxCount {
			break
		}

		element := elements[i]

		// Get coordinates
		var lat, lon float64
		var valid bool

		if element.Type == "node" {
			lat, lon = element.Lat, element.Lon
			valid = lat != 0 && lon != 0
		} else if element.Type == "way" && element.Center != nil {
			lat, lon = element.Center.Lat, element.Center.Lon
			valid = lat != 0 && lon != 0
		}

		if !valid {
			fmt.Printf("Warning: element %d has no valid coordinates\n", element.ID)
			continue
		}

		locationsToFetch = append(locationsToFetch, LocationRequest{
			Lat:     lat,
			Lon:     lon,
			Element: &element,
		})
	}

	// Process in batches
	totalLocations := len(locationsToFetch)
	for i := 0; i < totalLocations; i += e.BatchSize {
		end := i + e.BatchSize
		if end > totalLocations {
			end = totalLocations
		}

		batch := locationsToFetch[i:end]
		batchNum := (i / e.BatchSize) + 1
		totalBatches := (totalLocations + e.BatchSize - 1) / e.BatchSize

		fmt.Printf("Processing batch %d/%d (%d locations)...\n", batchNum, totalBatches, len(batch))

		results, err := e.BatchGetElevations(batch)
		if err != nil {
			fmt.Printf("Warning: batch request failed: %v\n", err)
			// Continue to next batch instead of failing completely
			continue
		}

		// Apply results to elements
		for _, result := range results {
			if result.Error != nil {
				fmt.Printf("Warning: failed to get elevation for element %d: %v\n", result.Element.ID, result.Error)
				continue
			}

			if result.Elevation != nil {
				// Create a new element with elevation data
				enrichedElement := *result.Element
				if enrichedElement.Tags == nil {
					enrichedElement.Tags = make(map[string]string)
				}
				enrichedElement.Tags["ele"] = fmt.Sprintf("%.1f", *result.Elevation)
				enrichedElement.Tags["ele:source"] = "SRTM"
				enrichedElement.ElevationFetched = result.Elevation

				enriched = append(enriched, enrichedElement)
			}
		}

		// Rate limiting between batches
		if end < totalLocations {
			time.Sleep(e.RateLimit)
		}
	}

	fmt.Printf("Successfully enriched %d/%d elements\n", len(enriched), totalLocations)

	return enriched
}
