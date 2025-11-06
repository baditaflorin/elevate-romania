.PHONY: help build run clean test demo

help: ## Show this help message
	@echo "Elevație OSM România - Go Edition"
	@echo "=================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o elevate-romania .

run: ## Run with arguments (use ARGS="--help")
	go run . $(ARGS)

extract: build ## Extract data from OSM
	./elevate-romania --extract

filter: build ## Filter data without elevation
	./elevate-romania --filter

enrich: build ## Enrich with elevation (use LIMIT=10)
	./elevate-romania --enrich --limit $(or $(LIMIT),10)

validate: build ## Validate elevation data
	./elevate-romania --validate

export: build ## Export to CSV
	./elevate-romania --export-csv

dry-run: build ## Complete dry-run workflow (use LIMIT=10)
	./elevate-romania --all --dry-run --limit $(or $(LIMIT),10)

upload-dry: build ## Preview upload changes
	./elevate-romania --upload --dry-run

test: build ## Run test workflow with 5 items
	./elevate-romania --all --dry-run --limit 5

demo: build ## Run demo with dry-run
	./elevate-romania --all --dry-run --limit 10

clean: ## Remove binary and output files
	rm -f elevate-romania elevate-romania.exe
	rm -rf output/

clean-output: ## Remove only output files
	rm -rf output/

deps: ## Download dependencies
	go mod download
	go mod tidy

install: build ## Install binary to GOPATH
	go install .
