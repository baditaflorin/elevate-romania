package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type CSVExporter struct{}

type ElementInfo struct {
	Category        string
	Type            string
	ID              string
	Name            string
	Lat             string
	Lon             string
	Elevation       string
	ElevationSource string
	Tourism         string
	Railway         string
	OSMLink         string
}

func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

func (e *CSVExporter) getElementInfo(element OSMElement, category string) ElementInfo {
	info := ElementInfo{
		Category: category,
		Type:     element.Type,
		ID:       strconv.FormatInt(element.ID, 10),
	}

	// Get coordinates
	if element.Type == "node" {
		info.Lat = fmt.Sprintf("%.6f", element.Lat)
		info.Lon = fmt.Sprintf("%.6f", element.Lon)
	} else if element.Type == "way" && element.Center != nil {
		info.Lat = fmt.Sprintf("%.6f", element.Center.Lat)
		info.Lon = fmt.Sprintf("%.6f", element.Center.Lon)
	}

	// Get tags
	if element.Tags != nil {
		if name, ok := element.Tags["name"]; ok {
			info.Name = name
		} else if ref, ok := element.Tags["ref"]; ok {
			info.Name = ref
		}

		info.Elevation = element.Tags["ele"]
		info.ElevationSource = element.Tags["ele:source"]
		info.Tourism = element.Tags["tourism"]
		info.Railway = element.Tags["railway"]
	}

	// OSM link
	info.OSMLink = fmt.Sprintf("https://www.openstreetmap.org/%s/%d", element.Type, element.ID)

	return info
}

func (e *CSVExporter) ExportToCSV(data ValidatedData, outputFile string) (int, error) {
	var rows []ElementInfo

	// Process all categories
	categories := map[string][]OSMElement{
		"train_stations":       data.TrainStations.ValidElements,
		"alpine_huts":          data.AlpineHuts.ValidElements,
		"other_accommodations": data.OtherAccommodations.ValidElements,
	}

	for category, elements := range categories {
		for _, element := range elements {
			info := e.getElementInfo(element, category)
			rows = append(rows, info)
		}
	}

	if len(rows) == 0 {
		fmt.Println("No data to export")
		return 0, nil
	}

	// Create CSV file
	file, err := os.Create(outputFile)
	if err != nil {
		return 0, fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"category", "type", "id", "name", "lat", "lon",
		"elevation", "elevation_source", "tourism", "railway", "osm_link",
	}
	if err := writer.Write(header); err != nil {
		return 0, fmt.Errorf("failed to write header: %v", err)
	}

	// Write rows
	for _, row := range rows {
		record := []string{
			row.Category,
			row.Type,
			row.ID,
			row.Name,
			row.Lat,
			row.Lon,
			row.Elevation,
			row.ElevationSource,
			row.Tourism,
			row.Railway,
			row.OSMLink,
		}
		if err := writer.Write(record); err != nil {
			return 0, fmt.Errorf("failed to write row: %v", err)
		}
	}

	fmt.Printf("Exported %d elements to %s\n", len(rows), outputFile)
	return len(rows), nil
}

func runExportCSV() error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("STEP 5: EXPORT - Creating CSV output")
	fmt.Println(string(repeat('=', 60)))

	// Load validated data
	var data ValidatedData
	if err := loadJSON("output/osm_data_validated.json", &data); err != nil {
		return fmt.Errorf("output/osm_data_validated.json not found. Run --validate first: %v", err)
	}

	// Export to CSV
	exporter := NewCSVExporter()
	count, err := exporter.ExportToCSV(data, "output/elevation_data.csv")
	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Exported %d elements to output/elevation_data.csv\n\n", count)

	return nil
}
