package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ElevationEnricher struct {
	APIType   string
	RateLimit time.Duration
	BaseURL   string
}

type OpenTopoDataResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Elevation float64 `json:"elevation"`
	} `json:"results"`
}

type OpenElevationResponse struct {
	Results []struct {
		Elevation float64 `json:"elevation"`
	} `json:"results"`
}

func NewElevationEnricher(apiType string, rateLimit float64) *ElevationEnricher {
	e := &ElevationEnricher{
		APIType:   apiType,
		RateLimit: time.Duration(rateLimit * float64(time.Millisecond)),
	}
	if apiType == "opentopo" {
		e.BaseURL = "https://go.proxy.okssh.com/?url=https://api.opentopodata.org/v1/srtm30m"
	} else {
		e.BaseURL = "https://api.open-elevation.com/api/v1/lookup"
	}

	return e
}

func (e *ElevationEnricher) GetElevation(lat, lon float64) (*float64, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var resp *http.Response
	var err error

	if e.APIType == "opentopo" {
		url := fmt.Sprintf("%s?locations=%.6f,%.6f", e.BaseURL, lat, lon)
		resp, err = client.Get(url)
	} else {
		// Open-Elevation (not implemented in this example, but structure is here)
		return nil, fmt.Errorf("open-elevation not implemented yet")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch elevation for %.6f,%.6f: %v", lat, lon, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("elevation API returned status %d", resp.StatusCode)
	}

	var result OpenTopoDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if result.Status == "OK" && len(result.Results) > 0 {
		elevation := result.Results[0].Elevation
		return &elevation, nil
	}

	return nil, fmt.Errorf("no elevation data returned")
}

func (e *ElevationEnricher) EnrichElement(element OSMElement) (*OSMElement, error) {
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
		return nil, fmt.Errorf("no valid coordinates")
	}

	// Get elevation
	elevation, err := e.GetElevation(lat, lon)
	if err != nil {
		return nil, err
	}

	if elevation != nil {
		// Add elevation to element
		if element.Tags == nil {
			element.Tags = make(map[string]string)
		}
		element.Tags["ele"] = fmt.Sprintf("%.1f", *elevation)
		element.Tags["ele:source"] = "SRTM"
		element.ElevationFetched = elevation
	}

	// Rate limiting
	time.Sleep(e.RateLimit)

	return &element, nil
}

func (e *ElevationEnricher) EnrichElements(elements []OSMElement, maxCount int) []OSMElement {
	var enriched []OSMElement
	count := 0

	for _, element := range elements {
		if maxCount > 0 && count >= maxCount {
			break
		}

		enrichedElement, err := e.EnrichElement(element)
		if err != nil {
			fmt.Printf("Warning: failed to enrich element %d: %v\n", element.ID, err)
			continue
		}

		if enrichedElement != nil {
			enriched = append(enriched, *enrichedElement)
			count++
			if count%10 == 0 {
				fmt.Printf("Processed %d elements...\n", count)
			}
		}
	}

	return enriched
}

type EnrichedData struct {
	TrainStations       []OSMElement `json:"train_stations"`
	AlpineHuts          []OSMElement `json:"alpine_huts"`
	OtherAccommodations []OSMElement `json:"other_accommodations"`
}

func runEnrich(maxItems int) error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("STEP 3: ENRICH - Fetching elevation from OpenTopoData")
	fmt.Println(string(repeat('=', 60)))

	// Load filtered data
	var data FilteredData
	if err := loadJSON("output/osm_data_filtered.json", &data); err != nil {
		return fmt.Errorf("output/osm_data_filtered.json not found. Run --filter first: %v", err)
	}

	// Enrich with elevation
	enricher := NewElevationEnricher("opentopo", 0.1)

	enriched := &EnrichedData{
		TrainStations:       []OSMElement{},
		AlpineHuts:          []OSMElement{},
		OtherAccommodations: []OSMElement{},
	}

	// Process alpine huts first (priority)
	if len(data.AlpineHuts) > 0 {
		fmt.Println("\n[PRIORITY] Enriching alpine huts...")
		enriched.AlpineHuts = enricher.EnrichElements(data.AlpineHuts, maxItems)
	}

	// Process train stations
	if len(data.TrainStations) > 0 {
		fmt.Println("\nEnriching train stations...")
		enriched.TrainStations = enricher.EnrichElements(data.TrainStations, maxItems)
	}

	// Process other accommodations
	if len(data.OtherAccommodations) > 0 {
		fmt.Println("\nEnriching other accommodations...")
		enriched.OtherAccommodations = enricher.EnrichElements(data.OtherAccommodations, maxItems)
	}

	// Save enriched data
	if err := saveJSON("output/osm_data_enriched.json", enriched); err != nil {
		return err
	}

	fmt.Println("\n✓ Enrichment complete!")
	fmt.Printf("  Alpine huts: %d\n", len(enriched.AlpineHuts))
	fmt.Printf("  Train stations: %d\n", len(enriched.TrainStations))
	fmt.Printf("  Other accommodations: %d\n", len(enriched.OtherAccommodations))
	fmt.Println("✓ Enriched data saved to output/osm_data_enriched.json")

	return nil
}
