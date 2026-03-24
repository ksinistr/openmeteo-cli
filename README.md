# openmeteo-cli

A command-line weather tool built especially for AI agents.

## Overview

`openmeteo-cli` fetches weather forecast data from the [Open-Meteo API](https://open-meteo.com/) and outputs it in human-readable `toon` or machine-readable `json` formats.

The CLI is designed especially for AI agents and automation workflows. `toon` is the default because it stays structured while using fewer tokens than heavier text or JSON output in normal agent conversations.

## Installation For Humans And Agents

Give this install document to Clawbot or another coding agent:

```text
Install and configure openmeteo-cli by following the instructions here:
https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/install.md
```

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/install.sh | bash
```

The installer downloads the latest GitHub Release binary into `~/.local/bin/openmeteo-cli`.

### Clawbot Skill

```bash
curl -fsSL https://raw.githubusercontent.com/ksinistr/openmeteo-cli/main/skills/openmeteo-cli/install.sh | bash -s -- /path/to/clawbot/skills
```

See [install.md](./install.md) for the full install flow, custom install paths, version-pinned installs, and Clawbot setup details.

### From Source

```bash
git clone https://github.com/ksinistr/openmeteo-cli.git
cd openmeteo-cli
make build
```

The binary will be created at `bin/openmeteo-cli`.

## Commands

- `hourly` - Get hourly weather forecast (max 2 days)
- `daily` - Get daily weather forecast (max 14 days)

## Usage

```bash
# Hourly forecast (1-2 days)
openmeteo-cli hourly --latitude <latitude> --longitude <longitude> --forecast-days <1-2>

# Daily forecast (1-14 days)
openmeteo-cli daily --latitude <latitude> --longitude <longitude> --forecast-days <1-14>
```

## Help

Show help message:

```bash
openmeteo-cli -h
openmeteo-cli --help
openmeteo-cli hourly --help
openmeteo-cli daily --help
```

## Options

- `--latitude <float>` - Latitude coordinate (required, -90 to 90)
- `--longitude <float>` - Longitude coordinate (required, -180 to 180)
- `--forecast-days <int>` - Number of days to forecast (required)
  - For `hourly`: 1-2 days
  - For `daily`: 1-14 days
- `--units metric|imperial` - Units: metric (default) or imperial
- `--format toon|json` - Output format: toon (default) or json

## Examples

### Hourly forecast for New York City
```bash
openmeteo-cli hourly --latitude 40.7128 --longitude -74.0060 --forecast-days 1
```

### 2-day hourly forecast
```bash
openmeteo-cli hourly --latitude 40.7128 --longitude -74.0060 --forecast-days 2
```

### 7-day daily forecast with imperial units
```bash
openmeteo-cli daily --latitude 40.7128 --longitude -74.0060 --forecast-days 7 --units imperial
```

### 14-day daily forecast
```bash
openmeteo-cli daily --latitude 40.7128 --longitude -74.0060 --forecast-days 14
```

### JSON output for programmatic use
```bash
openmeteo-cli hourly --latitude 40.7128 --longitude -74.0060 --forecast-days 1 --format json
```

### Using different locations

**London, UK:**
```bash
openmeteo-cli hourly --latitude 51.5074 --longitude -0.1278 --forecast-days 1
openmeteo-cli daily --latitude 51.5074 --longitude -0.1278 --forecast-days 7
```

**Tokyo, Japan:**
```bash
openmeteo-cli hourly --latitude 35.6762 --longitude 139.6503 --forecast-days 1
```

**Sydney, Australia:**
```bash
openmeteo-cli daily --latitude -33.8688 --longitude 151.2093 --forecast-days 14 --units imperial
```

## Defaults

- Units: `metric`
- Format: `toon`

## Validation Rules

- `--forecast-days` is required (no default)
- `hourly` command: `--forecast-days` must be 1-2
- `daily` command: `--forecast-days` must be 1-14

## Output Format Notes

### TOON Format (default)
TOON is a compact, human-readable text format designed for structured data and low token usage in agent workflows. Key features:
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
  "hours": [
    ...
  ]
}
```

## Public Contract

### Command-Line Interface

#### Syntax

```bash
openmeteo-cli (hourly | daily) --latitude <float> --longitude <float> --forecast-days <int> [options]
```

#### Commands

- `hourly`: Get hourly weather forecast (1-2 days)
- `daily`: Get daily weather forecast (1-14 days)

#### Required Options

- `--latitude <float>`: Latitude coordinate (must be between -90 and 90)
- `--longitude <float>`: Longitude coordinate (must be between -180 and 180)
- `--forecast-days <int>`: Number of days (1-2 for hourly, 1-14 for daily)

#### Optional Options

- `--units metric|imperial`: Weather units (default: `metric`)
- `--format toon|json`: Output format (default: `toon`)

### Output Schema

#### Meta Object (all outputs)
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

#### Units Object (all outputs)
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

#### Hour Object (hourly output)
Hourly forecast entry.

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

#### Day Object (daily output)
Daily forecast entry.

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
- `sunrise`/`sunset` (daily): Full local ISO 8601 timestamps
- `time` (hourly rows): Local time as `HH:MM` (e.g., `14:30`)

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 2 | Invalid arguments (missing required options, invalid flag format) |
| 3 | Validation error (lat/lon out of range, invalid units/format, missing --hourly/--daily) |
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