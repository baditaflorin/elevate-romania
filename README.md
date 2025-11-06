# Elevație OSM România

Proiect pentru adăugarea automată a datelor de altitudine (elevație) pentru obiectele din OpenStreetMap din România.

## 📋 Scop

Adăugarea tag-ului `ele` (elevation) pentru:
- ✅ Toate stațiile de tren din România
- ✅ Toate unitățile de cazare (hoteluri, pensiuni, cabane montane etc.)
- 🎯 **Prioritate**: Cabane montane

## 🚀 Funcționalități

Acest proiect implementează un workflow complet automatizat:

1. **Extract**: Descarcă date din OpenStreetMap prin Overpass API
2. **Filter**: Identifică obiectele care nu au tag-ul `ele`
3. **Enrich**: Obține altitudinea din OpenTopoData API (SRTM 30m)
4. **Validate**: Verifică că altitudinea este în range-ul valid (0-2600m pentru România)
5. **Export**: Generează fișier CSV cu datele
6. **Upload**: Încarcă modificările în OSM prin API (cu mod dry-run pentru review)

## 📦 Instalare

### Cerințe
- Python 3.7+
- pip

### Pași

```bash
# Clonare repository
git clone https://github.com/baditaflorin/elevate-romania.git
cd elevate-romania

# Instalare dependențe
pip install -r requirements.txt
```

## 🎯 Utilizare

### Workflow complet (recomandat pentru prima rulare)

```bash
# Dry-run complet - nu modifică nimic în OSM
python main.py --all --dry-run

# Pentru a limita numărul de obiecte (testare)
python main.py --all --dry-run --limit 10
```

### Workflow pas cu pas

```bash
# 1. Extrage date din OSM
python main.py --extract

# 2. Filtrează obiectele fără elevație
python main.py --filter

# 3. Îmbogățește cu date de altitudine
python main.py --enrich

# 4. Validează datele
python main.py --validate

# 5. Export CSV
python main.py --export-csv

# 6. Upload în OSM (dry-run)
python main.py --upload --dry-run
```

### Upload real în OSM

**⚠️ ATENȚIE**: Verifică datele din CSV înainte de upload real!

```bash
# Upload real (necesită credențiale OSM)
python main.py --upload --username "your_username" --password "your_password"
```

## 📊 Output

Proiectul generează următoarele fișiere:

- `osm_data_raw.json` - Date brute din OSM
- `osm_data_filtered.json` - Date filtrate (fără ele)
- `osm_data_enriched.json` - Date îmbogățite cu altitudine
- `osm_data_validated.json` - Date validate
- `elevation_data.csv` - Export CSV final pentru review

### Format CSV

```csv
category,type,id,name,lat,lon,elevation,elevation_source,tourism,railway,osm_link
alpine_huts,node,123456,Cabana Padina,45.123,25.456,1850.0,SRTM,alpine_hut,,https://www.openstreetmap.org/node/123456
train_stations,node,234567,Gara Sinaia,45.234,25.567,850.0,SRTM,,station,https://www.openstreetmap.org/node/234567
```

## 🔧 Module

### `extract.py`
Extrage date din OpenStreetMap folosind Overpass API.

Queries:
- Stații de tren: `railway=station`, `railway=halt`
- Cazare: `tourism=hotel|guest_house|alpine_hut|chalet|hostel|motel`

### `filter.py`
Filtrează obiectele care nu au tag-ul `ele` și prioritizează cabane montane.

### `enrich.py`
Obține altitudinea de la OpenTopoData API (dataset SRTM 30m).
- Rate limiting: 1 secundă între request-uri
- Suport pentru OpenTopoData și Open-Elevation

### `validate.py`
Validează că altitudinea este în range-ul valid:
- Minimum: 0m (Marea Neagră)
- Maximum: 2600m (Vârful Moldoveanu ~2544m)

### `upload.py`
Upload în OSM folosind `osmapi`:
- Mod dry-run pentru preview
- Adaugă tag-uri: `ele=XXX`, `ele:source=SRTM`
- Gestionare changeset-uri

### `csv_export.py`
Export date în format CSV pentru review manual.

### `main.py`
Script principal de orchestrare cu CLI.

## 🎨 Exemple

### Testare rapidă pe 5 obiecte

```bash
python main.py --extract --filter --enrich --validate --export-csv --limit 5
```

### Procesare doar cabane montane

După extragere și filtrare, editează `osm_data_filtered.json` să conțină doar categoria `alpine_huts`, apoi:

```bash
python main.py --enrich --validate --export-csv
```

### Review înainte de upload

```bash
# 1. Procesează datele
python main.py --all --dry-run

# 2. Verifică elevation_data.csv

# 3. Dry-run upload pentru preview
python main.py --upload --dry-run

# 4. Upload real (doar după verificare manuală!)
python main.py --upload --username "user" --password "pass"
```

## 📝 Notițe importante

1. **Rate Limiting**: API-urile folosite au limite de request-uri
   - OpenTopoData: folosește rate limiting de 1s între request-uri
   - Overpass API: folosește timeout de 300s

2. **Validare date**: Verifică întotdeauna CSV-ul înainte de upload!

3. **Changeset OSM**: Fiecare upload creează un changeset cu:
   - Comment: "Add elevation data"
   - Created by: "elevate-romania script"
   - Source: tag `ele:source=SRTM`

4. **Prioritate cabane**: Cabane montane (`tourism=alpine_hut`) sunt procesate primele

## 🤝 Contribuții

Contribuțiile sunt binevenite! Pentru bug-uri sau feature requests, deschide un issue.

## 📜 Licență

MIT License - vezi fișierul LICENSE

## 🔗 Link-uri utile

- [OpenStreetMap](https://www.openstreetmap.org)
- [Overpass API](https://overpass-api.de/)
- [OpenTopoData](https://www.opentopodata.org/)
- [OSM Wiki - Key:ele](https://wiki.openstreetmap.org/wiki/Key:ele)
- [osmapi Python library](https://github.com/metaodi/osmapi)