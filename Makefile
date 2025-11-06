.PHONY: help build demo run clean

help: ## Show this help message
	@echo "Elevație OSM România - Docker Commands"
	@echo "======================================"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build Docker image
	docker compose build

demo: ## Run demo with sample data
	docker compose run --rm --entrypoint "" elevate-romania python demo.py

run: ## Run with arguments (use ARGS="--help")
	docker compose run --rm elevate-romania $(ARGS)

extract: ## Extract data from OSM
	docker compose run --rm elevate-romania --extract

filter: ## Filter data without elevation
	docker compose run --rm elevate-romania --filter

enrich: ## Enrich with elevation (use LIMIT=10)
	docker compose run --rm elevate-romania --enrich --limit $(or $(LIMIT),10)

validate: ## Validate elevation data
	docker compose run --rm elevate-romania --validate

export: ## Export to CSV
	docker compose run --rm elevate-romania --export-csv

dry-run: ## Complete dry-run workflow (use LIMIT=10)
	docker compose run --rm elevate-romania --all --dry-run --limit $(or $(LIMIT),10)

upload-dry: ## Preview upload changes
	docker compose run --rm elevate-romania --upload --dry-run

clean: ## Remove output files and Docker image
	rm -rf output/*
	docker compose down --rmi all

clean-output: ## Remove only output files
	rm -rf output/*

shell: ## Open shell in container
	docker compose run --rm -it elevate-romania /bin/bash

logs: ## Show container logs
	docker compose logs -f
