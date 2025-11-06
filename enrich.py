"""
Enrich module - Fetch elevation data from elevation APIs
"""
import requests
import time


class ElevationEnricher:
    """Enrich OSM data with elevation from external APIs"""
    
    def __init__(self, api_type="opentopo", rate_limit=1.0):
        """
        Initialize enricher
        
        Args:
            api_type: 'opentopo' for OpenTopoData or 'open-elevation' for Open-Elevation
            rate_limit: seconds to wait between requests
        """
        self.api_type = api_type
        self.rate_limit = rate_limit
        
        if api_type == "opentopo":
            # OpenTopoData with SRTM dataset
            self.base_url = "https://api.opentopodata.org/v1/srtm30m"
        else:
            # Open-Elevation
            self.base_url = "https://api.open-elevation.com/api/v1/lookup"
    
    def get_elevation(self, lat, lon):
        """
        Get elevation for a single coordinate
        
        Returns:
            Elevation in meters or None if failed
        """
        try:
            if self.api_type == "opentopo":
                params = {"locations": f"{lat},{lon}"}
                response = requests.get(self.base_url, params=params, timeout=30)
                response.raise_for_status()
                data = response.json()
                
                if data.get("status") == "OK" and data.get("results"):
                    elevation = data["results"][0].get("elevation")
                    return elevation
            else:
                # Open-Elevation
                data = {"locations": [{"latitude": lat, "longitude": lon}]}
                response = requests.post(self.base_url, json=data, timeout=30)
                response.raise_for_status()
                result = response.json()
                
                if result.get("results"):
                    elevation = result["results"][0].get("elevation")
                    return elevation
            
            return None
            
        except requests.exceptions.RequestException as e:
            print(f"Error fetching elevation for {lat},{lon}: {e}")
            return None
    
    def enrich_element(self, element):
        """
        Add elevation to a single OSM element
        
        Returns:
            Element with elevation added or None if coordinates missing
        """
        # Get coordinates
        if element["type"] == "node":
            lat = element.get("lat")
            lon = element.get("lon")
        elif element["type"] == "way" and "center" in element:
            lat = element["center"].get("lat")
            lon = element["center"].get("lon")
        else:
            return None
        
        if lat is None or lon is None:
            return None
        
        # Get elevation
        elevation = self.get_elevation(lat, lon)
        
        if elevation is not None:
            # Add elevation to element
            if "tags" not in element:
                element["tags"] = {}
            element["tags"]["ele"] = str(round(elevation, 1))
            element["tags"]["ele:source"] = "SRTM"
            element["elevation_fetched"] = elevation
        
        # Rate limiting
        time.sleep(self.rate_limit)
        
        return element
    
    def enrich_elements(self, elements, max_count=None):
        """
        Enrich multiple elements with elevation
        
        Args:
            elements: List of OSM elements
            max_count: Maximum number to process (None for all)
        
        Returns:
            List of enriched elements
        """
        enriched = []
        count = 0
        
        for element in elements:
            if max_count and count >= max_count:
                break
            
            enriched_element = self.enrich_element(element)
            if enriched_element:
                enriched.append(enriched_element)
                count += 1
                if count % 10 == 0:
                    print(f"Processed {count} elements...")
        
        return enriched


if __name__ == "__main__":
    import json
    
    # Load filtered data
    with open("osm_data_filtered.json", "r", encoding="utf-8") as f:
        data = json.load(f)
    
    # Enrich with elevation
    enricher = ElevationEnricher(api_type="opentopo", rate_limit=1.0)
    
    enriched_data = {}
    
    # Process alpine huts first (priority)
    if data.get("alpine_huts"):
        print("\nEnriching alpine huts...")
        enriched_data["alpine_huts"] = enricher.enrich_elements(data["alpine_huts"])
    
    # Process train stations
    if data.get("train_stations"):
        print("\nEnriching train stations...")
        enriched_data["train_stations"] = enricher.enrich_elements(data["train_stations"])
    
    # Process other accommodations
    if data.get("other_accommodations"):
        print("\nEnriching other accommodations...")
        enriched_data["other_accommodations"] = enricher.enrich_elements(data["other_accommodations"])
    
    # Save enriched data
    with open("osm_data_enriched.json", "w", encoding="utf-8") as f:
        json.dump(enriched_data, f, indent=2, ensure_ascii=False)
    
    print("\nEnrichment complete!")
    print(f"Alpine huts: {len(enriched_data.get('alpine_huts', []))}")
    print(f"Train stations: {len(enriched_data.get('train_stations', []))}")
    print(f"Other accommodations: {len(enriched_data.get('other_accommodations', []))}")
