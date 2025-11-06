#!/usr/bin/env python3
"""
Demo script with sample data to demonstrate the workflow
without hitting external APIs
"""

import json
import os

# Sample OSM data
sample_train_stations = [
    {
        "type": "node",
        "id": 1000001,
        "lat": 45.5013,
        "lon": 25.4514,
        "tags": {
            "name": "Gara Sinaia",
            "railway": "station"
        }
    },
    {
        "type": "node",
        "id": 1000002,
        "lat": 45.3505,
        "lon": 25.5450,
        "tags": {
            "name": "Gara Bușteni",
            "railway": "station",
            "ele": "900"  # This one already has elevation
        }
    }
]

sample_accommodations = [
    {
        "type": "node",
        "id": 2000001,
        "lat": 45.4167,
        "lon": 25.4500,
        "tags": {
            "name": "Cabana Padina",
            "tourism": "alpine_hut"
        }
    },
    {
        "type": "node",
        "id": 2000002,
        "lat": 44.4268,
        "lon": 26.1025,
        "tags": {
            "name": "Hotel București",
            "tourism": "hotel"
        }
    },
    {
        "type": "way",
        "id": 2000003,
        "center": {
            "lat": 45.6428,
            "lon": 25.5887
        },
        "tags": {
            "name": "Pensiunea Brașov",
            "tourism": "guest_house"
        }
    }
]

def create_demo_data():
    """Create demo data files"""
    print("Creating demo data files...")
    
    # Raw data
    raw_data = {
        "train_stations": sample_train_stations,
        "accommodations": sample_accommodations
    }
    
    with open("demo_osm_data_raw.json", "w", encoding="utf-8") as f:
        json.dump(raw_data, f, indent=2, ensure_ascii=False)
    
    print("✓ Created demo_osm_data_raw.json")
    
    # Simulate filtered data (remove items with elevation)
    from filter import ElevationFilter
    
    filter_obj = ElevationFilter()
    filtered = filter_obj.filter_data(raw_data)
    
    with open("demo_osm_data_filtered.json", "w", encoding="utf-8") as f:
        json.dump(filtered, f, indent=2, ensure_ascii=False)
    
    print(f"✓ Created demo_osm_data_filtered.json")
    print(f"  - Train stations without elevation: {len(filtered['train_stations'])}")
    print(f"  - Alpine huts without elevation: {len(filtered['alpine_huts'])}")
    print(f"  - Other accommodations: {len(filtered['other_accommodations'])}")
    
    # Simulate enriched data (add mock elevations)
    enriched = {
        "train_stations": [],
        "alpine_huts": [],
        "other_accommodations": []
    }
    
    # Add mock elevations
    for station in filtered["train_stations"]:
        station_copy = station.copy()
        if "tags" not in station_copy:
            station_copy["tags"] = {}
        # Mock elevation based on Sinaia area
        station_copy["tags"]["ele"] = "850.0"
        station_copy["tags"]["ele:source"] = "SRTM"
        station_copy["elevation_fetched"] = 850.0
        enriched["train_stations"].append(station_copy)
    
    for hut in filtered["alpine_huts"]:
        hut_copy = hut.copy()
        if "tags" not in hut_copy:
            hut_copy["tags"] = {}
        # Alpine huts are typically at higher elevation
        hut_copy["tags"]["ele"] = "1850.0"
        hut_copy["tags"]["ele:source"] = "SRTM"
        hut_copy["elevation_fetched"] = 1850.0
        enriched["alpine_huts"].append(hut_copy)
    
    for acc in filtered["other_accommodations"]:
        acc_copy = acc.copy()
        if "tags" not in acc_copy:
            acc_copy["tags"] = {}
        # Lower elevation for hotels
        acc_copy["tags"]["ele"] = "90.0"
        acc_copy["tags"]["ele:source"] = "SRTM"
        acc_copy["elevation_fetched"] = 90.0
        enriched["other_accommodations"].append(acc_copy)
    
    with open("demo_osm_data_enriched.json", "w", encoding="utf-8") as f:
        json.dump(enriched, f, indent=2, ensure_ascii=False)
    
    print("✓ Created demo_osm_data_enriched.json (with mock elevations)")
    
    # Validate
    from validate import ElevationValidator
    
    validator = ElevationValidator()
    results = validator.validate_all(enriched)
    
    output = {}
    for category, validation in results.items():
        output[category] = {
            "valid_count": len(validation["valid"]),
            "invalid_count": len(validation["invalid"]),
            "valid_elements": validation["valid"]
        }
    
    with open("demo_osm_data_validated.json", "w", encoding="utf-8") as f:
        json.dump(output, f, indent=2, ensure_ascii=False)
    
    print("✓ Created demo_osm_data_validated.json")
    
    # Export CSV
    from csv_export import CSVExporter
    
    exporter = CSVExporter()
    count = exporter.export_to_csv(output, "demo_elevation_data.csv")
    
    print(f"✓ Created demo_elevation_data.csv ({count} elements)")
    
    # Dry-run upload
    from upload import OSMUploader
    
    valid_data = {}
    for category, info in output.items():
        if "valid_elements" in info:
            valid_data[category] = info["valid_elements"]
    
    print("\n" + "="*60)
    print("Simulating upload (dry-run)...")
    print("="*60)
    
    uploader = OSMUploader(dry_run=True)
    stats = uploader.upload_all(valid_data)
    
    print("\n" + "="*60)
    print("DEMO COMPLETE!")
    print("="*60)
    print("\nGenerated files:")
    print("  - demo_osm_data_raw.json")
    print("  - demo_osm_data_filtered.json")
    print("  - demo_osm_data_enriched.json")
    print("  - demo_osm_data_validated.json")
    print("  - demo_elevation_data.csv")
    print("\nThis demonstrates the complete workflow with sample data.")
    print("For real usage, use main.py with actual OSM data.")
    print("="*60 + "\n")


if __name__ == "__main__":
    create_demo_data()
