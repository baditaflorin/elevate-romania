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

	flag.Parse()

	// Check if any action is specified
	if !(*extract || *filter || *enrich || *validate || *exportCSV || *upload || *all) {
		flag.Usage()
		fmt.Println("\nExamples:")
		fmt.Println("  elevate-romania --all --dry-run")
		fmt.Println("  elevate-romania --extract --filter")
		fmt.Println("  elevate-romania --enrich --limit 10")
		fmt.Println("  elevate-romania --upload --dry-run")
		fmt.Println("  elevate-romania --upload --oauth-interactive")
		return
	}

	fmt.Println("=" + string(repeat('=', 60)))
	fmt.Println("ELEVAȚIE OSM ROMÂNIA")
	fmt.Println("Adding elevation to train stations and accommodations")
	fmt.Printf("Started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=" + string(repeat('=', 60)))

	// Create output directory
	if err := os.MkdirAll("output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Run steps
	if *all || *extract {
		if err := runExtract(); err != nil {
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
		var clientID, clientSecret, accessToken string

		if *oauthInteractive {
			var err error
			clientID, clientSecret, accessToken, err = getOAuthCredentials()
			if err != nil {
				log.Fatalf("OAuth setup failed: %v", err)
			}
		} else {
			clientID = os.Getenv("CLIENT_ID")
			clientSecret = os.Getenv("CLIENT_SECRET")
			accessToken = os.Getenv("ACCESS_TOKEN")
		}

		isDryRun := *dryRun
		if !isDryRun && (clientID == "" || clientSecret == "" || accessToken == "") {
			fmt.Println("\nWarning: OAuth credentials not provided, running in dry-run mode")
			fmt.Println("Use --oauth-interactive for setup or set CLIENT_ID, CLIENT_SECRET, ACCESS_TOKEN")
			isDryRun = true
		}

		if err := runUpload(isDryRun, clientID, clientSecret, accessToken); err != nil {
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
