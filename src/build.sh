#!/bin/bash
# Build and run script for Elevație OSM România

set -e

echo "Elevație OSM România - Go Implementation"
echo "========================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go 1.21 or higher from https://golang.org/dl/"
    exit 1
fi

echo "Go version:"
go version
echo ""

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies ready"
echo ""

# Build
echo "Building binary..."
go build -o elevate-romania .
echo "✓ Build successful"
echo ""

# Show usage
echo "Binary created: ./elevate-romania"
echo ""
echo "Quick start:"
echo "  ./elevate-romania --help"
echo "  ./elevate-romania --all --dry-run --limit 10"
echo ""
echo "Next steps:"
echo "  1. Run extraction: ./elevate-romania --extract"
echo "  2. Filter data: ./elevate-romania --filter"
echo "  3. Enrich (test): ./elevate-romania --enrich --limit 10"
echo "  4. Validate: ./elevate-romania --validate"
echo "  5. Export CSV: ./elevate-romania --export-csv"
echo "  6. Upload (dry-run): ./elevate-romania --upload --dry-run"
echo ""
