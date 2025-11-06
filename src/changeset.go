package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// ChangesetManager handles OSM changeset operations
type ChangesetManager struct {
	client      *http.Client
	changesetID int
	dryRun      bool
}

// OSMChangeset represents the changeset XML structure
type OSMChangeset struct {
	XMLName   xml.Name      `xml:"osm"`
	Changeset ChangesetData `xml:"changeset"`
}

// ChangesetData contains changeset information
type ChangesetData struct {
	Tags []ChangesetTag `xml:"tag"`
}

// ChangesetTag represents a tag in the changeset
type ChangesetTag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

// NewChangesetManager creates a new changeset manager
func NewChangesetManager(client *http.Client, dryRun bool) *ChangesetManager {
	return &ChangesetManager{
		client: client,
		dryRun: dryRun,
	}
}

// Create creates a new changeset
func (cm *ChangesetManager) Create(comment string) error {
	if cm.dryRun {
		fmt.Printf("[DRY-RUN] Would create changeset: %s\n", comment)
		return nil
	}

	changesetXML := OSMChangeset{
		Changeset: ChangesetData{
			Tags: []ChangesetTag{
				{Key: "created_by", Value: "elevate-romania"},
				{Key: "comment", Value: comment},
			},
		},
	}

	xmlData, err := xml.Marshal(changesetXML)
	if err != nil {
		return fmt.Errorf("failed to marshal changeset XML: %v", err)
	}

	req, err := http.NewRequest("PUT", "https://api.openstreetmap.org/api/0.6/changeset/create", bytes.NewReader(xmlData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := cm.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create changeset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create changeset: status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Sscanf(string(body), "%d", &cm.changesetID)
	fmt.Printf("Created changeset #%d\n", cm.changesetID)

	return nil
}

// Close closes the changeset
func (cm *ChangesetManager) Close() error {
	if cm.dryRun || cm.changesetID == 0 {
		return nil
	}

	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/changeset/%d/close", cm.changesetID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := cm.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to close changeset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to close changeset: status code %d", resp.StatusCode)
	}

	fmt.Printf("Closed changeset #%d\n", cm.changesetID)
	return nil
}

// GetID returns the current changeset ID
func (cm *ChangesetManager) GetID() int {
	return cm.changesetID
}
