package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	redirectURI = "http://127.0.0.1:8080/callback"
)

// OAuthConfig holds OAuth 2.0 configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
}

// LoadOAuthConfig loads OAuth configuration from environment variables or .env file
func LoadOAuthConfig() (*OAuthConfig, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &OAuthConfig{
		ClientID:     os.Getenv("OSM_CLIENT_ID"),
		ClientSecret: os.Getenv("OSM_CLIENT_SECRET"),
		AccessToken:  os.Getenv("OSM_ACCESS_TOKEN"),
	}

	return config, nil
}

// SaveOAuthConfig saves OAuth configuration to .env file
func SaveOAuthConfig(config *OAuthConfig) error {
	envFile := ".env"
	
	// Read existing .env if present
	existingEnv := make(map[string]string)
	if data, err := os.ReadFile(envFile); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existingEnv[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Update OAuth values
	existingEnv["OSM_CLIENT_ID"] = config.ClientID
	existingEnv["OSM_CLIENT_SECRET"] = config.ClientSecret
	existingEnv["OSM_ACCESS_TOKEN"] = config.AccessToken

	// Write back to file
	var content strings.Builder
	content.WriteString("# OpenStreetMap OAuth 2.0 Credentials\n")
	content.WriteString(fmt.Sprintf("OSM_CLIENT_ID=%s\n", existingEnv["OSM_CLIENT_ID"]))
	content.WriteString(fmt.Sprintf("OSM_CLIENT_SECRET=%s\n", existingEnv["OSM_CLIENT_SECRET"]))
	content.WriteString(fmt.Sprintf("OSM_ACCESS_TOKEN=%s\n", existingEnv["OSM_ACCESS_TOKEN"]))
	
	// Add other existing env vars that aren't OAuth-related
	for key, value := range existingEnv {
		if !strings.HasPrefix(key, "OSM_") {
			content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}

	return os.WriteFile(envFile, []byte(content.String()), 0600)
}

// InteractiveOAuthSetup performs interactive OAuth setup
func InteractiveOAuthSetup() (*OAuthConfig, error) {
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

	// Start OAuth flow
	accessToken, err := startOAuthFlow(clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	config := &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
	}

	// Save to .env file
	if err := SaveOAuthConfig(config); err != nil {
		fmt.Printf("Warning: Failed to save credentials to .env: %v\n", err)
	} else {
		fmt.Println("✓ Credentials saved to .env file")
	}

	fmt.Println("✓ Access token obtained successfully!")

	return config, nil
}

// startOAuthFlow performs the OAuth 2.0 authorization flow
func startOAuthFlow(clientID, clientSecret string) (string, error) {
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

// exchangeCodeForToken exchanges authorization code for access token
func exchangeCodeForToken(clientID, clientSecret, code string) (string, error) {
	oauth2Config := &oauth2.Config{
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

// CreateOAuthClient creates an authenticated HTTP client
func CreateOAuthClient(config *OAuthConfig) (*oauth2.Config, *http.Client, error) {
	if config.AccessToken == "" {
		return nil, nil, fmt.Errorf("OAuth access token required")
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
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

	token := &oauth2.Token{
		AccessToken: config.AccessToken,
		TokenType:   "Bearer",
	}

	ctx := context.Background()
	client := oauth2Cfg.Client(ctx, token)

	return oauth2Cfg, client, nil
}
