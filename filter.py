"""
Filter module - Filter OSM objects without elevation data
"""
import json


class ElevationFilter:
    """Filter OSM objects based on elevation data presence"""
    
    @staticmethod
    def has_elevation(element):
        """Check if element has elevation tag"""
        if "tags" in element:
            return "ele" in element["tags"]
        return False
    
    @staticmethod
    def is_alpine_hut(element):
        """Check if element is an alpine hut (cabană montană)"""
        if "tags" in element:
            return element["tags"].get("tourism") == "alpine_hut"
        return False
    
    @staticmethod
    def get_coordinates(element):
        """Extract coordinates from element"""
        if element["type"] == "node":
            return element.get("lat"), element.get("lon")
        elif element["type"] == "way" and "center" in element:
            return element["center"].get("lat"), element["center"].get("lon")
        return None, None
    
    def filter_missing_elevation(self, elements):
        """Filter elements that don't have elevation data"""
        missing_ele = []
        
        for element in elements:
            if not self.has_elevation(element):
                lat, lon = self.get_coordinates(element)
                if lat is not None and lon is not None:
                    missing_ele.append(element)
        
        return missing_ele
    
    def prioritize_alpine_huts(self, elements):
        """Separate alpine huts from other elements for priority processing"""
        alpine_huts = []
        others = []
        
        for element in elements:
            if self.is_alpine_hut(element):
                alpine_huts.append(element)
            else:
                others.append(element)
        
        return alpine_huts, others
    
    def filter_data(self, data):
        """Filter all data and prioritize alpine huts"""
        result = {
            "train_stations": [],
            "alpine_huts": [],
            "other_accommodations": []
        }
        
        # Filter train stations
        if "train_stations" in data:
            result["train_stations"] = self.filter_missing_elevation(data["train_stations"])
        
        # Filter accommodations and prioritize alpine huts
        if "accommodations" in data:
            missing_ele = self.filter_missing_elevation(data["accommodations"])
            alpine_huts, others = self.prioritize_alpine_huts(missing_ele)
            result["alpine_huts"] = alpine_huts
            result["other_accommodations"] = others
        
        return result


if __name__ == "__main__":
    # Load raw data
    with open("output/osm_data_raw.json", "r", encoding="utf-8") as f:
        data = json.load(f)
    
    # Filter
    filter_obj = ElevationFilter()
    filtered = filter_obj.filter_data(data)
    
    # Save filtered data
    with open("output/osm_data_filtered.json", "w", encoding="utf-8") as f:
        json.dump(filtered, f, indent=2, ensure_ascii=False)
    
    print(f"Train stations without elevation: {len(filtered['train_stations'])}")
    print(f"Alpine huts without elevation: {len(filtered['alpine_huts'])}")
    print(f"Other accommodations without elevation: {len(filtered['other_accommodations'])}")
    print("Filtered data saved to output/osm_data_filtered.json")
