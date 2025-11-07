package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config provides configuration management with defaults
type Config struct {
	values map[string]string
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	return &Config{
		values: make(map[string]string),
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	// OSM OAuth Configuration
	c.Set("OSM_CLIENT_ID", os.Getenv("OSM_CLIENT_ID"))
	c.Set("OSM_CLIENT_SECRET", os.Getenv("OSM_CLIENT_SECRET"))
	c.Set("OSM_ACCESS_TOKEN", os.Getenv("OSM_ACCESS_TOKEN"))
	
	// API Configuration
	c.SetDefault("OVERPASS_URL", "https://overpass-api.de/api/interpreter")
	c.SetDefault("OPENTOPO_URL", "https://api.opentopodata.org/v1/srtm30m")
	c.SetDefault("OSM_API_URL", "https://api.openstreetmap.org/api/0.6")
	
	// Rate Limiting
	c.SetDefault("API_RATE_LIMIT_MS", "1000")
	c.SetDefault("BATCH_SIZE", "100")
	c.SetDefault("API_TIMEOUT_SEC", "30")
	
	// OAuth
	c.SetDefault("OAUTH_REDIRECT_URI", "http://127.0.0.1:8080/callback")
}

// Get retrieves a configuration value
func (c *Config) Get(key string) string {
	return c.values[key]
}

// Set sets a configuration value
func (c *Config) Set(key, value string) {
	c.values[key] = value
}

// SetDefault sets a configuration value only if it doesn't exist or is empty
func (c *Config) SetDefault(key, value string) {
	existingValue, exists := c.values[key]
	if !exists || existingValue == "" {
		c.values[key] = value
	}
}

// GetInt retrieves a configuration value as an integer
func (c *Config) GetInt(key string) int {
	val := c.Get(key)
	if val == "" {
		return 0
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return i
}

// GetFloat retrieves a configuration value as a float64
func (c *Config) GetFloat(key string) float64 {
	val := c.Get(key)
	if val == "" {
		return 0.0
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	}
	return f
}

// GetBool retrieves a configuration value as a boolean
func (c *Config) GetBool(key string) bool {
	val := c.Get(key)
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return b
}

// Validate checks if required configuration values are present
func (c *Config) Validate(requiredKeys []string) error {
	missing := []string{}
	for _, key := range requiredKeys {
		if c.Get(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %v", missing)
	}
	return nil
}
