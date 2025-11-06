# Quick Start Guide - Elevație OSM România

## 🚀 Getting Started in 5 Minutes

### 1. Install Dependencies

```bash
pip install -r requirements.txt
```

### 2. Run Demo (Recommended First Step)

```bash
python demo.py
```

This will:
- Create sample OSM data
- Process it through all pipeline steps
- Generate demo CSV output
- Show dry-run upload preview

### 3. Try the Real Workflow

#### Option A: Full Pipeline (Dry-Run)
```bash
python main.py --all --dry-run --limit 10
```

This will:
1. Extract 10 items from OSM
2. Filter those without elevation
3. Fetch elevation from OpenTopoData
4. Validate the data
5. Export to CSV
6. Show what would be uploaded (dry-run)

#### Option B: Step by Step

```bash
# Extract data
python main.py --extract

# Filter
python main.py --filter

# Enrich (limit to 5 for testing)
python main.py --enrich --limit 5

# Validate
python main.py --validate

# Export to CSV
python main.py --export-csv

# Preview upload
python main.py --upload --dry-run
```

### 4. Review the Output

Check `elevation_data.csv` to see the results:
- Verify elevations are reasonable
- Check OSM links work
- Ensure priority items (alpine huts) are included

### 5. Real Upload (After Review!)

⚠️ **Only after verifying the CSV data!**

```bash
python main.py --upload --username "your_osm_username" --password "your_password"
```

## 📊 Expected Output

After running the demo or full pipeline:

```
osm_data_raw.json       - Raw data from OSM
osm_data_filtered.json  - Filtered (no elevation)
osm_data_enriched.json  - With elevation data added
osm_data_validated.json - Validated data
elevation_data.csv      - CSV for review
```

## ⚠️ Important Notes

1. **Rate Limits**: OpenTopoData has rate limits. The script waits 1 second between requests.
2. **Start Small**: Use `--limit 10` for testing
3. **Always Review**: Check CSV before real upload
4. **Dry-Run First**: Always run `--dry-run` before actual upload
5. **Priorities**: Alpine huts are processed first

## 🐛 Troubleshooting

### "ModuleNotFoundError"
```bash
pip install -r requirements.txt
```

### "osm_data_raw.json not found"
Run extract first:
```bash
python main.py --extract
```

### API Rate Limit Errors
- Reduce batch size with `--limit`
- Wait a few minutes
- Consider running in smaller batches

## 📞 Need Help?

- Check `README.md` for full documentation
- Review output CSV carefully before upload
- Test with `--limit 5` first

## ✅ Checklist for Real Upload

- [ ] Ran demo successfully
- [ ] Tested with `--limit 10`
- [ ] Reviewed CSV output
- [ ] Verified elevations are reasonable (0-2600m for Romania)
- [ ] Ran `--upload --dry-run`
- [ ] Have OSM credentials ready
- [ ] Ready for real upload!
