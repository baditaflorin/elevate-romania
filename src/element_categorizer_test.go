package main

import "testing"

func TestElementCategorizerCategorize(t *testing.T) {
	categorizer := NewElementCategorizer()

	tests := []struct {
		name     string
		element  OSMElement
		expected ElementCategory
	}{
		{
			name: "Alpine hut",
			element: OSMElement{
				Tags: map[string]string{"tourism": "alpine_hut"},
			},
			expected: CategoryAlpineHut,
		},
		{
			name: "Train station",
			element: OSMElement{
				Tags: map[string]string{"railway": "station"},
			},
			expected: CategoryTrainStation,
		},
		{
			name: "Train halt",
			element: OSMElement{
				Tags: map[string]string{"railway": "halt"},
			},
			expected: CategoryTrainStation,
		},
		{
			name: "Hotel",
			element: OSMElement{
				Tags: map[string]string{"tourism": "hotel"},
			},
			expected: CategoryOtherAccommodation,
		},
		{
			name: "Guest house",
			element: OSMElement{
				Tags: map[string]string{"tourism": "guest_house"},
			},
			expected: CategoryOtherAccommodation,
		},
		{
			name: "Unknown element",
			element: OSMElement{
				Tags: map[string]string{"building": "yes"},
			},
			expected: CategoryUnknown,
		},
		{
			name:     "Element without tags",
			element:  OSMElement{},
			expected: CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := categorizer.Categorize(tt.element); got != tt.expected {
				t.Errorf("Categorize() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestElementCategorizerIsAlpineHut(t *testing.T) {
	categorizer := NewElementCategorizer()

	tests := []struct {
		name     string
		element  OSMElement
		expected bool
	}{
		{
			name: "Alpine hut",
			element: OSMElement{
				Tags: map[string]string{"tourism": "alpine_hut"},
			},
			expected: true,
		},
		{
			name: "Hotel",
			element: OSMElement{
				Tags: map[string]string{"tourism": "hotel"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := categorizer.IsAlpineHut(tt.element); got != tt.expected {
				t.Errorf("IsAlpineHut() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestElementCategorizerHasElevation(t *testing.T) {
	categorizer := NewElementCategorizer()

	tests := []struct {
		name     string
		element  OSMElement
		expected bool
	}{
		{
			name: "Element with elevation",
			element: OSMElement{
				Tags: map[string]string{"ele": "1000"},
			},
			expected: true,
		},
		{
			name: "Element without elevation",
			element: OSMElement{
				Tags: map[string]string{"name": "Test"},
			},
			expected: false,
		},
		{
			name:     "Element without tags",
			element:  OSMElement{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := categorizer.HasElevation(tt.element); got != tt.expected {
				t.Errorf("HasElevation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestElementCategorizerCategorizeMultiple(t *testing.T) {
	categorizer := NewElementCategorizer()

	elements := []OSMElement{
		{Tags: map[string]string{"tourism": "alpine_hut"}},
		{Tags: map[string]string{"tourism": "hotel"}},
		{Tags: map[string]string{"railway": "station"}},
		{Tags: map[string]string{"tourism": "alpine_hut"}},
	}

	result := categorizer.CategorizeMultiple(elements)

	if len(result[CategoryAlpineHut]) != 2 {
		t.Errorf("Expected 2 alpine huts, got %d", len(result[CategoryAlpineHut]))
	}
	if len(result[CategoryOtherAccommodation]) != 1 {
		t.Errorf("Expected 1 other accommodation, got %d", len(result[CategoryOtherAccommodation]))
	}
	if len(result[CategoryTrainStation]) != 1 {
		t.Errorf("Expected 1 train station, got %d", len(result[CategoryTrainStation]))
	}
}
