package main

import "testing"

func TestConfigSetAndGet(t *testing.T) {
	config := NewConfig()
	
	config.Set("TEST_KEY", "test_value")
	
	if got := config.Get("TEST_KEY"); got != "test_value" {
		t.Errorf("Get() = %v, want %v", got, "test_value")
	}
}

func TestConfigSetDefault(t *testing.T) {
	config := NewConfig()
	
	// Set default value
	config.SetDefault("KEY1", "default")
	if got := config.Get("KEY1"); got != "default" {
		t.Errorf("SetDefault() = %v, want %v", got, "default")
	}
	
	// Try to set default again (should not override)
	config.SetDefault("KEY1", "new_default")
	if got := config.Get("KEY1"); got != "default" {
		t.Errorf("SetDefault() should not override, got %v, want %v", got, "default")
	}
	
	// But Set should override
	config.Set("KEY1", "override")
	if got := config.Get("KEY1"); got != "override" {
		t.Errorf("Set() = %v, want %v", got, "override")
	}
}

func TestConfigGetInt(t *testing.T) {
	config := NewConfig()
	
	tests := []struct {
		name     string
		value    string
		expected int
	}{
		{"Valid integer", "123", 123},
		{"Zero", "0", 0},
		{"Invalid integer", "abc", 0},
		{"Empty string", "", 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Set("TEST_INT", tt.value)
			if got := config.GetInt("TEST_INT"); got != tt.expected {
				t.Errorf("GetInt() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfigGetFloat(t *testing.T) {
	config := NewConfig()
	
	tests := []struct {
		name     string
		value    string
		expected float64
	}{
		{"Valid float", "123.45", 123.45},
		{"Integer as float", "100", 100.0},
		{"Invalid float", "abc", 0.0},
		{"Empty string", "", 0.0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Set("TEST_FLOAT", tt.value)
			if got := config.GetFloat("TEST_FLOAT"); got != tt.expected {
				t.Errorf("GetFloat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfigGetBool(t *testing.T) {
	config := NewConfig()
	
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"True string", "true", true},
		{"False string", "false", false},
		{"True number", "1", true},
		{"False number", "0", false},
		{"Invalid bool", "abc", false},
		{"Empty string", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Set("TEST_BOOL", tt.value)
			if got := config.GetBool("TEST_BOOL"); got != tt.expected {
				t.Errorf("GetBool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	config := NewConfig()
	config.Set("KEY1", "value1")
	config.Set("KEY2", "value2")
	
	tests := []struct {
		name        string
		required    []string
		expectError bool
	}{
		{"All present", []string{"KEY1", "KEY2"}, false},
		{"One missing", []string{"KEY1", "KEY3"}, true},
		{"All missing", []string{"KEY3", "KEY4"}, true},
		{"Empty required", []string{}, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.Validate(tt.required)
			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
