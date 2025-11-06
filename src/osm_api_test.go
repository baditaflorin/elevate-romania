package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name         string
		existingTags []NodeTag
		newTags      map[string]string
		want         []NodeTag
	}{
		{
			name: "Add new tags to empty list",
			existingTags: []NodeTag{},
			newTags: map[string]string{
				"ele":        "100.5",
				"ele:source": "SRTM",
			},
			want: []NodeTag{
				{Key: "ele", Value: "100.5"},
				{Key: "ele:source", Value: "SRTM"},
			},
		},
		{
			name: "Update existing tag",
			existingTags: []NodeTag{
				{Key: "name", Value: "Test Station"},
				{Key: "railway", Value: "station"},
			},
			newTags: map[string]string{
				"ele":        "150.0",
				"ele:source": "SRTM",
			},
			want: []NodeTag{
				{Key: "name", Value: "Test Station"},
				{Key: "railway", Value: "station"},
				{Key: "ele", Value: "150.0"},
				{Key: "ele:source", Value: "SRTM"},
			},
		},
		{
			name: "Override existing elevation",
			existingTags: []NodeTag{
				{Key: "name", Value: "Mountain Hut"},
				{Key: "ele", Value: "1000"},
			},
			newTags: map[string]string{
				"ele":        "1050.5",
				"ele:source": "SRTM",
			},
			want: []NodeTag{
				{Key: "name", Value: "Mountain Hut"},
				{Key: "ele", Value: "1050.5"},
				{Key: "ele:source", Value: "SRTM"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeTags(tt.existingTags, tt.newTags)
			
			// Convert to maps for easier comparison
			gotMap := make(map[string]string)
			for _, tag := range got {
				gotMap[tag.Key] = tag.Value
			}
			
			wantMap := make(map[string]string)
			for _, tag := range tt.want {
				wantMap[tag.Key] = tag.Value
			}
			
			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("MergeTags() got = %v, want %v", gotMap, wantMap)
			}
		})
	}
}

func TestOAuthConfigSaveLoad(t *testing.T) {
	// Create a temporary .env file using t.TempDir()
	tmpDir := t.TempDir()
	tmpEnv := tmpDir + "/test_elevate_romania.env"

	// Test saving
	config := &OAuthConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		AccessToken:  "test_access_token",
	}

	// Save to custom location
	content := fmt.Sprintf("OSM_CLIENT_ID=%s\nOSM_CLIENT_SECRET=%s\nOSM_ACCESS_TOKEN=%s\n",
		config.ClientID, config.ClientSecret, config.AccessToken)
	
	if err := os.WriteFile(tmpEnv, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test .env: %v", err)
	}

	// Read back
	data, err := os.ReadFile(tmpEnv)
	if err != nil {
		t.Fatalf("Failed to read test .env: %v", err)
	}

	// Verify content contains our values
	content = string(data)
	if !strings.Contains(content, "OSM_CLIENT_ID=test_client_id") {
		t.Error("OSM_CLIENT_ID not found in saved file")
	}
	if !strings.Contains(content, "OSM_CLIENT_SECRET=test_client_secret") {
		t.Error("OSM_CLIENT_SECRET not found in saved file")
	}
	if !strings.Contains(content, "OSM_ACCESS_TOKEN=test_access_token") {
		t.Error("OSM_ACCESS_TOKEN not found in saved file")
	}
}
