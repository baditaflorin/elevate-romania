#!/usr/bin/env python3
"""
Main orchestration script for OSM elevation enrichment
Elevație OSM România - Add elevation to train stations and accommodations

Usage:
    python main.py --help
    python main.py --extract           # Extract data from OSM
    python main.py --filter            # Filter data without elevation
    python main.py --enrich            # Enrich with elevation data
    python main.py --validate          # Validate elevation ranges
    python main.py --export-csv        # Export to CSV
    python main.py --upload --dry-run  # Dry-run upload to OSM
    python main.py --all --dry-run     # Run complete pipeline
"""

import argparse
import json
import sys
from datetime import datetime

from extract import OverpassExtractor
from filter import ElevationFilter
from enrich import ElevationEnricher
from validate import ElevationValidator
from csv_export import CSVExporter
from upload import OSMUploader


def run_extract():
    """Step 1: Extract data from OSM"""
    print("\n" + "="*60)
    print("STEP 1: EXTRACT - Querying Overpass API")
    print("="*60)
    
    extractor = OverpassExtractor()
    data = extractor.get_all_data()
    
    # Save to file
    with open("osm_data_raw.json", "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    
    print(f"\n✓ Extracted {len(data['train_stations'])} train stations")
    print(f"✓ Extracted {len(data['accommodations'])} accommodations")
    print("✓ Data saved to osm_data_raw.json\n")
    
    return data


def run_filter():
    """Step 2: Filter elements without elevation"""
    print("\n" + "="*60)
    print("STEP 2: FILTER - Identifying elements without elevation")
    print("="*60)
    
    # Load raw data
    try:
        with open("osm_data_raw.json", "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Error: osm_data_raw.json not found. Run --extract first.")
        return None
    
    # Filter
    filter_obj = ElevationFilter()
    filtered = filter_obj.filter_data(data)
    
    # Save filtered data
    with open("osm_data_filtered.json", "w", encoding="utf-8") as f:
        json.dump(filtered, f, indent=2, ensure_ascii=False)
    
    print(f"\n✓ Train stations without elevation: {len(filtered['train_stations'])}")
    print(f"✓ Alpine huts without elevation: {len(filtered['alpine_huts'])} (PRIORITY)")
    print(f"✓ Other accommodations without elevation: {len(filtered['other_accommodations'])}")
    print("✓ Filtered data saved to osm_data_filtered.json\n")
    
    return filtered


def run_enrich(max_items=None):
    """Step 3: Enrich with elevation data"""
    print("\n" + "="*60)
    print("STEP 3: ENRICH - Fetching elevation from OpenTopoData")
    print("="*60)
    
    # Load filtered data
    try:
        with open("osm_data_filtered.json", "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Error: osm_data_filtered.json not found. Run --filter first.")
        return None
    
    # Enrich with elevation
    enricher = ElevationEnricher(api_type="opentopo", rate_limit=1.0)
    
    enriched_data = {}
    
    # Process alpine huts first (priority)
    if data.get("alpine_huts"):
        print("\n[PRIORITY] Enriching alpine huts...")
        enriched_data["alpine_huts"] = enricher.enrich_elements(
            data["alpine_huts"], max_count=max_items
        )
    
    # Process train stations
    if data.get("train_stations"):
        print("\nEnriching train stations...")
        enriched_data["train_stations"] = enricher.enrich_elements(
            data["train_stations"], max_count=max_items
        )
    
    # Process other accommodations
    if data.get("other_accommodations"):
        print("\nEnriching other accommodations...")
        enriched_data["other_accommodations"] = enricher.enrich_elements(
            data["other_accommodations"], max_count=max_items
        )
    
    # Save enriched data
    with open("osm_data_enriched.json", "w", encoding="utf-8") as f:
        json.dump(enriched_data, f, indent=2, ensure_ascii=False)
    
    print("\n✓ Enrichment complete!")
    print(f"  Alpine huts: {len(enriched_data.get('alpine_huts', []))}")
    print(f"  Train stations: {len(enriched_data.get('train_stations', []))}")
    print(f"  Other accommodations: {len(enriched_data.get('other_accommodations', []))}")
    print("✓ Enriched data saved to osm_data_enriched.json\n")
    
    return enriched_data


def run_validate():
    """Step 4: Validate elevation ranges"""
    print("\n" + "="*60)
    print("STEP 4: VALIDATE - Checking elevation ranges (0-2600m)")
    print("="*60)
    
    # Load enriched data
    try:
        with open("osm_data_enriched.json", "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Error: osm_data_enriched.json not found. Run --enrich first.")
        return None
    
    # Validate
    validator = ElevationValidator(min_elevation=0, max_elevation=2600)
    results = validator.validate_all(data)
    
    # Save validation results
    output = {}
    for category, validation in results.items():
        output[category] = {
            "valid_count": len(validation["valid"]),
            "invalid_count": len(validation["invalid"]),
            "valid_elements": validation["valid"]
        }
    
    with open("osm_data_validated.json", "w", encoding="utf-8") as f:
        json.dump(output, f, indent=2, ensure_ascii=False)
    
    print("\n✓ Validation complete! Results saved to osm_data_validated.json\n")
    
    return results


def run_export_csv():
    """Step 5: Export to CSV"""
    print("\n" + "="*60)
    print("STEP 5: EXPORT - Creating CSV output")
    print("="*60)
    
    # Load validated data
    try:
        with open("osm_data_validated.json", "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Error: osm_data_validated.json not found. Run --validate first.")
        return None
    
    # Export to CSV
    exporter = CSVExporter()
    count = exporter.export_to_csv(data, "elevation_data.csv")
    
    print(f"\n✓ Exported {count} elements to elevation_data.csv\n")
    
    return count


def run_upload(dry_run=True, username=None, password=None):
    """Step 6: Upload to OSM"""
    print("\n" + "="*60)
    if dry_run:
        print("STEP 6: UPLOAD (DRY-RUN) - Preview changes")
    else:
        print("STEP 6: UPLOAD - Uploading to OpenStreetMap")
    print("="*60)
    
    # Load validated data
    try:
        with open("osm_data_validated.json", "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError:
        print("Error: osm_data_validated.json not found. Run --validate first.")
        return None
    
    # Extract valid elements
    valid_data = {}
    for category, info in data.items():
        if "valid_elements" in info:
            valid_data[category] = info["valid_elements"]
    
    # Upload
    uploader = OSMUploader(username=username, password=password, dry_run=dry_run)
    stats = uploader.upload_all(valid_data)
    
    # Display statistics
    print("\n" + "="*60)
    if dry_run:
        print("UPLOAD STATISTICS (DRY-RUN)")
    else:
        print("UPLOAD STATISTICS")
    print("="*60)
    
    for category, category_stats in stats.items():
        print(f"\n{category}:")
        print(f"  Total: {category_stats['total']}")
        print(f"  Successful: {category_stats['successful']}")
        print(f"  Failed: {category_stats['failed']}")
    
    print("\n" + "="*60 + "\n")
    
    return stats


def main():
    parser = argparse.ArgumentParser(
        description="Elevație OSM România - Add elevation to train stations and accommodations",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --all --dry-run              # Run complete pipeline in dry-run mode
  %(prog)s --extract --filter           # Extract and filter data
  %(prog)s --enrich --limit 10          # Enrich first 10 items (for testing)
  %(prog)s --upload --dry-run           # Preview upload changes
  %(prog)s --upload                     # Actual upload (requires credentials)
        """
    )
    
    # Step options
    parser.add_argument("--extract", action="store_true", help="Extract data from OSM")
    parser.add_argument("--filter", action="store_true", help="Filter elements without elevation")
    parser.add_argument("--enrich", action="store_true", help="Enrich with elevation data")
    parser.add_argument("--validate", action="store_true", help="Validate elevation ranges")
    parser.add_argument("--export-csv", action="store_true", help="Export to CSV")
    parser.add_argument("--upload", action="store_true", help="Upload to OSM")
    parser.add_argument("--all", action="store_true", help="Run all steps")
    
    # Options
    parser.add_argument("--dry-run", action="store_true", help="Dry-run mode (don't upload)")
    parser.add_argument("--limit", type=int, help="Limit number of items to process (for testing)")
    parser.add_argument("--username", help="OSM username")
    parser.add_argument("--password", help="OSM password")
    
    args = parser.parse_args()
    
    # Check if any action is specified
    if not any([args.extract, args.filter, args.enrich, args.validate, 
                args.export_csv, args.upload, args.all]):
        parser.print_help()
        return
    
    print("\n" + "="*60)
    print("ELEVAȚIE OSM ROMÂNIA")
    print("Adding elevation to train stations and accommodations")
    print(f"Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("="*60)
    
    try:
        # Run steps
        if args.all or args.extract:
            run_extract()
        
        if args.all or args.filter:
            run_filter()
        
        if args.all or args.enrich:
            run_enrich(max_items=args.limit)
        
        if args.all or args.validate:
            run_validate()
        
        if args.all or args.export_csv:
            run_export_csv()
        
        if args.all or args.upload:
            # Default to dry-run unless explicitly disabled
            dry_run = args.dry_run if args.dry_run else True
            if not dry_run and not args.username:
                print("\nError: Username and password required for actual upload")
                print("Use --dry-run for testing or provide --username and --password")
                return
            
            run_upload(dry_run=dry_run, username=args.username, password=args.password)
        
        print("\n" + "="*60)
        print("COMPLETED SUCCESSFULLY!")
        print(f"Finished: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*60 + "\n")
        
    except KeyboardInterrupt:
        print("\n\nProcess interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n\nError: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
