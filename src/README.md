# Elevație OSM România - Go Implementation

Add elevation data to train stations and accommodations in Romania on OpenStreetMap.

## Features

- Extract OSM data for train stations and accommodations in Romania
- Filter elements missing elevation data
- Enrich with elevation from OpenTopoData (SRTM dataset)
- Validate elevation ranges (0-2600m for Romania)
- Export results to CSV
- Upload to OSM with OAuth 2.0 authentication
- Dry-run mode for testing

## Installation

### Prerequisites

- Go 1.21 or higher
- OSM account with OAuth 2.0 app registered at https://www.openstreetmap.org/oauth2/applications

### Build

```bash
# Clone or copy the project
cd elevate-romania

# Download dependencies
go mod download

# Build the binary
go build -o elevate-romania .
```

## Configuration

### OAuth 2.0 Setup

1. Register an OAuth 2.0 application at https://www.openstreetmap.org/oauth2/applications
2. Set redirect URI to: `http://127.0.0.1:8080/callback`
3. Request permissions: `read_prefs`, `write_prefs`, `write_api`
4. Save your credentials

### Environment Variables

Create a `.env` file:

```env
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret
ACCESS_TOKEN=your_access_token
```

Or use the interactive OAuth flow with `--oauth-interactive`.

## Usage

### Basic Commands

```bash
# Show help
./elevate-romania --help

# Run complete pipeline in dry-run mode
./elevate-romania --all --dry-run

# Run individual steps
./elevate-romania --extract                  # Step 1: Extract from OSM
./elevate-romania --filter                   # Step 2: Filter missing elevation
./elevate-romania --enrich --limit 10        # Step 3: Enrich (test with 10 items)
./elevate-romania --validate                 # Step 4: Validate ranges
./elevate-romania --export-csv               # Step 5: Export to CSV
./elevate-romania --upload --dry-run         # Step 6: Preview upload

# Run with OAuth interactive setup
./elevate-romania --upload --oauth-interactive
```

### Complete Workflow

```bash
# 1. Extract data from OSM
./elevate-romania --extract

# 2. Filter elements without elevation
./elevate-romania --filter

# 3. Enrich with elevation data (limited for testing)
./elevate-romania --enrich --limit 10

# 4. Validate elevation ranges
./elevate-romania --validate

# 5. Export to CSV
./elevate-romania --export-csv

# 6. Upload to OSM (dry-run first!)
./elevate-romania --upload --dry-run

# 7. Actual upload (with OAuth)
./elevate-romania --upload --oauth-interactive
```

### Run Everything at Once

```bash
# Test the complete workflow with 10 items
./elevate-romania --all --dry-run --limit 10

# Production run (uploads to OSM)
./elevate-romania --all --oauth-interactive
```

## Output Files

All files are saved in the `output/` directory:

- `osm_data_raw.json` - Raw data from Overpass API
- `osm_data_filtered.json` - Elements without elevation
- `osm_data_enriched.json` - Elements with fetched elevation
- `osm_data_validated.json` - Validated elements (0-2600m)
- `elevation_data.csv` - CSV export for analysis

## Architecture

### Modules

- `main.go` - CLI and orchestration
- `extract.go` - Query Overpass API for OSM data
- `filter.go` - Filter elements without elevation
- `enrich.go` - Fetch elevation from OpenTopoData
- `validate.go` - Validate elevation ranges
- `csv_export.go` - Export to CSV format
- `upload.go` - Upload to OSM with OAuth 2.0
- `utils.go` - JSON I/O utilities

### Data Flow

```
OSM (Overpass API)
  ↓
Extract → Filter → Enrich → Validate → Export CSV
                                      ↓
                                    Upload to OSM
```

## API Rate Limits

- **Overpass API**: Respect the fair use policy, add delays between requests
- **OpenTopoData**: 1 request per second (configurable in code)
- **OSM API**: 1 request per second for uploads

## Safety Features

- **Dry-run mode**: Preview changes before uploading
- **Validation**: Check elevation ranges (0-2600m for Romania)
- **Priority processing**: Alpine huts processed first
- **Rate limiting**: Automatic delays between API calls
- **Changeset management**: Groups changes with descriptive comments

## Elevation Data Sources

- **OpenTopoData**: SRTM 30m resolution dataset
- **Coverage**: Global, suitable for Romania
- **Accuracy**: ±16m vertical accuracy

## Contributing

1. Test changes with `--dry-run` flag
2. Use `--limit` for testing with small datasets
3. Validate before uploading to OSM
4. Follow OSM's data import guidelines

## License

Apache License 2.0

## Acknowledgments

- OpenStreetMap contributors
- OpenTopoData for elevation API
- SRTM for elevation dataset

## Troubleshomarks

### OAuth Errors

If you get OAuth errors:
1. Verify redirect URI is exactly: `http://127.0.0.1:8080/callback`
2. Check that all required scopes are enabled
3. Use `--oauth-interactive` for guided setup

### Rate Limiting

If you hit rate limits:
1. Increase delays in `enrich.go`
2. Use `--limit` to process fewer items
3. Wait and retry later

### Validation Errors

If elements fail validation:
1. Check `output/osm_data_validated.json` for details
2. Verify elevation ranges are appropriate
3. Review invalid examples in console output

## Support

For issues with:
- **OSM data**: Check OpenStreetMap forums
- **Elevation API**: See OpenTopoData documentation
- **This tool**: Open an issue on GitHub
