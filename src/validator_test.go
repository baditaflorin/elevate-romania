package main

import "testing"

func TestElementValidatorValidate(t *testing.T) {
	validator := NewElementValidator()
	
	tests := []struct {
		name        string
		element     OSMElement
		expectValid bool
	}{
		{
			name: "Valid node element",
			element: OSMElement{
				Type: "node",
				ID:   123,
				Lat:  45.5,
				Lon:  25.5,
				Tags: map[string]string{"name": "Test"},
			},
			expectValid: true,
		},
		{
			name: "Valid way element",
			element: OSMElement{
				Type:   "way",
				ID:     456,
				Center: &OSMCenter{Lat: 45.5, Lon: 25.5},
				Tags:   map[string]string{"name": "Test"},
			},
			expectValid: true,
		},
		{
			name: "Element with zero ID",
			element: OSMElement{
				Type: "node",
				ID:   0,
				Lat:  45.5,
				Lon:  25.5,
				Tags: map[string]string{"name": "Test"},
			},
			expectValid: false,
		},
		{
			name: "Element with invalid type",
			element: OSMElement{
				Type: "relation",
				ID:   123,
				Lat:  45.5,
				Lon:  25.5,
				Tags: map[string]string{"name": "Test"},
			},
			expectValid: false,
		},
		{
			name: "Element with no coordinates",
			element: OSMElement{
				Type: "node",
				ID:   123,
				Lat:  0,
				Lon:  0,
				Tags: map[string]string{"name": "Test"},
			},
			expectValid: false,
		},
		{
			name: "Element with no tags",
			element: OSMElement{
				Type: "node",
				ID:   123,
				Lat:  45.5,
				Lon:  25.5,
			},
			expectValid: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := validator.Validate(tt.element)
			if valid != tt.expectValid {
				t.Errorf("Validate() = %v, want %v", valid, tt.expectValid)
			}
		})
	}
}

func TestElementValidatorValidateElevation(t *testing.T) {
	validator := NewElementValidator()
	
	tests := []struct {
		name        string
		element     OSMElement
		expectValid bool
	}{
		{
			name: "Valid elevation",
			element: OSMElement{
				Tags: map[string]string{
					"ele":        "1234.5",
					"ele:source": "SRTM",
				},
			},
			expectValid: true,
		},
		{
			name: "Valid negative elevation",
			element: OSMElement{
				Tags: map[string]string{
					"ele":        "-10.5",
					"ele:source": "SRTM",
				},
			},
			expectValid: true,
		},
		{
			name: "Valid integer elevation",
			element: OSMElement{
				Tags: map[string]string{
					"ele":        "1000",
					"ele:source": "SRTM",
				},
			},
			expectValid: true,
		},
		{
			name: "Missing elevation",
			element: OSMElement{
				Tags: map[string]string{
					"name": "Test",
				},
			},
			expectValid: false,
		},
		{
			name: "Invalid elevation format",
			element: OSMElement{
				Tags: map[string]string{
					"ele":        "abc",
					"ele:source": "SRTM",
				},
			},
			expectValid: false,
		},
		{
			name: "Missing elevation source",
			element: OSMElement{
				Tags: map[string]string{
					"ele": "1000",
				},
			},
			expectValid: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := validator.ValidateElevation(tt.element)
			if valid != tt.expectValid {
				t.Errorf("ValidateElevation() = %v, want %v", valid, tt.expectValid)
			}
		})
	}
}

func TestElementValidatorValidateElevationData(t *testing.T) {
	validator := NewElementValidator()
	
	elements := []OSMElement{
		{Tags: map[string]string{"ele": "1000", "ele:source": "SRTM"}},
		{Tags: map[string]string{"ele": "abc", "ele:source": "SRTM"}},
		{Tags: map[string]string{"ele": "2000", "ele:source": "GPS"}},
		{Tags: map[string]string{"name": "Test"}},
	}
	
	valid, invalid := validator.ValidateElevationData(elements)
	
	if len(valid) != 2 {
		t.Errorf("Expected 2 valid elements, got %d", len(valid))
	}
	
	if len(invalid) != 2 {
		t.Errorf("Expected 2 invalid elements, got %d", len(invalid))
	}
}
