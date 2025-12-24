# asciigis
 a terminal-based viewer for vector geospatial data.

## Usage

```bash
# start with a path
go run ./cmd/asciigis /path/to/data.geojson

# start with a fixed canvas size
go run ./cmd/asciigis -W 60 -H 20 /path/to/data.geojson

# start without a path, then type it in the UI
go run ./cmd/asciigis
```

### Keys

- `q` / `Ctrl+C`: quit
- `r`: reload
- `a` / `d`: canvas width -/+ 
- `w` / `s`: canvas height +/-
- `/` or `p`: set GeoJSON path
- (path input) `Enter`: load, `Esc`: cancel, `Ctrl+U`: clear
