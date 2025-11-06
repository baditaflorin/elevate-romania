"""
Upload module - Upload elevation data to OSM using osmapi
"""
import osmapi
import json
from getpass import getpass


class OSMUploader:
    """Upload elevation tags to OpenStreetMap"""
    
    def __init__(self, username=None, password=None, dry_run=True):
        """
        Initialize OSM uploader
        
        Args:
            username: OSM username
            password: OSM password
            dry_run: If True, don't actually upload (default: True)
        """
        self.dry_run = dry_run
        self.username = username
        self.password = password
        self.api = None
        
        if not dry_run:
            if not username or not password:
                raise ValueError("Username and password required for actual upload")
            
            # Initialize OSM API
            self.api = osmapi.OsmApi(username=username, password=password)
            print("Connected to OSM API")
        else:
            print("Running in DRY-RUN mode - no changes will be uploaded")
    
    def upload_element(self, element, changeset_comment="Add elevation data"):
        """
        Upload elevation tag for a single element
        
        Returns:
            Success status and message
        """
        element_type = element.get("type")
        element_id = element.get("id")
        tags = element.get("tags", {})
        
        if not element_id or not element_type:
            return False, "Missing element ID or type"
        
        # Check if elevation data exists
        if "ele" not in tags or "ele:source" not in tags:
            return False, "Missing elevation data in tags"
        
        ele_value = tags["ele"]
        
        if self.dry_run:
            # Dry run - just log what would be done
            print(f"[DRY-RUN] Would update {element_type} {element_id}:")
            print(f"  ele={ele_value}, ele:source=SRTM")
            return True, "Dry-run successful"
        
        try:
            # Get current element data from OSM
            if element_type == "node":
                osm_element = self.api.NodeGet(element_id)
            elif element_type == "way":
                osm_element = self.api.WayGet(element_id)
            else:
                return False, f"Unsupported element type: {element_type}"
            
            # Update tags
            if "tag" not in osm_element:
                osm_element["tag"] = {}
            
            osm_element["tag"]["ele"] = ele_value
            osm_element["tag"]["ele:source"] = "SRTM"
            
            # Create changeset and upload
            changeset = self.api.ChangesetCreate({
                "comment": changeset_comment,
                "created_by": "elevate-romania script"
            })
            
            if element_type == "node":
                self.api.NodeUpdate(osm_element)
            elif element_type == "way":
                self.api.WayUpdate(osm_element)
            
            self.api.ChangesetClose()
            
            print(f"Updated {element_type} {element_id} with ele={ele_value}")
            return True, "Upload successful"
            
        except Exception as e:
            return False, f"Error uploading: {str(e)}"
    
    def upload_elements(self, elements, category_name="elements"):
        """
        Upload elevation data for multiple elements
        
        Returns:
            Statistics dict
        """
        stats = {
            "total": len(elements),
            "successful": 0,
            "failed": 0,
            "errors": []
        }
        
        print(f"\nUploading {category_name}...")
        
        for i, element in enumerate(elements):
            success, message = self.upload_element(element)
            
            if success:
                stats["successful"] += 1
            else:
                stats["failed"] += 1
                stats["errors"].append({
                    "element_id": element.get("id"),
                    "error": message
                })
            
            if (i + 1) % 10 == 0:
                print(f"Processed {i + 1}/{len(elements)}...")
        
        return stats
    
    def upload_all(self, data):
        """Upload all validated data"""
        all_stats = {}
        
        for category, elements in data.items():
            if elements:
                stats = self.upload_elements(elements, category)
                all_stats[category] = stats
        
        return all_stats


if __name__ == "__main__":
    # Load validated data
    with open("output/osm_data_validated.json", "r", encoding="utf-8") as f:
        data = json.load(f)
    
    # Extract valid elements
    valid_data = {}
    for category, info in data.items():
        if "valid_elements" in info:
            valid_data[category] = info["valid_elements"]
    
    # Dry run upload
    print("Starting upload process (DRY-RUN mode)...")
    uploader = OSMUploader(dry_run=True)
    stats = uploader.upload_all(valid_data)
    
    # Display statistics
    print("\n" + "="*50)
    print("UPLOAD STATISTICS (DRY-RUN)")
    print("="*50)
    for category, category_stats in stats.items():
        print(f"\n{category}:")
        print(f"  Total: {category_stats['total']}")
        print(f"  Successful: {category_stats['successful']}")
        print(f"  Failed: {category_stats['failed']}")
    
    print("\n" + "="*50)
    print("To perform actual upload, run with dry_run=False")
    print("and provide OSM credentials")
