package main

import (
	"fmt"
)

// ElevationFilter filters OSM elements based on elevation and coordinates
type ElevationFilter struct {
	coordExtractor  *CoordinateExtractor
	categorizer     *ElementCategorizer
}

// FilteredData contains categorized OSM elements
type FilteredData struct {
	TrainStations       []OSMElement `json:"train_stations"`
	AlpineHuts          []OSMElement `json:"alpine_huts"`
	OtherAccommodations []OSMElement `json:"other_accommodations"`
}

// NewElevationFilter creates a new elevation filter
func NewElevationFilter() *ElevationFilter {
	return &ElevationFilter{
		coordExtractor:  NewCoordinateExtractor(),
		categorizer:     NewElementCategorizer(),
	}
}

// filterMissingElevation filters elements without elevation data
func (f *ElevationFilter) filterMissingElevation(elements []OSMElement) []OSMElement {
	var result []OSMElement

	for _, element := range elements {
		if !f.categorizer.HasElevation(element) {
			if f.coordExtractor.HasValidCoordinates(element) {
				result = append(result, element)
			}
		}
	}

	return result
}

// prioritizeAlpineHuts separates alpine huts from other elements
func (f *ElevationFilter) prioritizeAlpineHuts(elements []OSMElement) ([]OSMElement, []OSMElement) {
	var alpineHuts []OSMElement
	var others []OSMElement

	for _, element := range elements {
		if f.categorizer.IsAlpineHut(element) {
			alpineHuts = append(alpineHuts, element)
		} else {
			others = append(others, element)
		}
	}

	return alpineHuts, others
}

// FilterData filters OSM data by elevation status and categorizes elements
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
	fmt.Println("✓ Filtered data saved to output/osm_data_filtered.json")

	return nil
}
