# openmeteo-cli

A command-line tool for fetching weather forecasts from Open-Meteo.

## Overview

`openmeteo-cli` fetches weather forecast data from the Open-Meteo API and outputs it in human-readable (TOON) or machine-readable (JSON) formats.

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd openmeteo-cli

# Build the binary
make build

# The binary will be created at bin/openmeteo-cli
```

### Using go install

```bash
go install openmeteo-cli@latest
```

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

## Help

Show help message for the binary or a specific command:

```bash
openmeteo-cli -h
openmeteo-cli --help
openmeteo-cli today -h
openmeteo-cli week --help
```

## Options

- `--lat <float>` - Latitude coordinate (required, -90 to 90)
- `--lon <float>` - Longitude coordinate (required, -180 to 180)
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

### Using different locations

**London, UK:**
```bash
openmeteo-cli today --lat 51.5074 --lon -0.1278
openmeteo-cli week --lat 51.5074 --lon -0.1278 --units metric
```

**Tokyo, Japan:**
```bash
openmeteo-cli today --lat 35.6762 --lon 139.6503
```

**Sydney, Australia:**
```bash
openmeteo-cli week --lat -33.8688 --lon 151.2093 --units imperial
```

### Getting hourly forecast for a past or future date
```bash
openmeteo-cli day --lat 34.0522 --lon -118.2437 --date 2026-04-15
```

## Defaults

- Units: `metric`
- Format: `toon`

## Output Format Notes

### TOON Format (default)
TOON is a compact, human-readable text format designed for structured data. Key features:
- Uses `#` prefix for section headers
- Key-value pairs on separate lines with `  ` (two-space) indentation
- Numeric values remain numeric (not quoted strings)
- Next line after a table header, not on the same line

Example:
```
# meta
  generated_at 2026-03-21T14:30
  timezone Europe/London
  latitude 51.5074
  longitude -0.1278
# units
  temperature C
  humidity %
  wind_speed km/h
  ...
```

### JSON Format
Standard JSON output with full pretty-printing. Use for programmatic consumption or debugging.

Example:
```json
{
  "meta": {
    "generated_at": "2026-03-21T14:30:00Z",
    "units": {
      "temperature": "C",
      "humidity": "%",
      ...
    },
    ...
  },
  "current": {
    ...
  }
}
```

## Public Contract

### Command-Line Interface

#### Syntax

```bash
openmeteo-cli <command> --lat <float> --lon <float> [options]
```

#### Commands

| Command | Description | Required Options | Optional Options |
|---------|-------------|------------------|------------------|
| `today` | Get today's forecast with hourly rows | `--lat`, `--lon` | `--units`, `--format` |
| `day` | Get forecast for a specific date | `--lat`, `--lon`, `--date` | `--units`, `--format` |
| `week` | Get 7-day forecast | `--lat`, `--lon` | `--units`, `--format` |

#### Required Options

- `--lat <float>`: Latitude coordinate (must be between -90 and 90)
- `--lon <float>`: Longitude coordinate (must be between -180 and 180)
- `--date YYYY-MM-DD`: Required only for the `day` command

#### Optional Options

- `--units metric|imperial`: Weather units (default: `metric`)
- `--format toon|json`: Output format (default: `toon`)

### Output Schema

All commands return a consistent structure with command-specific data.

#### Meta Object (all commands)
Metadata about the forecast generation and location.

```json
{
  "generated_at": "2026-03-21T14:30:00Z",
  "units": { ... },
  "timezone": "Europe/London",
  "latitude": 51.5074,
  "longitude": -0.1278
}
```

#### Units Object (all commands)
Units for all numeric fields in the response.

```json
{
  "temperature": "C",
  "humidity": "%",
  "wind_speed": "km/h",
  "wind_direction": "deg",
  "precipitation": "mm",
  "precipitation_probability": "%",
  "uv_index": "index"
}
```

#### Current Object (today command)
Current weather conditions.

```json
{
  "time": "14:30",
  "weather": "Mainly clear",
  "temperature": 12.5,
  "apparent_temperature": 10.2,
  "humidity": 65,
  "precipitation": 0.0,
  "precipitation_probability": 10,
  "wind_speed": 15.3,
  "wind_gusts": 22.1,
  "wind_direction": 245,
  "uv_index": 3.2
}
```

#### Day Object (day command)
Daily summary for the requested date.

```json
{
  "date": "2026-03-22",
  "weather": "Partly cloudy",
  "temp_min": 8.1,
  "temp_max": 16.3,
  "precipitation_sum": 2.5,
  "precipitation_probability_max": 30,
  "wind_speed_max": 18.7,
  "wind_gusts_max": 25.4,
  "uv_index_max": 4.1,
  "sunrise": "2026-03-22T06:30:00Z",
  "sunset": "2026-03-22T19:45:00Z"
}
```

#### Hours Array (today and day commands)
Hourly forecast rows for the requested local date.

```json
{
  "time": "14:00",
  "weather": "Mainly clear",
  "temperature": 13.1,
  "apparent_temperature": 10.8,
  "humidity": 62,
  "precipitation": 0.0,
  "precipitation_probability": 10,
  "wind_speed": 14.5,
  "wind_gusts": 21.3,
  "wind_direction": 242,
  "uv_index": 3.5
}
```

#### Days Array (week command)
7 daily forecast rows starting from the current local date.

```json
{
  "date": "2026-03-22",
  "weather": "Partly cloudy",
  "temp_min": 8.1,
  "temp_max": 16.3,
  "precipitation_sum": 2.5,
  "precipitation_probability_max": 30,
  "wind_speed_max": 18.7,
  "wind_gusts_max": 25.4,
  "uv_index_max": 4.1,
  "sunrise": "2026-03-22T06:30:00Z",
  "sunset": "2026-03-22T19:45:00Z"
}
```

### Weather Code Mapping

All numeric weather codes are mapped to human-readable descriptions:

| Code | Description |
|------|-------------|
| 0 | Clear sky |
| 1 | Mainly clear |
| 2 | Partly cloudy |
| 3 | Overcast |
| 45 | Fog |
| 48 | Depositing rime fog |
| 51 | Light drizzle |
| 53 | Moderate drizzle |
| 55 | Dense drizzle |
| 56 | Light freezing drizzle |
| 57 | Dense freezing drizzle |
| 61 | Slight rain |
| 63 | Moderate rain |
| 65 | Heavy rain |
| 66 | Light freezing rain |
| 67 | Heavy freezing rain |
| 71 | Slight snow fall |
| 73 | Moderate snow fall |
| 75 | Heavy snow fall |
| 77 | Snow grains |
| 80 | Slight rain showers |
| 81 | Moderate rain showers |
| 82 | Violent rain showers |
| 85 | Slight snow showers |
| 86 | Heavy snow showers |
| 95 | Thunderstorm |
| 96 | Thunderstorm with slight hail |
| 99 | Thunderstorm with heavy hail |

Unknown codes are reported as `Unknown weather code: <code>`.

### Time Formatting

- `generated_at`: Full ISO 8601 timestamp (e.g., `2026-03-21T14:30:00Z`)
- `sunrise`/`sunset` (day/week): Full local ISO 8601 timestamps
- `time` (hourly rows): Local time as `HH:MM` (e.g., `14:30`)

### Date Filtering Rules

- `today`: Returns hourly rows for the current local forecast date based on the location's timezone
- `day`: Returns hourly rows for the requested local date
- `week`: Returns exactly 7 consecutive daily rows starting from the current local date

**Note:** DST transitions may result in 23 or 25 hourly rows for a local day instead of 24.

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 2 | Invalid arguments (missing required options, invalid flag format) |
| 3 | Validation error (lat/lon out of range, invalid units/format, invalid date format) |
| 4 | Upstream API error (network issue, Open-Meteo API error) |
| 5 | Requested date unavailable (date outside forecast range) |
| 6 | Output encoding error |

## Development

### Prerequisites

- Go 1.21 or higher
- GNU Make (optional, for Makefile targets)

### Build

```bash
make build
```

The binary will be created at `bin/openmeteo-cli`.

### Test

```bash
make test
```

### Format

```bash
make fmt
```

### Lint

```bash
make lint
```

### Clean

```bash
make clean
```

### All (build + test)

```bash
make all
```

## License

This project is licensed under the MIT License.
