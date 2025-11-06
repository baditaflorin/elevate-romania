"""
Extract module - Query OSM data using Overpass API
"""
import requests
import json
import time


class OverpassExtractor:
    """Extract OSM data using Overpass API"""
    
    def __init__(self, overpass_url="https://overpass-api.de/api/interpreter"):
        self.overpass_url = overpass_url
    
    def _query_overpass(self, query):
        """Execute Overpass query and return results"""
        try:
            response = requests.post(
                self.overpass_url,
                data={"data": query},
                timeout=300
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Error querying Overpass API: {e}")
            return None
    
    def get_train_stations(self, area="Romania"):
        """Extract all train stations in Romania"""
        query = """
        [out:json][timeout:300];
        area["name"="România"]["admin_level"="2"]->.romania;
        (
          node["railway"="station"](area.romania);
          node["railway"="halt"](area.romania);
        );
        out body;
        """
        
        print("Querying train stations...")
        result = self._query_overpass(query)
        if result and "elements" in result:
            print(f"Found {len(result['elements'])} train stations")
            return result["elements"]
        return []
    
    def get_accommodations(self, area="Romania"):
        """Extract all accommodations (hotels, guest houses, alpine huts, etc.)"""
        query = """
        [out:json][timeout:300];
        area["name"="România"]["admin_level"="2"]->.romania;
        (
          node["tourism"="hotel"](area.romania);
          node["tourism"="guest_house"](area.romania);
          node["tourism"="alpine_hut"](area.romania);
          node["tourism"="chalet"](area.romania);
          node["tourism"="hostel"](area.romania);
          node["tourism"="motel"](area.romania);
          way["tourism"="hotel"](area.romania);
          way["tourism"="guest_house"](area.romania);
          way["tourism"="alpine_hut"](area.romania);
          way["tourism"="chalet"](area.romania);
          way["tourism"="hostel"](area.romania);
          way["tourism"="motel"](area.romania);
        );
        out center;
        """
        
        print("Querying accommodations...")
        result = self._query_overpass(query)
        if result and "elements" in result:
            print(f"Found {len(result['elements'])} accommodations")
            return result["elements"]
        return []
    
    def get_all_data(self):
        """Extract both train stations and accommodations"""
        stations = self.get_train_stations()
        time.sleep(2)  # Be nice to Overpass API
        accommodations = self.get_accommodations()
        
        return {
            "train_stations": stations,
            "accommodations": accommodations
        }


if __name__ == "__main__":
    extractor = OverpassExtractor()
    data = extractor.get_all_data()
    
    # Save to file
    with open("osm_data_raw.json", "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    
    print(f"Total train stations: {len(data['train_stations'])}")
    print(f"Total accommodations: {len(data['accommodations'])}")
    print("Data saved to osm_data_raw.json")
