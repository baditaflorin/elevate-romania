# Architecture Documentation

## Overview

This document describes the modular architecture of the elevate-romania application, following DRY (Don't Repeat Yourself) and SOLID principles.

## Design Principles

### 1. Single Responsibility Principle (SRP)
Each module has a single, well-defined responsibility:
- **coordinates.go**: Coordinate extraction and validation
- **element_categorizer.go**: Element type categorization
- **config.go**: Configuration management
- **logger.go**: Structured logging
- **factory.go**: Object creation and dependency injection
- **http_client.go**: HTTP operations with retry logic
- **errors.go**: Error handling utilities

### 2. Open/Closed Principle (OCP)
- Configuration values are externalized, making the system open for extension
- Interfaces allow for different implementations without modifying existing code
- Factory pattern enables adding new API clients without changing existing code

### 3. Liskov Substitution Principle (LSP)
- All implementations can be substituted through interfaces
- Logger, Config, and HTTP clients are all interface-based

### 4. Interface Segregation Principle (ISP)
- Small, focused interfaces defined in `interfaces.go`
- Clients depend only on methods they use

### 5. Dependency Inversion Principle (DIP)
- High-level modules depend on abstractions (interfaces)
- Factory pattern provides proper dependency injection
- Configuration is injected rather than hard-coded

## Module Structure

### Core Utilities

#### coordinates.go
- **Purpose**: Centralized coordinate handling
- **Key Types**: `Coordinates`, `CoordinateExtractor`
- **Benefits**: Eliminates duplicate coordinate extraction logic (used in 3+ places)

#### element_categorizer.go
- **Purpose**: Element type classification and filtering
- **Key Types**: `ElementCategorizer`, `ElementCategory`
- **Benefits**: Single source of truth for element categorization logic

#### config.go
- **Purpose**: Configuration management with type-safe accessors
- **Key Types**: `Config`
- **Features**:
  - Environment variable loading
  - Type-safe getters (GetInt, GetFloat, GetBool)
  - Default value support
  - Configuration validation

#### logger.go
- **Purpose**: Structured logging interface
- **Key Types**: `SimpleLogger`, `Logger` (interface)
- **Benefits**: Consistent logging across the application

#### factory.go
- **Purpose**: Object creation and dependency injection
- **Key Types**: `APIClientFactory`
- **Benefits**: 
  - Centralized configuration
  - Easy testing through dependency injection
  - Consistent object creation

#### http_client.go
- **Purpose**: HTTP operations with retry logic
- **Key Types**: `HTTPClientWrapper`, `RetryConfig`
- **Features**:
  - Automatic retry with exponential backoff
  - Configurable retry behavior
  - Error handling and logging

#### errors.go
- **Purpose**: Error handling utilities
- **Key Types**: `ErrorContext`
- **Features**:
  - Error wrapping with context
  - Structured error information

### Interfaces

All major components have well-defined interfaces in `interfaces.go`:
- `ElevationProvider`
- `BatchElevationProvider`
- `DataExtractor`
- `ElementFilter`
- `ElementValidator`
- `HTTPClient`
- `Logger`
- `ConfigProvider`

## Configuration

### Environment Variables

The application supports the following environment variables:

#### OAuth Configuration
- `OSM_CLIENT_ID`: OpenStreetMap OAuth client ID
- `OSM_CLIENT_SECRET`: OpenStreetMap OAuth client secret
- `OSM_ACCESS_TOKEN`: OpenStreetMap OAuth access token

#### API Configuration
- `OVERPASS_URL`: Overpass API endpoint (default: https://overpass-api.de/api/interpreter)
- `OPENTOPO_URL`: OpenTopoData API endpoint (default: https://api.opentopodata.org/v1/srtm30m)
- `OSM_API_URL`: OpenStreetMap API endpoint (default: https://api.openstreetmap.org/api/0.6)

#### Performance Configuration
- `API_RATE_LIMIT_MS`: Rate limit between API requests in milliseconds (default: 1000)
- `BATCH_SIZE`: Batch size for batch API requests (default: 100)
- `API_TIMEOUT_SEC`: HTTP request timeout in seconds (default: 30)

#### OAuth Flow
- `OAUTH_REDIRECT_URI`: OAuth redirect URI (default: http://127.0.0.1:8080/callback)

## Testing

### Test Coverage

The refactored code includes comprehensive test coverage:

1. **coordinates_test.go**: 
   - Coordinate validation
   - Coordinate extraction from elements
   - Multi-element coordinate extraction

2. **element_categorizer_test.go**:
   - Element categorization
   - Category-specific checks
   - Elevation presence detection
   - Batch categorization

3. **config_test.go**:
   - Configuration get/set operations
   - Type conversions (int, float, bool)
   - Default value handling
   - Validation logic

4. **osm_api_test.go**:
   - Tag merging logic
   - OAuth configuration save/load

5. **batch_enricher_test.go**:
   - Batch size validation
   - Batch processing logic

### Running Tests

```bash
cd src
go test -v ./...
```

## Migration from Old Code

### Before (DRY Violations)

Coordinate extraction was repeated in multiple files:

```go
// In filter.go
func (f *ElevationFilter) getCoordinates(element OSMElement) (float64, float64, bool) {
    if element.Type == "node" {
        if element.Lat != 0 && element.Lon != 0 {
            return element.Lat, element.Lon, true
        }
    } else if element.Type == "way" && element.Center != nil {
        if element.Center.Lat != 0 && element.Center.Lon != 0 {
            return element.Center.Lat, element.Center.Lon, true
        }
    }
    return 0, 0, false
}

// Similar code in enrich.go and batch_enricher.go
```

### After (DRY Compliant)

Single coordinate extractor used everywhere:

```go
// In coordinates.go
extractor := NewCoordinateExtractor()
coords, valid := extractor.Extract(element)
if valid {
    // Use coords.Lat and coords.Lon
}
```

### Hard-coded Values Before

```go
e.BaseURL = "https://api.opentopodata.org/v1/srtm30m"
batchEnricher := NewBatchElevationEnricher("opentopo", 1000.0, 100)
```

### Configuration-driven After

```go
config := NewConfig()
config.LoadFromEnv()
factory := NewAPIClientFactory(config, logger)
batchEnricher := factory.CreateBatchElevationEnricher("opentopo")
```

## Benefits Achieved

### 1. Code Reusability
- Coordinate extraction logic: 1 implementation (was 3+)
- Element categorization logic: 1 implementation (was 2+)
- Configuration management: Centralized
- HTTP client logic: Reusable with retry

### 2. Maintainability
- Changes to coordinate logic only need to be made once
- Configuration changes don't require code changes
- Clear separation of concerns

### 3. Testability
- All utilities have comprehensive unit tests
- Mock interfaces enable easy testing
- Factory pattern enables dependency injection

### 4. Production Readiness
- Structured logging for better observability
- Retry logic for API resilience
- Configuration validation
- Error context for better debugging

### 5. SOLID Compliance
- **S**: Each module has single responsibility
- **O**: Open for extension through configuration
- **L**: Interface-based design enables substitution
- **I**: Small, focused interfaces
- **D**: Dependencies injected through factory

## Future Enhancements

Potential areas for further improvement:
1. Add metrics/observability
2. Implement circuit breaker pattern for API calls
3. Add request/response caching
4. Implement more sophisticated rate limiting
5. Add distributed tracing support
