package main

// ElementCategory represents different categories of OSM elements
type ElementCategory string

const (
	CategoryAlpineHut          ElementCategory = "alpine_hut"
	CategoryTrainStation       ElementCategory = "train_station"
	CategoryOtherAccommodation ElementCategory = "other_accommodation"
	CategoryUnknown            ElementCategory = "unknown"
)

// ElementCategorizer provides utilities for categorizing OSM elements
type ElementCategorizer struct{}

// NewElementCategorizer creates a new element categorizer
func NewElementCategorizer() *ElementCategorizer {
	return &ElementCategorizer{}
}

// Categorize determines the category of an OSM element
func (ec *ElementCategorizer) Categorize(element OSMElement) ElementCategory {
	if element.Tags == nil {
		return CategoryUnknown
	}
	
	// Check for alpine hut
	if element.Tags["tourism"] == "alpine_hut" {
		return CategoryAlpineHut
	}
	
	// Check for train station
	railway := element.Tags["railway"]
	if railway == "station" || railway == "halt" {
		return CategoryTrainStation
	}
	
	// Check for other accommodations
	tourism := element.Tags["tourism"]
	accommodationTypes := []string{"hotel", "guest_house", "chalet", "hostel", "motel"}
	for _, accType := range accommodationTypes {
		if tourism == accType {
			return CategoryOtherAccommodation
		}
	}
	
	return CategoryUnknown
}

// IsAlpineHut checks if an element is an alpine hut
func (ec *ElementCategorizer) IsAlpineHut(element OSMElement) bool {
	return ec.Categorize(element) == CategoryAlpineHut
}

// IsTrainStation checks if an element is a train station
func (ec *ElementCategorizer) IsTrainStation(element OSMElement) bool {
	return ec.Categorize(element) == CategoryTrainStation
}

// IsAccommodation checks if an element is any type of accommodation
func (ec *ElementCategorizer) IsAccommodation(element OSMElement) bool {
	category := ec.Categorize(element)
	return category == CategoryAlpineHut || category == CategoryOtherAccommodation
}

// HasElevation checks if an element has elevation data
func (ec *ElementCategorizer) HasElevation(element OSMElement) bool {
	if element.Tags == nil {
		return false
	}
	_, exists := element.Tags["ele"]
	return exists
}

// CategorizeMultiple categorizes multiple elements and groups them by category
func (ec *ElementCategorizer) CategorizeMultiple(elements []OSMElement) map[ElementCategory][]OSMElement {
	result := make(map[ElementCategory][]OSMElement)
	
	for _, element := range elements {
		category := ec.Categorize(element)
		result[category] = append(result[category], element)
	}
	
	return result
}
