# Getting Started with Elevație OSM România (Go)

This guide will help you set up and run the Go implementation of the OSM elevation enrichment tool.

## Quick Start

### 1. Prerequisites

Install Go 1.21 or higher:
- Download from: https://golang.org/dl/
- Verify installation: `go version`

### 2. Build the Application

```bash
# Make the build script executable
chmod +x build.sh

# Run the build script
./build.sh
```

Or manually:

```bash
# Download dependencies
go mod download

# Build the binary
go build -o elevate-romania .
```

### 3. Run Your First Test

```bash
# Test with dry-run mode (no actual uploads)
./elevate-romania --all --dry-run --limit 5
```

This will:
- Extract 5 items from OSM
- Filter those without elevation
- Fetch elevation data
- Validate the results
- Export to CSV
- Show what would be uploaded (but not actually upload)

## Step-by-Step Usage

### Step 1: Extract OSM Data

```bash
./elevate-romania --extract
```

This queries the Overpass API for:
- Train stations in Romania
- Accommodations (hotels, alpine huts, etc.)

Output: `output/osm_data_raw.json`

### Step 2: Filter Missing Elevation

```bash
./elevate-romania --filter
```

Identifies elements without elevation tags and prioritizes alpine huts.

Output: `output/osm_data_filtered.json`

### Step 3: Enrich with Elevation

```bash
# Test with limited items first
./elevate-romania --enrich --limit 10
```

Fetches elevation from OpenTopoData (SRTM dataset).

Output: `output/osm_data_enriched.json`

⚠️ **Rate Limiting**: The API has a 1 req/sec limit. Processing large datasets takes time.

### Step 4: Validate Data

```bash
./elevate-romania --validate
```

Validates elevation ranges (0-2600m for Romania).

Output: `output/osm_data_validated.json`

### Step 5: Export to CSV

```bash
./elevate-romania --export-csv
```

Creates a CSV file for analysis.

Output: `output/elevation_data.csv`

### Step 6: Upload to OSM

#### Dry-Run First (Always!)

```bash
./elevate-romania --upload --dry-run
```

This shows what would be uploaded without making actual changes.

#### Actual Upload

You need OAuth 2.0 credentials. Two options:

**Option A: Interactive Setup**

```bash
./elevate-romania --upload --oauth-interactive
```

**Option B: Environment Variables**

Create a `.env` file:

```env
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret
ACCESS_TOKEN=your_access_token
```

Then run:

```bash
./elevate-romania --upload
```

## OAuth 2.0 Setup

### Register an OSM Application

1. Go to: https://www.openstreetmap.org/oauth2/applications
2. Click "Register new application"
3. Fill in:
   - **Name**: Elevație OSM România
   - **Redirect URI**: `http://127.0.0.1:8080/callback`
   - **Permissions**: `read_prefs`, `write_prefs`, `write_api`
4. Save your **Client ID** and **Client Secret**

### Get Access Token

Use the interactive flow:

```bash
./elevate-romania --oauth-interactive
```

Follow the prompts to authorize and get your access token.

## Common Workflows

### Testing Workflow

Test everything with a small dataset:

```bash
./elevate-romania --all --dry-run --limit 5
```

### Production Workflow

Process all data and upload:

```bash
# 1. Extract all data
./elevate-romania --extract

# 2. Filter
./elevate-romania --filter

# 3. Enrich (this will take time!)
./elevate-romania --enrich

# 4. Validate
./elevate-romania --validate

# 5. Export
./elevate-romania --export-csv

# 6. Dry-run upload first
./elevate-romania --upload --dry-run

# 7. If everything looks good, actual upload
./elevate-romania --upload --oauth-interactive
```

Or run all at once:

```bash
./elevate-romania --all --oauth-interactive
```

## Using Make

If you have `make` installed:

```bash
make help          # Show all commands
make build         # Build binary
make test          # Test with 5 items
make demo          # Demo with 10 items
make dry-run       # Full dry-run with 10 items
make clean         # Clean build artifacts
```

## Troubleshooting

### "Go: command not found"

Install Go from https://golang.org/dl/

### "Failed to query Overpass API"

The Overpass API might be busy. Wait a few minutes and try again.

### "Rate limit exceeded"

You're hitting the API too fast. The tool already includes rate limiting, but if you run multiple instances, you might hit limits.

### OAuth Errors

- Verify redirect URI is exactly: `http://127.0.0.1:8080/callback`
- Check all scopes are enabled: `read_prefs`, `write_prefs`, `write_api`
- Try the interactive flow: `--oauth-interactive`

### Elevation Validation Failures

Check `output/osm_data_validated.json` for details. Some common issues:
- Coordinates outside Romania
- API returned invalid data
- Network errors during enrichment

## Best Practices

1. **Always test first**: Use `--dry-run` and `--limit` flags
2. **Respect rate limits**: Don't run multiple instances simultaneously
3. **Validate before uploading**: Review the validated data file
4. **Use changesets wisely**: The tool groups changes into a single changeset
5. **Check results**: After upload, verify a few elements on OSM

## File Structure

```
elevate-romania/
├── main.go              # CLI and orchestration
├── extract.go           # Overpass API queries
├── filter.go            # Filter logic
├── enrich.go            # Elevation API
├── validate.go          # Data validation
├── csv_export.go        # CSV export
├── upload.go            # OSM upload with OAuth
├── utils.go             # Utilities
├── go.mod               # Go dependencies
├── README.md            # Full documentation
├── GETTING_STARTED.md   # This file
├── Makefile             # Build automation
├── build.sh             # Build script
├── .env.example         # Example environment file
└── output/              # Output directory (created automatically)
    ├── osm_data_raw.json
    ├── osm_data_filtered.json
    ├── osm_data_enriched.json
    ├── osm_data_validated.json
    └── elevation_data.csv
```

## Next Steps

1. Read the full [README.md](README.md) for detailed documentation
2. Review the [OSM import guidelines](https://wiki.openstreetmap.org/wiki/Import/Guidelines)
3. Join the OSM Romania community for questions
4. Test with small datasets before processing everything

## Support

- **OSM Romania**: https://wiki.openstreetmap.org/wiki/Romania
- **OpenTopoData**: https://www.opentopodata.org/
- **OSM Help**: https://help.openstreetmap.org/

Happy mapping! 🗺️
