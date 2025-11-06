package main

import (
	"fmt"
)

type ElevationFilter struct{}

type FilteredData struct {
	TrainStations       []OSMElement `json:"train_stations"`
	AlpineHuts          []OSMElement `json:"alpine_huts"`
	OtherAccommodations []OSMElement `json:"other_accommodations"`
}

func NewElevationFilter() *ElevationFilter {
	return &ElevationFilter{}
}

func (f *ElevationFilter) hasElevation(element OSMElement) bool {
	if element.Tags == nil {
		return false
	}
	_, exists := element.Tags["ele"]
	return exists
}

func (f *ElevationFilter) isAlpineHut(element OSMElement) bool {
	if element.Tags == nil {
		return false
	}
	return element.Tags["tourism"] == "alpine_hut"
}

func (f *ElevationFilter) getCoordinates(element OSMElement) (float64, float64, bool) {
	if element.Type == "node" {
		if element.Lat != 0 && element.Lon != 0 {
			return element.Lat, element.Lon, true
		}
	} else if element.Type == "way" && element.Center != nil {
		if element.Center.Lat != 0 && element.Center.Lon != 0 {
			return element.Center.Lat, element.Center.Lon, true
		}
	}
	return 0, 0, false
}

func (f *ElevationFilter) filterMissingElevation(elements []OSMElement) []OSMElement {
	var result []OSMElement

	for _, element := range elements {
		if !f.hasElevation(element) {
			if _, _, valid := f.getCoordinates(element); valid {
				result = append(result, element)
			}
		}
	}

	return result
}

func (f *ElevationFilter) prioritizeAlpineHuts(elements []OSMElement) ([]OSMElement, []OSMElement) {
	var alpineHuts []OSMElement
	var others []OSMElement

	for _, element := range elements {
		if f.isAlpineHut(element) {
			alpineHuts = append(alpineHuts, element)
		} else {
			others = append(others, element)
		}
	}

	return alpineHuts, others
}

func (f *ElevationFilter) FilterData(data *OSMData) *FilteredData {
	result := &FilteredData{
		TrainStations:       []OSMElement{},
		AlpineHuts:          []OSMElement{},
		OtherAccommodations: []OSMElement{},
	}

	// Filter train stations
	result.TrainStations = f.filterMissingElevation(data.TrainStations)

	// Filter accommodations and prioritize alpine huts
	missingEle := f.filterMissingElevation(data.Accommodations)
	alpineHuts, others := f.prioritizeAlpineHuts(missingEle)
	result.AlpineHuts = alpineHuts
	result.OtherAccommodations = others

	return result
}

func runFilter() error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("STEP 2: FILTER - Identifying elements without elevation")
	fmt.Println(string(repeat('=', 60)))

	// Load raw data
	var data OSMData
	if err := loadJSON("output/osm_data_raw.json", &data); err != nil {
		return fmt.Errorf("output/osm_data_raw.json not found. Run --extract first: %v", err)
	}

	// Filter
	filter := NewElevationFilter()
	filtered := filter.FilterData(&data)

	// Save filtered data
	if err := saveJSON("output/osm_data_filtered.json", filtered); err != nil {
		return err
	}

	fmt.Printf("\n✓ Train stations without elevation: %d\n", len(filtered.TrainStations))
	fmt.Printf("✓ Alpine huts without elevation: %d (PRIORITY)\n", len(filtered.AlpineHuts))
	fmt.Printf("✓ Other accommodations without elevation: %d\n", len(filtered.OtherAccommodations))
	fmt.Println("✓ Filtered data saved to output/osm_data_filtered.json\n")

	return nil
}
