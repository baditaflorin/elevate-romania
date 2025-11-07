package main

import (
	"fmt"
	"regexp"
)

// Pre-compiled regex for elevation validation
var elevationRegex = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

// ElementValidationResult contains the result of element validation
type ElementValidationResult struct {
	Valid   bool
	Errors  []string
	Element OSMElement
}

// ElementValidatorImpl implements element validation logic
type ElementValidatorImpl struct {
	coordExtractor *CoordinateExtractor
	categorizer    *ElementCategorizer
}

// NewElementValidator creates a new element validator
func NewElementValidator() *ElementValidatorImpl {
	return &ElementValidatorImpl{
		coordExtractor: NewCoordinateExtractor(),
		categorizer:    NewElementCategorizer(),
	}
}

// Validate validates an OSM element
func (v *ElementValidatorImpl) Validate(element OSMElement) (bool, string) {
	var errors []string
	
	// Check element ID
	if element.ID == 0 {
		errors = append(errors, "element ID is zero")
	}
	
	// Check element type
	if element.Type != "node" && element.Type != "way" {
		errors = append(errors, fmt.Sprintf("invalid element type: %s", element.Type))
	}
	
	// Check coordinates
	if !v.coordExtractor.HasValidCoordinates(element) {
		errors = append(errors, "element has no valid coordinates")
	}
	
	// Check tags
	if element.Tags == nil || len(element.Tags) == 0 {
		errors = append(errors, "element has no tags")
	}
	
	if len(errors) > 0 {
		return false, fmt.Sprintf("validation failed: %v", errors)
	}
	
	return true, "validation passed"
}

// ValidateElevation validates elevation data on an element
func (v *ElementValidatorImpl) ValidateElevation(element OSMElement) (bool, string) {
	if !v.categorizer.HasElevation(element) {
		return false, "element has no elevation data"
	}
	
	// Check elevation format (should be numeric) using pre-compiled regex
	eleValue := element.Tags["ele"]
	if !elevationRegex.MatchString(eleValue) {
		return false, fmt.Sprintf("invalid elevation format: %s", eleValue)
	}
	
	// Check elevation source
	eleSource := element.Tags["ele:source"]
	if eleSource == "" {
		return false, "missing ele:source tag"
	}
	
	return true, "elevation validation passed"
}

// ValidateMultiple validates multiple elements
func (v *ElementValidatorImpl) ValidateMultiple(elements []OSMElement) []ElementValidationResult {
	results := make([]ElementValidationResult, len(elements))
	
	for i, element := range elements {
		valid, message := v.Validate(element)
		
		var errors []string
		if !valid {
			errors = append(errors, message)
		}
		
		results[i] = ElementValidationResult{
			Valid:   valid,
			Errors:  errors,
			Element: element,
		}
	}
	
	return results
}

// ValidateElevationData validates elevation data for multiple elements
func (v *ElementValidatorImpl) ValidateElevationData(elements []OSMElement) (valid, invalid []OSMElement) {
	for _, element := range elements {
		if isValid, _ := v.ValidateElevation(element); isValid {
			valid = append(valid, element)
		} else {
			invalid = append(invalid, element)
		}
	}
	return valid, invalid
}
