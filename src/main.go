package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Define command-line flags
	extract := flag.Bool("extract", false, "Extract data from OSM")
	filter := flag.Bool("filter", false, "Filter elements without elevation")
	enrich := flag.Bool("enrich", false, "Enrich with elevation data")
	validate := flag.Bool("validate", false, "Validate elevation ranges")
	exportCSV := flag.Bool("export-csv", false, "Export to CSV")
	upload := flag.Bool("upload", false, "Upload to OSM")
	all := flag.Bool("all", false, "Run all steps")
	dryRun := flag.Bool("dry-run", false, "Dry-run mode (don't upload)")
	limit := flag.Int("limit", 0, "Limit number of items to process (for testing)")
	oauthInteractive := flag.Bool("oauth-interactive", false, "Interactive OAuth setup")
	country := flag.String("country", "România", "Country name to target (int_name from OSM)")
	listCountries := flag.Bool("list-countries", false, "List all available admin_level=2 countries")
	processAllCountries := flag.Bool("process-all-countries", false, "Process all available countries sequentially")

	flag.Parse()

	// Handle list-countries flag
	if *listCountries {
		if err := runListCountries(); err != nil {
			log.Fatalf("List countries failed: %v", err)
		}
		return
	}

	// Handle process-all-countries flag
	if *processAllCountries {
		if err := runProcessAllCountries(*limit, *dryRun, *oauthInteractive); err != nil {
			log.Fatalf("Process all countries failed: %v", err)
		}
		return
	}

	// Check if any action is specified
	if !(*extract || *filter || *enrich || *validate || *exportCSV || *upload || *all) {
		flag.Usage()
		fmt.Println("\nExamples:")
		fmt.Println("  elevate-romania --all --dry-run")
		fmt.Println("  elevate-romania --extract --filter")
		fmt.Println("  elevate-romania --enrich --limit 10")
		fmt.Println("  elevate-romania --upload --dry-run")
		fmt.Println("  elevate-romania --upload --oauth-interactive")
		fmt.Println("  elevate-romania --country \"Moldova\" --extract")
		fmt.Println("  elevate-romania --list-countries")
		fmt.Println("  elevate-romania --process-all-countries --limit 2000 --dry-run")
		return
	}

	fmt.Println("=" + string(repeat('=', 60)))
	fmt.Println("ELEVAȚIE OSM")
	fmt.Printf("Adding elevation to train stations and accommodations in %s\n", *country)
	fmt.Printf("Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=" + string(repeat('=', 60)))

	// Create output directory
	if err := os.MkdirAll("output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Run steps
	if *all || *extract {
		if err := runExtract(*country); err != nil {
			log.Fatalf("Extract failed: %v", err)
		}
	}

	if *all || *filter {
		if err := runFilter(); err != nil {
			log.Fatalf("Filter failed: %v", err)
		}
	}

	if *all || *enrich {
		if err := runEnrich(*limit); err != nil {
			log.Fatalf("Enrich failed: %v", err)
		}
	}

	if *all || *validate {
		if err := runValidate(); err != nil {
			log.Fatalf("Validate failed: %v", err)
		}
	}

	if *all || *exportCSV {
		if err := runExportCSV(); err != nil {
			log.Fatalf("Export CSV failed: %v", err)
		}
	}

	if *all || *upload {
		// Handle OAuth credentials
		var oauthConfig *OAuthConfig
		var err error

		if *oauthInteractive {
			oauthConfig, err = InteractiveOAuthSetup()
			if err != nil {
				log.Fatalf("OAuth setup failed: %v", err)
			}
		} else {
			oauthConfig, err = LoadOAuthConfig()
			if err != nil {
				log.Fatalf("Failed to load OAuth config: %v", err)
			}
		}

		isDryRun := *dryRun
		if !isDryRun && (oauthConfig.ClientID == "" || oauthConfig.ClientSecret == "" || oauthConfig.AccessToken == "") {
			fmt.Println("\nWarning: OAuth credentials not provided, running in dry-run mode")
			fmt.Println("Use --oauth-interactive for setup or set OSM_CLIENT_ID, OSM_CLIENT_SECRET, OSM_ACCESS_TOKEN in .env")
			isDryRun = true
		}

		if err := runUpload(isDryRun, oauthConfig, *country); err != nil {
			log.Fatalf("Upload failed: %v", err)
		}
	}

	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("COMPLETED SUCCESSFULLY!")
	fmt.Printf("Finished: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(string(repeat('=', 60)) + "\n")
}

func repeat(char rune, count int) []rune {
	result := make([]rune, count)
	for i := range result {
		result[i] = char
	}
	return result
}

// runProcessAllCountries fetches all countries and processes each one with the full pipeline
func runProcessAllCountries(limit int, dryRun bool, oauthInteractive bool) error {
	fmt.Println("\n" + string(repeat('=', 60)))
	fmt.Println("GLOBAL PROCESSING - Processing all countries")
	fmt.Println(string(repeat('=', 60)))
	fmt.Printf("Limit per country: %d\n", limit)
	fmt.Printf("Dry-run mode: %v\n", dryRun)
	fmt.Printf("Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(string(repeat('=', 60)))

	// Fetch all countries
	fmt.Println("\nFetching list of all countries...")
	countries, err := fetchAllCountries()
	if err != nil {
		return fmt.Errorf("failed to fetch countries: %v", err)
	}

	fmt.Printf("\nFound %d countries to process\n", len(countries))
	
	// Track statistics
	successCount := 0
	failedCountries := []string{}
	
	// Process each country
	for i, country := range countries {
		countryName := country.Name
		fmt.Println("\n" + string(repeat('=', 60)))
		fmt.Printf("Processing country %d/%d: %s\n", i+1, len(countries), countryName)
		fmt.Println(string(repeat('=', 60)))
		
		// Process this country
		if err := processCountry(countryName, limit, dryRun, oauthInteractive); err != nil {
			log.Printf("ERROR: Failed to process %s: %v\n", countryName, err)
			failedCountries = append(failedCountries, countryName)
			// Continue with next country instead of stopping
			continue
		}
		
		successCount++
		
		// Add delay between countries to be nice to APIs
		if i < len(countries)-1 {
			fmt.Println("\nWaiting 5 seconds before processing next country...")
			time.Sleep(5 * time.Second)
		}
	}
	
	// Print summary
	fmt.Println("\n" + string(repeat('=', 80)))
	fmt.Println("GLOBAL PROCESSING SUMMARY")
	fmt.Println(string(repeat('=', 80)))
	fmt.Printf("Total countries: %d\n", len(countries))
	fmt.Printf("Successfully processed: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(failedCountries))
	
	if len(failedCountries) > 0 {
		fmt.Println("\nFailed countries:")
		for _, c := range failedCountries {
			fmt.Printf("  - %s\n", c)
		}
	}
	
	fmt.Printf("\nCompleted: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(string(repeat('=', 80)) + "\n")
	
	return nil
}

// processCountry runs the full pipeline for a single country
func processCountry(country string, limit int, dryRun bool, oauthInteractive bool) error {
	// Create output directory
	if err := os.MkdirAll("output", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Step 1: Extract
	fmt.Println("\nStep 1: Extract")
	if err := runExtract(country); err != nil {
		return fmt.Errorf("extract failed: %v", err)
	}

	// Step 2: Filter
	fmt.Println("\nStep 2: Filter")
	if err := runFilter(); err != nil {
		return fmt.Errorf("filter failed: %v", err)
	}

	// Step 3: Enrich
	fmt.Println("\nStep 3: Enrich")
	if err := runEnrich(limit); err != nil {
		return fmt.Errorf("enrich failed: %v", err)
	}

	// Step 4: Validate
	fmt.Println("\nStep 4: Validate")
	if err := runValidate(); err != nil {
		return fmt.Errorf("validate failed: %v", err)
	}

	// Step 5: Export CSV
	fmt.Println("\nStep 5: Export CSV")
	if err := runExportCSV(); err != nil {
		return fmt.Errorf("export CSV failed: %v", err)
	}

	// Step 6: Upload (only if not dry-run)
	fmt.Println("\nStep 6: Upload")
	var oauthConfig *OAuthConfig
	var err error

	if oauthInteractive {
		oauthConfig, err = InteractiveOAuthSetup()
		if err != nil {
			return fmt.Errorf("OAuth setup failed: %v", err)
		}
	} else {
		oauthConfig, err = LoadOAuthConfig()
		if err != nil {
			return fmt.Errorf("failed to load OAuth config: %v", err)
		}
	}

	isDryRun := dryRun
	if !isDryRun && (oauthConfig.ClientID == "" || oauthConfig.ClientSecret == "" || oauthConfig.AccessToken == "") {
		fmt.Println("\nWarning: OAuth credentials not provided, running in dry-run mode")
		isDryRun = true
	}

	if err := runUpload(isDryRun, oauthConfig, country); err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	return nil
}
