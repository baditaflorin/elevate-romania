#!/bin/bash
# Run the demo in Docker

set -e

echo "Building Docker image..."
docker compose build

echo "Running demo..."
docker compose run --rm elevate-romania python demo.py

echo ""
echo "Demo complete! Check the output directory for generated files."
