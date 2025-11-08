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

// clusterProcessor handles processing of a single cluster
type clusterProcessor struct {
	uploader   *OSMUploader
	categorizer *ElementCategorizer
}

// newClusterProcessor creates a new cluster processor
func newClusterProcessor(uploader *OSMUploader) *clusterProcessor {
	return &clusterProcessor{
		uploader:    uploader,
		categorizer: NewElementCategorizer(),
	}
}

// categorizeElements splits elements into categories
func (cp *clusterProcessor) categorizeElements(elements []OSMElement) (alpineHuts, trainStations, otherAccommodations []OSMElement) {
	for _, element := range elements {
		category := cp.categorizer.Categorize(element)
		switch category {
		case CategoryAlpineHut:
			alpineHuts = append(alpineHuts, element)
		case CategoryTrainStation:
			trainStations = append(trainStations, element)
		case CategoryOtherAccommodation:
			otherAccommodations = append(otherAccommodations, element)
		}
	}
	return
}

// processCluster processes a single cluster with its own changeset
func (cp *clusterProcessor) processCluster(cluster ElementCluster, clusterNum, totalClusters int, categoryStats map[string]*UploadStats) error {
	clusterSize := len(cluster.Elements)
	
	// Print cluster header
	cp.printClusterHeader(clusterNum, totalClusters, clusterSize, cluster.BBox)

	// Categorize elements
	alpineHuts, trainStations, otherAccommodations := cp.categorizeElements(cluster.Elements)

	// Create changeset for this cluster
	changesetComment := fmt.Sprintf("Add elevation data to %d locations in %s - cluster %d/%d (alpine huts, train stations, accommodations)",
		clusterSize, cp.uploader.country, clusterNum, totalClusters)
	
	if err := cp.uploader.CreateChangeset(changesetComment); err != nil {
		cp.handleChangesetCreationError(cluster.Elements, err, categoryStats)
		return err
	}

	// Upload elements by category
	cp.uploadCategoryElements(alpineHuts, "alpine_huts", clusterNum, categoryStats)
	cp.uploadCategoryElements(trainStations, "train_stations", clusterNum, categoryStats)
	cp.uploadCategoryElements(otherAccommodations, "other_accommodations", clusterNum, categoryStats)

	// Close changeset
	if err := cp.uploader.CloseChangeset(); err != nil {
		fmt.Printf("WARNING: Failed to close changeset for cluster %d: %v\n", clusterNum, err)
	}

	// Rate limiting delay
	if clusterNum < totalClusters && !cp.uploader.dryRun {
		fmt.Printf("\nWaiting 2 seconds before next cluster...\n")
		time.Sleep(2 * time.Second)
	}

	return nil
}

// printClusterHeader prints the cluster processing header
func (cp *clusterProcessor) printClusterHeader(clusterNum, totalClusters, clusterSize int, bbox BoundingBox) {
	fmt.Printf("\n%s\n", string(repeat('=', 60)))
	fmt.Printf("Processing cluster %d/%d (%d elements)\n", clusterNum, totalClusters, clusterSize)
	fmt.Printf("Bounding box: [%.4f,%.4f] to [%.4f,%.4f] (diagonal: %.4f°)\n",
		bbox.MinLat, bbox.MinLon,
		bbox.MaxLat, bbox.MaxLon,
		bbox.Diagonal())
	fmt.Printf("%s\n", string(repeat('=', 60)))
}

// handleChangesetCreationError handles errors when creating a changeset
func (cp *clusterProcessor) handleChangesetCreationError(elements []OSMElement, err error, categoryStats map[string]*UploadStats) {
	fmt.Printf("WARNING: Failed to create changeset: %v\n", err)
	
	// Mark all elements in this cluster as failed
	for _, elem := range elements {
		category := cp.categorizer.Categorize(elem)
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
}

// uploadCategoryElements uploads elements of a specific category
func (cp *clusterProcessor) uploadCategoryElements(elements []OSMElement, categoryKey string, clusterNum int, categoryStats map[string]*UploadStats) {
	if len(elements) == 0 {
		return
	}
	
	stats := cp.uploader.UploadElements(elements, fmt.Sprintf("%s (cluster %d)", categoryKey, clusterNum))
	categoryStats[categoryKey].Total += stats.Total
	categoryStats[categoryKey].Successful += stats.Successful
	categoryStats[categoryKey].Failed += stats.Failed
	categoryStats[categoryKey].Errors = append(categoryStats[categoryKey].Errors, stats.Errors...)
}

// initializeCategoryStats creates the initial stats structure
func initializeCategoryStats() map[string]*UploadStats {
	return map[string]*UploadStats{
		"alpine_huts":          {Total: 0, Successful: 0, Failed: 0, Errors: []UploadError{}},
		"train_stations":       {Total: 0, Successful: 0, Failed: 0, Errors: []UploadError{}},
		"other_accommodations": {Total: 0, Successful: 0, Failed: 0, Errors: []UploadError{}},
	}
}

// collectAllElements gathers all elements from validated data
func collectAllElements(data ValidatedData) []OSMElement {
	allElements := make([]OSMElement, 0)
	allElements = append(allElements, data.AlpineHuts.ValidElements...)
	allElements = append(allElements, data.TrainStations.ValidElements...)
	allElements = append(allElements, data.OtherAccommodations.ValidElements...)
	return allElements
}

// printClusteringSummary prints information about the clustering
func printClusteringSummary(totalElements int, clusters []ElementCluster) {
	fmt.Printf("\nGrouping %d elements by geographic proximity...\n", totalElements)
	fmt.Printf("Created %d geographic clusters to avoid bounding box size limits\n", len(clusters))
	fmt.Printf("Each changeset will cover a maximum area of %.2f degrees diagonal\n\n", MaxBoundingBoxDiagonal)
}

func (u *OSMUploader) UploadAll(data ValidatedData) (map[string]UploadStats, error) {
	allStats := make(map[string]UploadStats)

	// Collect all elements
	allElements := collectAllElements(data)
	totalElements := len(allElements)
	
	if totalElements == 0 {
		return allStats, fmt.Errorf("no elements to upload")
	}

	// Cluster elements by geographic proximity
	clusters := ClusterElements(allElements, MaxBoundingBoxDiagonal)
	printClusteringSummary(totalElements, clusters)

	// Initialize stats tracking
	categoryStats := initializeCategoryStats()

	// Process each cluster
	processor := newClusterProcessor(u)
	for clusterIdx, cluster := range clusters {
		processor.processCluster(cluster, clusterIdx+1, len(clusters), categoryStats)
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
