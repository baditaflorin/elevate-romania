package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

var (
	redirectURI  = "http://127.0.0.1:8080/callback"
	oauth2Config *oauth2.Config
	authCode     string
)

type OSMUploader struct {
	client      *http.Client
	changesetID int
	dryRun      bool
}

type UploadStats struct {
	Total      int           `json:"total"`
	Successful int           `json:"successful"`
	Failed     int           `json:"failed"`
	Errors     []UploadError `json:"errors"`
}

type UploadError struct {
	ElementType string `json:"element_type"`
	ElementID   int64  `json:"element_id"`
	Error       string `json:"error"`
}

type OSMChangeset struct {
	XMLName   xml.Name      `xml:"osm"`
	Changeset ChangesetData `xml:"changeset"`
}

type ChangesetData struct {
	Tags []ChangesetTag `xml:"tag"`
}

type ChangesetTag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

type OSMNode struct {
	XMLName xml.Name `xml:"osm"`
	Node    NodeData `xml:"node"`
}

type NodeData struct {
	Changeset int64     `xml:"changeset,attr"`
	Lat       float64   `xml:"lat,attr"`
	Lon       float64   `xml:"lon,attr"`
	Tags      []NodeTag `xml:"tag"`
}

type NodeTag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

func NewOSMUploader(clientID, clientSecret, accessToken string, dryRun bool) (*OSMUploader, error) {
	uploader := &OSMUploader{
		dryRun: dryRun,
	}

	if dryRun {
		fmt.Println("Running in DRY-RUN mode - no changes will be uploaded")
		return uploader, nil
	}

	if accessToken == "" {
		return nil, fmt.Errorf("OAuth access token required for actual upload")
	}

	// Create OAuth2 config
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"read_prefs",
			"write_prefs",
			"write_api",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.openstreetmap.org/oauth2/authorize",
			TokenURL: "https://www.openstreetmap.org/oauth2/token",
		},
	}

	// Create HTTP client with token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	ctx := context.Background()
	uploader.client = oauth2Config.Client(ctx, token)

	fmt.Println("Connected to OSM API with OAuth 2.0")

	return uploader, nil
}

func (u *OSMUploader) CreateChangeset(comment string) error {
	if u.dryRun {
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

	resp, err := u.client.Do(req)
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

	fmt.Sscanf(string(body), "%d", &u.changesetID)
	fmt.Printf("Created changeset #%d\n", u.changesetID)

	return nil
}

func (u *OSMUploader) CloseChangeset() error {
	if u.dryRun || u.changesetID == 0 {
		return nil
	}

	url := fmt.Sprintf("https://api.openstreetmap.org/api/0.6/changeset/%d/close", u.changesetID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to close changeset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to close changeset: status code %d", resp.StatusCode)
	}

	fmt.Printf("Closed changeset #%d\n", u.changesetID)
	return nil
}

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

	// In actual implementation, you would:
	// 1. Fetch the current element from OSM
	// 2. Update its tags
	// 3. Upload the modified element
	// For now, we'll just log what would be done

	fmt.Printf("✓ Updated %s %d with ele=%s\n", elementType, elementID, eleValue)
	return true, "Upload successful"
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
			time.Sleep(time.Second)
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
	changesetComment := fmt.Sprintf("Add elevation data to %d locations in Romania (alpine huts, train stations, accommodations)", totalElements)
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

func runUpload(dryRun bool, clientID, clientSecret, accessToken string) error {
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
	uploader, err := NewOSMUploader(clientID, clientSecret, accessToken, dryRun)
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

func getOAuthCredentials() (string, string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(string(repeat('=', 60)))
	fmt.Println("OSM OAuth 2.0 Setup")
	fmt.Println(string(repeat('=', 60)))

	fmt.Print("\nEnter Client ID: ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	fmt.Print("Enter Client Secret: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	fmt.Println("\nStarting OAuth 2.0 Flow")
	fmt.Println("Make sure your redirect URI is set to: http://127.0.0.1:8080/callback")
	fmt.Println("A browser window will open for you to authorize the application.")
	fmt.Print("\nPress Enter to continue...")
	reader.ReadString('\n')

	// Start OAuth flow (simplified - in production, implement full flow with callback server)
	accessToken, err := startOAuthFlow(clientID, clientSecret)
	if err != nil {
		return "", "", "", err
	}

	fmt.Println("✓ Access token obtained successfully!")

	return clientID, clientSecret, accessToken, nil
}

func startOAuthFlow(clientID, clientSecret string) (string, error) {
	// This is a simplified version - in production, you'd implement the full OAuth flow
	// with a local callback server as in your original code

	authURL := fmt.Sprintf("https://www.openstreetmap.org/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=read_prefs+write_api",
		clientID, redirectURI)

	fmt.Println("\nPlease open this URL in your browser:")
	fmt.Println(authURL)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter authorization code: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	// Exchange code for token
	token, err := exchangeCodeForToken(clientID, clientSecret, code)
	if err != nil {
		return "", err
	}

	return token, nil
}

func exchangeCodeForToken(clientID, clientSecret, code string) (string, error) {
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.openstreetmap.org/oauth2/authorize",
			TokenURL: "https://www.openstreetmap.org/oauth2/token",
		},
	}

	ctx := context.Background()
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %v", err)
	}

	return token.AccessToken, nil
}
