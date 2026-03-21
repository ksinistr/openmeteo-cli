# openmeteo-cli

A command-line tool for fetching weather forecasts from Open-Meteo.

## Overview

`openmeteo-cli` fetches weather forecast data from the Open-Meteo API and outputs it in human-readable or machine-readable formats.

## Commands

- `today` - Get today's weather forecast with hourly rows
- `day` - Get weather forecast for a specific date
- `week` - Get a 7-day weather forecast

## Usage

```bash
openmeteo-cli today --lat <latitude> --lon <longitude> [options]
openmeteo-cli day --lat <latitude> --lon <longitude> --date YYYY-MM-DD [options]
openmeteo-cli week --lat <latitude> --lon <longitude> [options]
```

## Options

- `--lat <float>` - Latitude coordinate (required)
- `--lon <float>` - Longitude coordinate (required)
- `--date YYYY-MM-DD` - Date for the day command (required for `day` command)
- `--units metric|imperial` - Units: metric (default) or imperial
- `--format toon|json` - Output format: toon (default) or json

## Examples

### Today's weather for New York City
```bash
openmeteo-cli today --lat 40.7128 --lon -74.0060
```

### Weather for a specific date
```bash
openmeteo-cli day --lat 40.7128 --lon -74.0060 --date 2026-03-22
```

### 7-day forecast with imperial units
```bash
openmeteo-cli week --lat 40.7128 --lon -74.0060 --units imperial
```

### JSON output for debugging
```bash
openmeteo-cli today --lat 40.7128 --lon -74.0060 --format json
```

## Defaults

- Units: `metric`
- Format: `toon`

## Output Schema

The output includes:
- `meta` - Metadata including generated timestamp, units, timezone, and coordinates
- `current` or `day` or `days` - Main forecast data
- `hours` or `hours` - Hourly forecast entries
- Weather descriptions (human-readable, not numeric codes)

## Exit Codes

- `2` - Invalid arguments
- `3` - Validation error
- `4` - Upstream API error
- `5` - Requested date unavailable
- `6` - Output encoding error

## Development

### Build
```bash
make build
```

### Test
```bash
make test
```

### Format
```bash
make fmt
```

## License

This project is licensed under the MIT License.
