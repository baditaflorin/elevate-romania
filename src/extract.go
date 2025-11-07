package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type OverpassExtractor struct {
	OverpassURL string
	Country     string
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

func NewOverpassExtractor(country string) *OverpassExtractor {
	return &OverpassExtractor{
		OverpassURL: "https://overpass-api.de/api/interpreter",
		Country:     country,
	}
}

// escapeCountryName escapes double quotes in country name to prevent query injection
func escapeCountryName(country string) string {
	return strings.ReplaceAll(country, `"`, `\"`)
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
	escapedCountry := escapeCountryName(e.Country)
	query := fmt.Sprintf(`
[out:json][timeout:180];
area["name"="%s"]["admin_level"="2"]->.country;
(
  node["railway"="station"]["ele"!~".*"](area.country);
  node["railway"="halt"]["ele"!~".*"](area.country);
);
out body;
`, escapedCountry)

	fmt.Printf("Querying train stations in %s...\n", e.Country)
	elements, err := e.queryOverpass(query)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d train stations\n", len(elements))
	return elements, nil
}

func (e *OverpassExtractor) GetAccommodations() ([]OSMElement, error) {
	escapedCountry := escapeCountryName(e.Country)
	query := fmt.Sprintf(`
[out:json][timeout:300];
area["name"="%s"]["admin_level"="2"]->.country;
(
  node["tourism"="hotel"]["ele"!~".*"](area.country);
  node["tourism"="guest_house"]["ele"!~".*"](area.country);
  node["tourism"="alpine_hut"]["ele"!~".*"](area.country);
  node["tourism"="chalet"]["ele"!~".*"](area.country);
  node["tourism"="hostel"]["ele"!~".*"](area.country);
  node["tourism"="motel"]["ele"!~".*"](area.country);
  way["tourism"="hotel"]["ele"!~".*"](area.country);
  way["tourism"="guest_house"]["ele"!~".*"](area.country);
  way["tourism"="alpine_hut"]["ele"!~".*"](area.country);
  way["tourism"="chalet"]["ele"!~".*"](area.country);
  way["tourism"="hostel"]["ele"!~".*"](area.country);
  way["tourism"="motel"]["ele"!~".*"](area.country);
);
out center;
`, escapedCountry)

	fmt.Printf("Querying accommodations in %s...\n", e.Country)
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

func runExtract(country string) error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Printf("STEP 1: EXTRACT - Querying Overpass API for %s\n", country)
	fmt.Println(string(repeat('=', 60)))

	// Initialize configuration and factory
	config := NewConfig()
	config.LoadFromEnv()
	config.Set("COUNTRY", country)
	logger := NewLogger("Extractor")
	factory := NewAPIClientFactory(config, logger)

	// Create extractor using factory
	extractor := factory.CreateOverpassExtractor()
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

// CountryInfo holds information about a country
type CountryInfo struct {
	Name    string `json:"name"`
	IntName string `json:"int_name,omitempty"`
}

// runListCountries queries and lists all available admin_level=2 countries
func runListCountries() error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("Available Countries (admin_level=2)")
	fmt.Println(string(repeat('=', 60)))

	extractor := &OverpassExtractor{
		OverpassURL: "https://overpass-api.de/api/interpreter",
	}

	query := `
[out:json][timeout:60];
area["admin_level"="2"];
out tags;
`

	fmt.Println("Querying Overpass API for all countries...")
	
	client := &http.Client{
		Timeout: 2 * time.Minute,
	}

	resp, err := client.Post(
		extractor.OverpassURL,
		"application/x-www-form-urlencoded",
		bytes.NewBufferString("data="+query),
	)
	if err != nil {
		return fmt.Errorf("failed to query Overpass API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Overpass API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Elements []struct {
			Tags map[string]string `json:"tags"`
		} `json:"elements"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Collect unique countries
	countriesMap := make(map[string]CountryInfo)
	for _, element := range result.Elements {
		if name, ok := element.Tags["name"]; ok && name != "" {
			country := CountryInfo{
				Name: name,
			}
			if intName, ok := element.Tags["int_name"]; ok && intName != "" {
				country.IntName = intName
			}
			countriesMap[name] = country
		}
	}

	// Collect and sort countries by name
	var countries []CountryInfo
	for _, country := range countriesMap {
		countries = append(countries, country)
	}
	
	// Sort countries alphabetically by name
	sort.Slice(countries, func(i, j int) bool {
		return countries[i].Name < countries[j].Name
	})

	fmt.Printf("\nFound %d countries:\n\n", len(countries))
	
	// Display in columns
	for _, country := range countries {
		if country.IntName != "" && country.IntName != country.Name {
			fmt.Printf("  %-40s (int_name: %s)\n", country.Name, country.IntName)
		} else {
			fmt.Printf("  %s\n", country.Name)
		}
	}

	fmt.Println("\nUsage: elevate-romania --country \"Country Name\" --extract")
	fmt.Println("Note: Use the exact name (case-sensitive) as shown above")
	fmt.Println("\n" + string(repeat('=', 60)) + "\n")

	return nil
}
