# Elevație OSM România - Quick Reference

## Build & Run

```bash
# Build
./build.sh
# or
go build -o elevate-romania .

# Help
./elevate-romania --help

# Quick test
./elevate-romania --all --dry-run --limit 5
```

## Command Flags

| Flag | Description |
|------|-------------|
| `--extract` | Extract data from OSM |
| `--filter` | Filter elements without elevation |
| `--enrich` | Fetch elevation data |
| `--validate` | Validate elevation ranges |
| `--export-csv` | Export to CSV |
| `--upload` | Upload to OSM |
| `--all` | Run all steps |
| `--dry-run` | Don't actually upload |
| `--limit N` | Process max N items |
| `--oauth-interactive` | Interactive OAuth setup |

## Common Commands

```bash
# Extract only
./elevate-romania --extract

# Test enrichment with 10 items
./elevate-romania --enrich --limit 10

# Complete dry-run
./elevate-romania --all --dry-run --limit 10

# Production upload
./elevate-romania --all --oauth-interactive
```

## Make Shortcuts

```bash
make help        # Show help
make build       # Build binary
make test        # Test with 5 items
make demo        # Demo with 10 items
make dry-run     # Full dry-run
make clean       # Clean outputs
```

## Output Files

```
output/
├── osm_data_raw.json        # Step 1: Raw OSM data
├── osm_data_filtered.json   # Step 2: Filtered data
├── osm_data_enriched.json   # Step 3: With elevation
├── osm_data_validated.json  # Step 4: Validated
└── elevation_data.csv       # Step 5: CSV export
```

## OAuth Setup

1. Register app: https://www.openstreetmap.org/oauth2/applications
2. Redirect URI: `http://127.0.0.1:8080/callback`
3. Scopes: `read_prefs`, `write_prefs`, `write_api`
4. Run: `./elevate-romania --oauth-interactive`

## Environment Variables

Create `.env` file:

```env
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret
ACCESS_TOKEN=your_access_token
```

## Workflow

```
┌─────────┐
│ Extract │ --extract
└────┬────┘
     │
┌────▼───┐
│ Filter │ --filter
└────┬───┘
     │
┌────▼────┐
│ Enrich  │ --enrich [--limit N]
└────┬────┘
     │
┌────▼─────┐
│ Validate │ --validate
└────┬─────┘
     │
┌────▼────┐
│ Export  │ --export-csv
└────┬────┘
     │
┌────▼────┐
│ Upload  │ --upload [--dry-run]
└─────────┘
```

## Error Messages

| Error | Solution |
|-------|----------|
| "Go: command not found" | Install Go from golang.org |
| "OSM data not found" | Run previous step first |
| "Rate limit exceeded" | Wait and retry |
| "OAuth error" | Check credentials and scopes |
| "Validation failed" | Review output/osm_data_validated.json |

## Rate Limits

- **Overpass API**: Fair use, add delays
- **OpenTopoData**: 1 req/sec (built-in)
- **OSM API**: 1 req/sec (built-in)

## Best Practices

1. ✅ Always test with `--dry-run` first
2. ✅ Use `--limit` for initial tests
3. ✅ Validate before uploading
4. ✅ Review CSV output
5. ✅ Check a few uploaded elements on OSM
6. ⚠️ Don't run multiple instances
7. ⚠️ Respect API rate limits
8. ⚠️ Follow OSM import guidelines

## File Sizes (Approximate)

| Item Count | Raw JSON | Enriched | CSV |
|------------|----------|----------|-----|
| 10 | 5 KB | 6 KB | 1 KB |
| 100 | 50 KB | 60 KB | 10 KB |
| 1,000 | 500 KB | 600 KB | 100 KB |
| 10,000 | 5 MB | 6 MB | 1 MB |

## Processing Time (Approximate)

| Step | 10 Items | 100 Items | 1,000 Items |
|------|----------|-----------|-------------|
| Extract | 10s | 20s | 30s |
| Filter | < 1s | < 1s | < 1s |
| Enrich | 10s | 100s | 1,000s |
| Validate | < 1s | < 1s | < 1s |
| Export | < 1s | < 1s | < 1s |
| Upload | 10s | 100s | 1,000s |

**Note**: Enrich and Upload are rate-limited to 1 req/sec

## Data Coverage

### Train Stations
- Type: `railway=station` or `railway=halt`
- Region: România (admin_level=2)
- Geometry: Nodes only

### Accommodations
- Types: hotel, guest_house, alpine_hut, chalet, hostel, motel
- Region: România (admin_level=2)
- Geometry: Nodes and ways (with center)

### Elevation Range
- Minimum: 0m (Black Sea)
- Maximum: 2,600m (Mount Moldoveanu: 2,544m)

## Troubleshooting Quick Fixes

```bash
# Clean and rebuild
make clean
./build.sh

# Test connectivity
curl -I https://overpass-api.de/api/interpreter
curl -I https://api.opentopodata.org/

# Check output files
ls -lh output/

# View last 10 lines of output
tail -n 10 output/osm_data_validated.json

# Count elements
grep -c '"type"' output/osm_data_raw.json
```

## Support Links

- 📖 Full docs: [README.md](README.md)
- 🚀 Getting started: [GETTING_STARTED.md](GETTING_STARTED.md)
- 📊 Project summary: [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)
- 🗺️ OSM Romania: https://wiki.openstreetmap.org/wiki/Romania
- 🔒 OAuth setup: https://www.openstreetmap.org/oauth2/applications
- ⛰️ Elevation API: https://www.opentopodata.org/

## Version Info

- **Go Version**: 1.21+
- **API**: OSM API v0.6
- **Elevation**: SRTM 30m
- **OAuth**: OAuth 2.0

---

💡 **Tip**: Start with `make demo` to see the complete workflow!
