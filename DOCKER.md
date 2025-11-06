# Docker Usage Guide - Elevație OSM România

## 🐳 Zero Python Installation Required!

**This application is fully Dockerized - you NEVER need to install Python, pip, or any Python packages!**

All you need is Docker. The application runs completely inside a container with all dependencies pre-installed.

## Prerequisites

- Docker installed ([Get Docker](https://docs.docker.com/get-docker/))
- Docker Compose installed (usually comes with Docker Desktop)

That's it! No Python, no pip, no virtual environments.

## 🚀 Quick Start (3 Steps)

### 1. Build the Docker Image

```bash
docker compose build
```

This downloads Python, installs all dependencies, and packages everything - **you never touch Python directly**.

### 2. Run the Demo

**Option A: Using Makefile (easiest)**
```bash
make demo
```

**Option B: Using shell script**
```bash
./demo.sh
```

**Option C: Using docker compose directly**
```bash
docker compose run --rm elevate-romania python demo.py
```

### 3. Run the Full Pipeline

**Option A: Using Makefile (easiest)**
```bash
make dry-run LIMIT=10
```

**Option B: Using shell script**
```bash
./run.sh --all --dry-run --limit 10
```

**Option C: Using docker compose directly**
```bash
docker compose run --rm elevate-romania --all --dry-run --limit 10
```

## 📖 Common Usage Patterns

### Using Makefile (Recommended)

The easiest way to use the application:

```bash
# See all available commands
make help

# Build image
make build

# Run demo
make demo

# Extract data
make extract

# Filter data
make filter

# Enrich with elevation (process 10 items)
make enrich LIMIT=10

# Complete dry-run workflow
make dry-run LIMIT=5

# View output files
ls -lh output/
```

### Using Shell Scripts

Convenient wrapper scripts:

```bash
# Run demo
./demo.sh

# Run any command
./run.sh --help
./run.sh --extract
./run.sh --all --dry-run --limit 10
```

### Show Help

```bash
./run.sh --help
```

### Extract Data from OSM

```bash
./run.sh --extract
```

Output files will be saved in the `output/` directory on your host machine.

### Filter Data

```bash
./run.sh --filter
```

### Enrich with Elevation (Limited)

```bash
./run.sh --enrich --limit 5
```

### Complete Workflow

```bash
# Dry-run mode (safe, doesn't upload to OSM)
./run.sh --all --dry-run --limit 10

# Check the output directory for CSV file
ls -lh output/
cat output/elevation_data.csv
```

### Real Upload to OSM

**⚠️ Only after reviewing the CSV!**

```bash
./run.sh --upload --username "your_osm_username" --password "your_password"
```

Or using environment variables:

```bash
export OSM_USERNAME="your_username"
export OSM_PASSWORD="your_password"
./run.sh --upload --username "$OSM_USERNAME" --password "$OSM_PASSWORD"
```

## 🔧 Advanced Usage

### Run Specific Steps

```bash
# Extract
./run.sh --extract

# Filter
./run.sh --filter

# Enrich
./run.sh --enrich --limit 20

# Validate
./run.sh --validate

# Export CSV
./run.sh --export-csv

# Upload (dry-run)
./run.sh --upload --dry-run
```

### Custom Docker Commands

If you need more control:

```bash
# Run with specific output directory
docker run --rm -v $(pwd)/output:/app/output elevate-romania:latest --all --dry-run

# Run interactively (for debugging)
docker run --rm -it -v $(pwd)/output:/app/output elevate-romania:latest /bin/bash

# Inside container, run commands manually:
python main.py --help
python demo.py
```

### Using docker compose.yml Directly

Edit `docker compose.yml` to set default command:

```yaml
services:
  elevate-romania:
    # ... other settings ...
    command: ["--all", "--dry-run", "--limit", "10"]
```

Then just run:
```bash
docker compose up
```

## 📁 File Locations

### Inside Container
- Application code: `/app/`
- Output directory: `/app/output/`

### On Host (Your Computer)
- Output files: `./output/` (automatically created)
- All generated CSVs and JSON files appear here

## 🔄 Rebuilding After Changes

If you modify any Python files or dependencies:

```bash
docker compose build --no-cache
```

## 🧹 Cleanup

### Remove Generated Files

```bash
rm -rf output/
```

### Remove Docker Image

```bash
docker compose down --rmi all
```

Or:
```bash
docker rmi elevate-romania:latest
```

## 🐛 Troubleshooting

### Permission Issues with Output Directory

```bash
# Create output directory with correct permissions
mkdir -p output
chmod 777 output
```

### Container Won't Start

```bash
# Rebuild without cache
docker compose build --no-cache

# Check logs
docker compose logs
```

### "Command not found" for ./run.sh

```bash
# Make scripts executable
chmod +x *.sh

# Or run directly
bash run.sh --help
```

## 📊 Complete Example Workflow

```bash
# 1. Build image (one time)
docker compose build

# 2. Run demo to test
./demo.sh

# 3. Run limited extraction and enrichment
./run.sh --all --dry-run --limit 5

# 4. Check results
ls -lh output/
cat output/elevation_data.csv

# 5. If results look good, run full extraction
./run.sh --extract
./run.sh --filter
./run.sh --enrich --limit 50  # Process 50 items
./run.sh --validate
./run.sh --export-csv

# 6. Review CSV carefully!
cat output/elevation_data.csv

# 7. Upload (after review)
./run.sh --upload --dry-run  # Preview first
./run.sh --upload --username "user" --password "pass"  # Real upload
```

## ✅ Benefits of Docker Approach

- ✅ No Python installation needed
- ✅ No dependency management
- ✅ Consistent environment across all systems
- ✅ Easy to share and deploy
- ✅ Isolated from your system
- ✅ Works on Windows, Mac, Linux identically

## 🔗 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- See `README.md` for Python-based usage (if you must)
- See `QUICKSTART.md` for quick reference

---

**No Python knowledge required!** Just use the shell scripts and Docker commands above.
