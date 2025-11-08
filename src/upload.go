package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	// MaxBoundingBoxDiagonal is the maximum diagonal distance (in degrees) for a changeset
	// OSM typically allows up to 0.5 degrees, but we use 0.25 to be conservative
	// 0.25 degrees is approximately 28km at the equator
	MaxBoundingBoxDiagonal = 0.25
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

	// Collect all elements
	allElements := make([]OSMElement, 0)
	allElements = append(allElements, data.AlpineHuts.ValidElements...)
	allElements = append(allElements, data.TrainStations.ValidElements...)
	allElements = append(allElements, data.OtherAccommodations.ValidElements...)

	totalElements := len(allElements)
	if totalElements == 0 {
		return allStats, fmt.Errorf("no elements to upload")
	}

	fmt.Printf("\nGrouping %d elements by geographic proximity...\n", totalElements)
	
	// Cluster elements by geographic proximity to avoid bounding box size limits
	clusters := ClusterElements(allElements, MaxBoundingBoxDiagonal)
	
	fmt.Printf("Created %d geographic clusters to avoid bounding box size limits\n", len(clusters))
	fmt.Printf("Each changeset will cover a maximum area of %.2f degrees diagonal\n\n", MaxBoundingBoxDiagonal)

	// Track stats across all clusters
	categoryStats := map[string]*UploadStats{
		"alpine_huts":          {Total: 0, Successful: 0, Failed: 0, Errors: []UploadError{}},
		"train_stations":       {Total: 0, Successful: 0, Failed: 0, Errors: []UploadError{}},
		"other_accommodations": {Total: 0, Successful: 0, Failed: 0, Errors: []UploadError{}},
	}

	// Process each cluster with its own changeset
	for clusterIdx, cluster := range clusters {
		clusterNum := clusterIdx + 1
		clusterSize := len(cluster.Elements)
		
		fmt.Printf("\n" + string(repeat('=', 60)) + "\n")
		fmt.Printf("Processing cluster %d/%d (%d elements)\n", clusterNum, len(clusters), clusterSize)
		fmt.Printf("Bounding box: [%.4f,%.4f] to [%.4f,%.4f] (diagonal: %.4f°)\n",
			cluster.BBox.MinLat, cluster.BBox.MinLon,
			cluster.BBox.MaxLat, cluster.BBox.MaxLon,
			cluster.BBox.Diagonal())
		fmt.Printf(string(repeat('=', 60)) + "\n")

		// Categorize elements in this cluster
		alpineHuts := []OSMElement{}
		trainStations := []OSMElement{}
		otherAccommodations := []OSMElement{}

		categorizer := NewElementCategorizer()
		for _, element := range cluster.Elements {
			category := categorizer.Categorize(element)
			switch category {
			case CategoryAlpineHut:
				alpineHuts = append(alpineHuts, element)
			case CategoryTrainStation:
				trainStations = append(trainStations, element)
			case CategoryOtherAccommodation:
				otherAccommodations = append(otherAccommodations, element)
			}
		}

		// Create changeset for this cluster
		changesetComment := fmt.Sprintf("Add elevation data to %d locations in %s - cluster %d/%d (alpine huts, train stations, accommodations)",
			clusterSize, u.country, clusterNum, len(clusters))
		if err := u.CreateChangeset(changesetComment); err != nil {
			fmt.Printf("WARNING: Failed to create changeset for cluster %d: %v\n", clusterNum, err)
			// Mark all elements in this cluster as failed
			for _, elem := range cluster.Elements {
				category := categorizer.Categorize(elem)
				categoryKey := categoryToKey(category)
				if stats, ok := categoryStats[categoryKey]; ok {
					stats.Total++
					stats.Failed++
					stats.Errors = append(stats.Errors, UploadError{
						ElementType: elem.Type,
						ElementID:   elem.ID,
						Error:       fmt.Sprintf("Failed to create changeset: %v", err),
					})
				}
			}
			continue
		}

		// Upload elements in this cluster by category
		if len(alpineHuts) > 0 {
			stats := u.UploadElements(alpineHuts, fmt.Sprintf("alpine_huts (cluster %d)", clusterNum))
			categoryStats["alpine_huts"].Total += stats.Total
			categoryStats["alpine_huts"].Successful += stats.Successful
			categoryStats["alpine_huts"].Failed += stats.Failed
			categoryStats["alpine_huts"].Errors = append(categoryStats["alpine_huts"].Errors, stats.Errors...)
		}

		if len(trainStations) > 0 {
			stats := u.UploadElements(trainStations, fmt.Sprintf("train_stations (cluster %d)", clusterNum))
			categoryStats["train_stations"].Total += stats.Total
			categoryStats["train_stations"].Successful += stats.Successful
			categoryStats["train_stations"].Failed += stats.Failed
			categoryStats["train_stations"].Errors = append(categoryStats["train_stations"].Errors, stats.Errors...)
		}

		if len(otherAccommodations) > 0 {
			stats := u.UploadElements(otherAccommodations, fmt.Sprintf("other_accommodations (cluster %d)", clusterNum))
			categoryStats["other_accommodations"].Total += stats.Total
			categoryStats["other_accommodations"].Successful += stats.Successful
			categoryStats["other_accommodations"].Failed += stats.Failed
			categoryStats["other_accommodations"].Errors = append(categoryStats["other_accommodations"].Errors, stats.Errors...)
		}

		// Close changeset for this cluster
		if err := u.CloseChangeset(); err != nil {
			fmt.Printf("WARNING: Failed to close changeset for cluster %d: %v\n", clusterNum, err)
		}

		// Add delay between clusters to respect rate limits
		if clusterNum < len(clusters) && !u.dryRun {
			fmt.Printf("\nWaiting 2 seconds before next cluster...\n")
			time.Sleep(2 * time.Second)
		}
	}

	// Convert to final stats format
	for category, stats := range categoryStats {
		allStats[category] = *stats
	}

	return allStats, nil
}

// categoryToKey converts an ElementCategory to the string key used in stats maps
func categoryToKey(category ElementCategory) string {
	switch category {
	case CategoryAlpineHut:
		return "alpine_huts"
	case CategoryTrainStation:
		return "train_stations"
	case CategoryOtherAccommodation:
		return "other_accommodations"
	default:
		return "unknown"
	}
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
