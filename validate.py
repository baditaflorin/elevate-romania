"""
Validate module - Validate elevation data
"""
import json


class ElevationValidator:
    """Validate elevation data"""
    
    def __init__(self, min_elevation=0, max_elevation=2600):
        """
        Initialize validator
        
        Args:
            min_elevation: Minimum valid elevation in meters (Romania: 0 = Black Sea)
            max_elevation: Maximum valid elevation in meters (Romania: ~2544m Moldoveanu)
        """
        self.min_elevation = min_elevation
        self.max_elevation = max_elevation
    
    def validate_element(self, element):
        """
        Validate elevation for a single element
        
        Returns:
            dict with validation results
        """
        result = {
            "valid": False,
            "element_id": element.get("id"),
            "element_type": element.get("type"),
            "elevation": None,
            "errors": []
        }
        
        # Check if elevation exists
        if "elevation_fetched" not in element:
            result["errors"].append("No elevation data")
            return result
        
        elevation = element["elevation_fetched"]
        result["elevation"] = elevation
        
        # Validate range
        if elevation < self.min_elevation:
            result["errors"].append(f"Elevation {elevation}m below minimum {self.min_elevation}m")
        elif elevation > self.max_elevation:
            result["errors"].append(f"Elevation {elevation}m above maximum {self.max_elevation}m")
        else:
            result["valid"] = True
        
        return result
    
    def validate_elements(self, elements):
        """
        Validate multiple elements
        
        Returns:
            dict with valid and invalid elements
        """
        valid = []
        invalid = []
        
        for element in elements:
            validation = self.validate_element(element)
            
            if validation["valid"]:
                valid.append(element)
            else:
                invalid.append({
                    "element": element,
                    "validation": validation
                })
        
        return {
            "valid": valid,
            "invalid": invalid
        }
    
    def validate_all(self, data):
        """Validate all categories of data"""
        results = {}
        
        for category, elements in data.items():
            if elements:
                print(f"\nValidating {category}...")
                validation = self.validate_elements(elements)
                results[category] = validation
                
                print(f"  Valid: {len(validation['valid'])}")
                print(f"  Invalid: {len(validation['invalid'])}")
                
                # Show invalid examples
                if validation['invalid']:
                    print(f"  Invalid examples:")
                    for item in validation['invalid'][:3]:
                        val = item['validation']
                        print(f"    - ID {val['element_id']}: {', '.join(val['errors'])}")
        
        return results


if __name__ == "__main__":
    # Load enriched data
    with open("output/osm_data_enriched.json", "r", encoding="utf-8") as f:
        data = json.load(f)
    
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
    
    with open("output/osm_data_validated.json", "w", encoding="utf-8") as f:
        json.dump(output, f, indent=2, ensure_ascii=False)
    
    print("\nValidation complete! Results saved to output/osm_data_validated.json")
