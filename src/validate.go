package main

import (
	"fmt"
)

type ElevationValidator struct {
	MinElevation float64
	MaxElevation float64
}

type ValidationResult struct {
	Valid       bool     `json:"valid"`
	ElementID   int64    `json:"element_id"`
	ElementType string   `json:"element_type"`
	Elevation   *float64 `json:"elevation"`
	Errors      []string `json:"errors"`
}

type ValidationResults struct {
	Valid   []OSMElement     `json:"valid"`
	Invalid []InvalidElement `json:"invalid"`
}

type InvalidElement struct {
	Element    OSMElement       `json:"element"`
	Validation ValidationResult `json:"validation"`
}

type ValidatedCategory struct {
	ValidCount    int          `json:"valid_count"`
	InvalidCount  int          `json:"invalid_count"`
	ValidElements []OSMElement `json:"valid_elements"`
}

type ValidatedData struct {
	TrainStations       ValidatedCategory `json:"train_stations"`
	AlpineHuts          ValidatedCategory `json:"alpine_huts"`
	OtherAccommodations ValidatedCategory `json:"other_accommodations"`
}

func NewElevationValidator(minElevation, maxElevation float64) *ElevationValidator {
	return &ElevationValidator{
		MinElevation: minElevation,
		MaxElevation: maxElevation,
	}
}

func (v *ElevationValidator) ValidateElement(element OSMElement) ValidationResult {
	result := ValidationResult{
		Valid:       false,
		ElementID:   element.ID,
		ElementType: element.Type,
		Elevation:   nil,
		Errors:      []string{},
	}

	// Check if elevation exists
	if element.ElevationFetched == nil {
		result.Errors = append(result.Errors, "No elevation data")
		return result
	}

	elevation := *element.ElevationFetched
	result.Elevation = &elevation

	// Validate range
	if elevation < v.MinElevation {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Elevation %.1fm below minimum %.1fm", elevation, v.MinElevation))
	} else if elevation > v.MaxElevation {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Elevation %.1fm above maximum %.1fm", elevation, v.MaxElevation))
	} else {
		result.Valid = true
	}

	return result
}

func (v *ElevationValidator) ValidateElements(elements []OSMElement) ValidationResults {
	results := ValidationResults{
		Valid:   []OSMElement{},
		Invalid: []InvalidElement{},
	}

	for _, element := range elements {
		validation := v.ValidateElement(element)

		if validation.Valid {
			results.Valid = append(results.Valid, element)
		} else {
			results.Invalid = append(results.Invalid, InvalidElement{
				Element:    element,
				Validation: validation,
			})
		}
	}

	return results
}

func (v *ElevationValidator) ValidateAll(data *EnrichedData) map[string]ValidationResults {
	results := make(map[string]ValidationResults)

	categories := map[string][]OSMElement{
		"train_stations":       data.TrainStations,
		"alpine_huts":          data.AlpineHuts,
		"other_accommodations": data.OtherAccommodations,
	}

	for category, elements := range categories {
		if len(elements) > 0 {
			fmt.Printf("\nValidating %s...\n", category)
			validation := v.ValidateElements(elements)
			results[category] = validation

			fmt.Printf("  Valid: %d\n", len(validation.Valid))
			fmt.Printf("  Invalid: %d\n", len(validation.Invalid))

			// Show invalid examples
			if len(validation.Invalid) > 0 {
				fmt.Println("  Invalid examples:")
				for i, item := range validation.Invalid {
					if i >= 3 {
						break
					}
					val := item.Validation
					fmt.Printf("    - ID %d: %v\n", val.ElementID, val.Errors)
				}
			}
		}
	}

	return results
}

func runValidate() error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("STEP 4: VALIDATE - Checking elevation ranges (0-2600m)")
	fmt.Println(string(repeat('=', 60)))

	// Load enriched data
	var data EnrichedData
	if err := loadJSON("output/osm_data_enriched.json", &data); err != nil {
		return fmt.Errorf("output/osm_data_enriched.json not found. Run --enrich first: %v", err)
	}

	// Validate
	validator := NewElevationValidator(0, 2600)
	results := validator.ValidateAll(&data)

	// Save validation results
	output := ValidatedData{
		TrainStations: ValidatedCategory{
			ValidCount:    len(results["train_stations"].Valid),
			InvalidCount:  len(results["train_stations"].Invalid),
			ValidElements: results["train_stations"].Valid,
		},
		AlpineHuts: ValidatedCategory{
			ValidCount:    len(results["alpine_huts"].Valid),
			InvalidCount:  len(results["alpine_huts"].Invalid),
			ValidElements: results["alpine_huts"].Valid,
		},
		OtherAccommodations: ValidatedCategory{
			ValidCount:    len(results["other_accommodations"].Valid),
			InvalidCount:  len(results["other_accommodations"].Invalid),
			ValidElements: results["other_accommodations"].Valid,
		},
	}

	if err := saveJSON("output/osm_data_validated.json", output); err != nil {
		return err
	}

	fmt.Println("\n✓ Validation complete! Results saved to output/osm_data_validated.json")

	return nil
}
