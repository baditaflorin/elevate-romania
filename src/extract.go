package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OverpassExtractor struct {
	OverpassURL string
}

type OSMElement struct {
	Type             string            `json:"type"`
	ID               int64             `json:"id"`
	Lat              float64           `json:"lat,omitempty"`
	Lon              float64           `json:"lon,omitempty"`
	Center           *OSMCenter        `json:"center,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
	ElevationFetched *float64          `json:"elevation_fetched,omitempty"`
}

type OSMCenter struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type OverpassResponse struct {
	Elements []OSMElement `json:"elements"`
}

type OSMData struct {
	TrainStations  []OSMElement `json:"train_stations"`
	Accommodations []OSMElement `json:"accommodations"`
}

func NewOverpassExtractor() *OverpassExtractor {
	return &OverpassExtractor{
		OverpassURL: "https://overpass-api.de/api/interpreter",
	}
}

func (e *OverpassExtractor) queryOverpass(query string) ([]OSMElement, error) {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Post(
		e.OverpassURL,
		"application/x-www-form-urlencoded",
		bytes.NewBufferString("data="+query),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query Overpass API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Overpass API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result OverpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Elements, nil
}

func (e *OverpassExtractor) GetTrainStations() ([]OSMElement, error) {
	query := `
[out:json][timeout:180];
area["name"="România"]["admin_level"="2"]->.hunedoara;
(
  node["railway"="station"]["ele"!~".*"](area.hunedoara);
  node["railway"="halt"]["ele"!~".*"](area.hunedoara);
);
out body;
`

	fmt.Println("Querying train stations...")
	elements, err := e.queryOverpass(query)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d train stations\n", len(elements))
	return elements, nil
}

func (e *OverpassExtractor) GetAccommodations() ([]OSMElement, error) {
	query := `
[out:json][timeout:300];
area["name"="România"]["admin_level"="2"]->.romania;
(
  node["tourism"="hotel"]["ele"!~".*"](area.romania);
  node["tourism"="guest_house"]["ele"!~".*"](area.romania);
  node["tourism"="alpine_hut"]["ele"!~".*"](area.romania);
  node["tourism"="chalet"]["ele"!~".*"](area.romania);
  node["tourism"="hostel"]["ele"!~".*"](area.romania);
  node["tourism"="motel"]["ele"!~".*"](area.romania);
  way["tourism"="hotel"]["ele"!~".*"](area.romania);
  way["tourism"="guest_house"]["ele"!~".*"](area.romania);
  way["tourism"="alpine_hut"]["ele"!~".*"](area.romania);
  way["tourism"="chalet"]["ele"!~".*"](area.romania);
  way["tourism"="hostel"]["ele"!~".*"](area.romania);
  way["tourism"="motel"]["ele"!~".*"](area.romania);
);
out center;
`

	fmt.Println("Querying accommodations...")
	elements, err := e.queryOverpass(query)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d accommodations\n", len(elements))
	return elements, nil
}

func (e *OverpassExtractor) GetAllData() (*OSMData, error) {
	stations, err := e.GetTrainStations()
	if err != nil {
		return nil, err
	}

	// Be nice to Overpass API
	time.Sleep(2 * time.Second)

	accommodations, err := e.GetAccommodations()
	if err != nil {
		return nil, err
	}

	return &OSMData{
		TrainStations:  stations,
		Accommodations: accommodations,
	}, nil
}

func runExtract() error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("STEP 1: EXTRACT - Querying Overpass API")
	fmt.Println(string(repeat('=', 60)))

	extractor := NewOverpassExtractor()
	data, err := extractor.GetAllData()
	if err != nil {
		return err
	}

	// Save to file
	if err := saveJSON("output/osm_data_raw.json", data); err != nil {
		return err
	}

	fmt.Printf("\n✓ Extracted %d train stations\n", len(data.TrainStations))
	fmt.Printf("✓ Extracted %d accommodations\n", len(data.Accommodations))
	fmt.Println("✓ Data saved to output/osm_data_raw.json")

	return nil
}
