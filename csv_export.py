"""
CSV Export module - Export data to CSV format
"""
import csv
import json


class CSVExporter:
    """Export OSM elevation data to CSV"""
    
    @staticmethod
    def get_element_info(element):
        """Extract relevant information from OSM element"""
        element_type = element.get("type", "")
        element_id = element.get("id", "")
        
        # Get coordinates
        if element_type == "node":
            lat = element.get("lat", "")
            lon = element.get("lon", "")
        elif element_type == "way" and "center" in element:
            lat = element["center"].get("lat", "")
            lon = element["center"].get("lon", "")
        else:
            lat = ""
            lon = ""
        
        # Get tags
        tags = element.get("tags", {})
        name = tags.get("name", tags.get("ref", ""))
        ele = tags.get("ele", "")
        ele_source = tags.get("ele:source", "")
        tourism = tags.get("tourism", "")
        railway = tags.get("railway", "")
        
        return {
            "type": element_type,
            "id": element_id,
            "name": name,
            "lat": lat,
            "lon": lon,
            "elevation": ele,
            "elevation_source": ele_source,
            "tourism": tourism,
            "railway": railway,
            "osm_link": f"https://www.openstreetmap.org/{element_type}/{element_id}"
        }
    
    def export_to_csv(self, data, output_file):
        """
        Export data to CSV file
        
        Args:
            data: Dict with categories of OSM elements
            output_file: Output CSV filename
        """
        rows = []
        
        # Process all categories
        for category, elements in data.items():
            # Handle both list of elements and dict with valid_elements
            if isinstance(elements, dict) and "valid_elements" in elements:
                elements = elements["valid_elements"]
            
            if not isinstance(elements, list):
                continue
            
            for element in elements:
                info = self.get_element_info(element)
                info["category"] = category
                rows.append(info)
        
        # Write to CSV
        if rows:
            fieldnames = [
                "category", "type", "id", "name", "lat", "lon", 
                "elevation", "elevation_source", "tourism", "railway", "osm_link"
            ]
            
            with open(output_file, "w", newline="", encoding="utf-8") as csvfile:
                writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
                writer.writeheader()
                writer.writerows(rows)
            
            print(f"Exported {len(rows)} elements to {output_file}")
            return len(rows)
        else:
            print("No data to export")
            return 0


if __name__ == "__main__":
    # Load validated data
    with open("osm_data_validated.json", "r", encoding="utf-8") as f:
        data = json.load(f)
    
    # Export to CSV
    exporter = CSVExporter()
    exporter.export_to_csv(data, "elevation_data.csv")
    print("CSV export complete!")
