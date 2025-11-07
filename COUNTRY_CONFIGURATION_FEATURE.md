# Country Configuration Feature

## Problem Statement
The application was hardcoded to work only with Romania ("România"). Users needed the ability to:
1. Target different countries without changing the code
2. Easily discover available countries from OpenStreetMap
3. Have the country name reflected in changeset messages

## Solution Implemented

### 1. CLI Country Parameter ✓

**Added Flags:**
- `--country` - Specify country name (default: "România")
- `--list-countries` - Query and display all admin_level=2 countries

**Example Usage:**
```bash
./elevate-romania --country "Moldova" --extract
./elevate-romania --country "France" --all --dry-run
./elevate-romania --list-countries
```

### 2. Dynamic Overpass Queries ✓

**Updated Files:**
- `src/extract.go` - Parameterized Overpass queries

**Implementation:**
- Added `Country` field to `OverpassExtractor` struct
- Factory method `NewOverpassExtractor(country string)` accepts country parameter
- Queries dynamically build `area["name"="<country>"]["admin_level"="2"]`
- Country name displayed in extraction logs

**Before:**
```go
query := `
[out:json][timeout:180];
area["name"="România"]["admin_level"="2"]->.romania;
...
`
```

**After:**
```go
query := fmt.Sprintf(`
[out:json][timeout:180];
area["name"="%s"]["admin_level"="2"]->.country;
...
`, e.Country)
```

### 3. Dynamic Changeset Messages ✓

**Updated Files:**
- `src/upload.go` - Parameterized changeset comments

**Implementation:**
- Added `country` field to `OSMUploader` struct
- Updated `NewOSMUploader()` to accept country parameter
- Changeset message includes specified country name

**Before:**
```go
changesetComment := fmt.Sprintf("Add elevation data to %d locations in Romania (...)", totalElements)
```

**After:**
```go
changesetComment := fmt.Sprintf("Add elevation data to %d locations in %s (...)", totalElements, u.country)
```

### 4. Configuration Support ✓

**Updated Files:**
- `src/main.go` - Added country flags and parameter passing
- `src/config.go` - Support for COUNTRY configuration key
- `src/factory.go` - Factory method uses country from config

**Implementation:**
- Country parameter flows from CLI → Config → Factory → Components
- Default value "România" maintained for backward compatibility
- Country displayed in main banner

### 5. List Countries Feature ✓

**New Function:**
- `runListCountries()` in `extract.go`

**Implementation:**
```go
query := `
[out:json][timeout:60];
area["admin_level"="2"];
out tags;
`
```

**Output Format:**
```
Found 195 countries:

  România                                   (int_name: Romania)
  Moldova
  France
  Deutschland                               (int_name: Germany)
  España                                    (int_name: Spain)
  ...

Usage: elevate-romania --country "Country Name" --extract
Note: Use the exact name (case-sensitive) as shown above
```

### 6. Comprehensive Testing ✓

**New Test File:**
- `src/extract_test.go` - Tests for country parameter functionality

**Test Coverage:**
```go
TestOverpassExtractorCountryParameter - Verify country field is set correctly
TestOverpassExtractorGetTrainStationsQuery - Verify train station queries use country
TestOverpassExtractorGetAccommodationsQuery - Verify accommodation queries use country
TestCountryInfoStructure - Verify CountryInfo struct
TestNewOverpassExtractor - Verify factory creates extractor with country
```

**Test Results:**
```
=== RUN   TestOverpassExtractorCountryParameter
=== RUN   TestOverpassExtractorCountryParameter/Default_Romania
=== RUN   TestOverpassExtractorCountryParameter/Moldova
=== RUN   TestOverpassExtractorCountryParameter/France
--- PASS: TestOverpassExtractorCountryParameter (0.00s)
...
PASS
ok  	elevate-romania	0.002s
```

All existing tests continue to pass.

### 7. Updated Documentation ✓

**Updated Files:**
- `src/README.md` - Comprehensive documentation for country feature

**Added Sections:**
- "Country Selection" - How to use the --country flag
- "Working with Different Countries" - Complete guide
- "List Available Countries" - Documentation for --list-countries
- Updated examples throughout to show multi-country usage

## How It Works

### Architecture Flow

```
CLI (--country "Moldova")
    ↓
main.go (country parameter)
    ↓
Config (Set "COUNTRY" key)
    ↓
Factory (CreateOverpassExtractor)
    ↓
OverpassExtractor (uses country in queries)
    ↓
Overpass API (queries Moldova data)

CLI (--country "Moldova")
    ↓
main.go (country parameter)
    ↓
upload.go (runUpload with country)
    ↓
OSMUploader (uses country in changeset)
    ↓
OSM API (changeset message includes "Moldova")
```

### Usage Examples

**1. Default Usage (România):**
```bash
./elevate-romania --extract
# Uses default country: România
```

**2. Different Country:**
```bash
./elevate-romania --country "Moldova" --all --dry-run
# Extracts, enriches, and previews upload for Moldova
```

**3. Discover Countries:**
```bash
./elevate-romania --list-countries
# Shows all available countries from OpenStreetMap
```

**4. Complete Workflow for New Country:**
```bash
# Step 1: Find the exact country name
./elevate-romania --list-countries

# Step 2: Extract data for that country
./elevate-romania --country "France" --extract

# Step 3: Continue with normal workflow
./elevate-romania --filter
./elevate-romania --enrich --limit 10
./elevate-romania --validate
./elevate-romania --upload --dry-run
```

## Files Modified

### Core Changes
- `src/main.go` - Added CLI flags, parameter passing (19 lines changed)
- `src/extract.go` - Parameterized queries, added list-countries (99 lines added)
- `src/upload.go` - Parameterized changeset messages (8 lines changed)
- `src/factory.go` - Factory method with country parameter (10 lines changed)

### Documentation
- `src/README.md` - Comprehensive country feature documentation (60 lines added)

### Tests
- `src/extract_test.go` - New test file (92 lines)

### Total Impact
- 5 files modified
- 1 file created
- ~290 lines added/modified
- 0 breaking changes (backward compatible)

## Backward Compatibility

✓ **Fully Backward Compatible**
- Default country is "România" (same as before)
- All existing commands work without changes
- No breaking changes to APIs or data structures

## Testing

### Manual Testing
```bash
# Build succeeds
go build -o elevate-romania .

# Help shows new flags
./elevate-romania --help
# Output includes:
#   -country string
#   	Country name to target (int_name from OSM) (default "România")
#   -list-countries
#   	List all available admin_level=2 countries

# Country parameter works
./elevate-romania --country "Moldova" --extract
# Output includes:
#   Adding elevation to train stations and accommodations in Moldova
#   STEP 1: EXTRACT - Querying Overpass API for Moldova
#   Querying train stations in Moldova...
```

### Automated Testing
```bash
cd src && go test -v
# All tests pass including new extract_test.go tests
```

## Benefits

1. **Flexibility** - Work with any country without code changes
2. **Discoverability** - Easy to find available countries via --list-countries
3. **Runtime Configuration** - No need to rebuild or edit code for different countries
4. **Proper Attribution** - Changeset messages reflect the correct country
5. **User-Friendly** - Clear CLI interface with helpful examples
6. **Maintainable** - Clean parameter passing through the application layers
7. **Testable** - Comprehensive test coverage for new functionality

## Migration Guide

### For Existing Users
No migration needed! The default behavior is unchanged:
```bash
./elevate-romania --extract  # Still uses România by default
```

### For New Countries
Simply specify the country name:
```bash
./elevate-romania --country "Moldova" --extract
```

### Finding Country Names
Use the new list-countries command:
```bash
./elevate-romania --list-countries
```

## Future Enhancements

Possible future improvements:
- Country aliases (e.g., "Germany" → "Deutschland")
- Country-specific elevation validation ranges
- Save last-used country in config file
- Support for regions/states within countries
- Batch processing multiple countries
