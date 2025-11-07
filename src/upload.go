package main

import (
	"fmt"
	"net/http"
	"time"
)

// OSMUploader handles uploading changes to OpenStreetMap
type OSMUploader struct {
	client           *http.Client
	changesetManager *ChangesetManager
	apiClient        *OSMAPIClient
	dryRun           bool
	country          string
}

// UploadStats contains statistics about uploads
type UploadStats struct {
	Total      int           `json:"total"`
	Successful int           `json:"successful"`
	Failed     int           `json:"failed"`
	Errors     []UploadError `json:"errors"`
}

// UploadError represents an error during upload
type UploadError struct {
	ElementType string `json:"element_type"`
	ElementID   int64  `json:"element_id"`
	Error       string `json:"error"`
}

// NewOSMUploader creates a new OSM uploader
func NewOSMUploader(oauthConfig *OAuthConfig, dryRun bool, country string) (*OSMUploader, error) {
	uploader := &OSMUploader{
		dryRun:  dryRun,
		country: country,
	}

	if dryRun {
		fmt.Println("Running in DRY-RUN mode - no changes will be uploaded")
		uploader.changesetManager = NewChangesetManager(nil, true)
		uploader.apiClient = NewOSMAPIClient(nil, true)
		return uploader, nil
	}

	if oauthConfig.AccessToken == "" {
		return nil, fmt.Errorf("OAuth access token required for actual upload")
	}

	// Create OAuth HTTP client
	_, client, err := CreateOAuthClient(oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth client: %v", err)
	}

	uploader.client = client
	uploader.changesetManager = NewChangesetManager(client, false)
	uploader.apiClient = NewOSMAPIClient(client, false)

	fmt.Println("Connected to OSM API with OAuth 2.0")

	return uploader, nil
}

// CreateChangeset creates a new changeset
func (u *OSMUploader) CreateChangeset(comment string) error {
	return u.changesetManager.Create(comment)
}

// CloseChangeset closes the current changeset
func (u *OSMUploader) CloseChangeset() error {
	return u.changesetManager.Close()
}

// UploadElement uploads a single element to OSM
func (u *OSMUploader) UploadElement(element OSMElement) (bool, string) {
	elementType := element.Type
	elementID := element.ID
	tags := element.Tags

	if tags == nil || tags["ele"] == "" || tags["ele:source"] == "" {
		return false, "Missing elevation data in tags"
	}

	eleValue := tags["ele"]

	if u.dryRun {
		fmt.Printf("[DRY-RUN] Would update %s %d:\n", elementType, elementID)
		fmt.Printf("  ele=%s, ele:source=SRTM\n", eleValue)
		return true, "Dry-run successful"
	}

	// Get changeset ID
	if !u.changesetManager.IsOpen() {
		return false, "No active changeset"
	}
	changesetID := u.changesetManager.GetID()

	// Prepare new tags to merge
	newTags := map[string]string{
		"ele":        eleValue,
		"ele:source": "SRTM",
	}

	// Fetch current element and update it
	var err error
	if elementType == "node" {
		err = u.uploadNode(elementID, newTags, changesetID)
	} else if elementType == "way" {
		err = u.uploadWay(elementID, newTags, changesetID)
	} else {
		return false, fmt.Sprintf("Unsupported element type: %s", elementType)
	}

	if err != nil {
		return false, fmt.Sprintf("Upload failed: %v", err)
	}

	fmt.Printf("✓ Updated %s %d with ele=%s\n", elementType, elementID, eleValue)
	return true, "Upload successful"
}

// uploadNode fetches and updates a node
func (u *OSMUploader) uploadNode(nodeID int64, newTags map[string]string, changesetID int) error {
	// Fetch current node
	node, err := u.apiClient.FetchNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to fetch node: %v", err)
	}

	// Merge tags
	node.Tags = MergeTags(node.Tags, newTags)

	// Update node
	if err := u.apiClient.UpdateNode(node, changesetID); err != nil {
		return fmt.Errorf("failed to update node: %v", err)
	}

	return nil
}

// uploadWay fetches and updates a way
func (u *OSMUploader) uploadWay(wayID int64, newTags map[string]string, changesetID int) error {
	// Fetch current way
	way, err := u.apiClient.FetchWay(wayID)
	if err != nil {
		return fmt.Errorf("failed to fetch way: %v", err)
	}

	// Merge tags
	way.Tags = MergeTags(way.Tags, newTags)

	// Update way
	if err := u.apiClient.UpdateWay(way, changesetID); err != nil {
		return fmt.Errorf("failed to update way: %v", err)
	}

	return nil
}

func (u *OSMUploader) UploadElements(elements []OSMElement, categoryName string) UploadStats {
	stats := UploadStats{
		Total:      len(elements),
		Successful: 0,
		Failed:     0,
		Errors:     []UploadError{},
	}

	if len(elements) == 0 {
		return stats
	}

	fmt.Printf("\nUploading %s...\n", categoryName)

	for i, element := range elements {
		success, message := u.UploadElement(element)

		if success {
			stats.Successful++
		} else {
			stats.Failed++
			stats.Errors = append(stats.Errors, UploadError{
				ElementType: element.Type,
				ElementID:   element.ID,
				Error:       message,
			})
		}

		// Progress update
		if (i+1)%10 == 0 {
			fmt.Printf("Progress: %d/%d\n", i+1, len(elements))
		}

		// Rate limiting
		if !u.dryRun {
			time.Sleep(time.Millisecond * 10)
		}
	}

	return stats
}

func (u *OSMUploader) UploadAll(data ValidatedData) (map[string]UploadStats, error) {
	allStats := make(map[string]UploadStats)

	// Calculate total elements
	totalElements := len(data.TrainStations.ValidElements) +
		len(data.AlpineHuts.ValidElements) +
		len(data.OtherAccommodations.ValidElements)

	if totalElements == 0 {
		return allStats, fmt.Errorf("no elements to upload")
	}

	// Create changeset
	changesetComment := fmt.Sprintf("Add elevation data to %d locations in %s (alpine huts, train stations, accommodations)", totalElements, u.country)
	if err := u.CreateChangeset(changesetComment); err != nil {
		return allStats, fmt.Errorf("failed to create changeset: %v", err)
	}

	// Upload by category
	allStats["alpine_huts"] = u.UploadElements(data.AlpineHuts.ValidElements, "alpine_huts")
	allStats["train_stations"] = u.UploadElements(data.TrainStations.ValidElements, "train_stations")
	allStats["other_accommodations"] = u.UploadElements(data.OtherAccommodations.ValidElements, "other_accommodations")

	// Close changeset
	if err := u.CloseChangeset(); err != nil {
		return allStats, fmt.Errorf("failed to close changeset: %v", err)
	}

	return allStats, nil
}

// runUpload runs the upload process
func runUpload(dryRun bool, oauthConfig *OAuthConfig, country string) error {
	fmt.Println("\n" + string(repeat('=', 60)))
	if dryRun {
		fmt.Println("STEP 6: UPLOAD (DRY-RUN) - Preview changes")
	} else {
		fmt.Println("STEP 6: UPLOAD - Uploading to OpenStreetMap")
	}
	fmt.Println(string(repeat('=', 60)))

	// Load validated data
	var data ValidatedData
	if err := loadJSON("output/osm_data_validated.json", &data); err != nil {
		return fmt.Errorf("output/osm_data_validated.json not found. Run --validate first: %v", err)
	}

	// Upload
	uploader, err := NewOSMUploader(oauthConfig, dryRun, country)
	if err != nil {
		return err
	}

	stats, err := uploader.UploadAll(data)
	if err != nil {
		return err
	}

	// Display statistics
	fmt.Println("\n" + string(repeat('=', 60)))
	if dryRun {
		fmt.Println("UPLOAD STATISTICS (DRY-RUN)")
	} else {
		fmt.Println("UPLOAD STATISTICS")
	}
	fmt.Println(string(repeat('=', 60)))

	for category, categoryStats := range stats {
		fmt.Printf("\n%s:\n", category)
		fmt.Printf("  Total: %d\n", categoryStats.Total)
		fmt.Printf("  Successful: %d\n", categoryStats.Successful)
		fmt.Printf("  Failed: %d\n", categoryStats.Failed)

		if categoryStats.Failed > 0 && len(categoryStats.Errors) > 0 {
			fmt.Println("  First errors:")
			for i, err := range categoryStats.Errors {
				if i >= 3 {
					break
				}
				fmt.Printf("    - %s %d: %s\n", err.ElementType, err.ElementID, err.Error)
			}
		}
	}

	fmt.Println("\n" + string(repeat('=', 60)) + "\n")

	return nil
}
