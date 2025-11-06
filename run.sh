#!/bin/bash
# Convenience script to run the application with Docker

set -e

# Build image if not exists
if [[ "$(docker images -q elevate-romania:latest 2> /dev/null)" == "" ]]; then
  echo "Building Docker image..."
  docker compose build
fi

# Run with provided arguments
echo "Running elevate-romania in Docker..."
docker compose run --rm elevate-romania "$@"
