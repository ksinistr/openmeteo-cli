# openmeteo-cli

A command-line weather tool built especially for AI agents.

## Overview

`openmeteo-cli` fetches weather forecast data from the [Open-Meteo API](https://open-meteo.com/) and outputs it in human-readable `toon` or machine-readable `json` formats.

The CLI is designed especially for AI agents and automation workflows. `toon` is the default because it stays structured while using fewer tokens than heavier text or JSON output in normal agent conversations.

## Key Features

- **Single `forecast` command** with Open-Meteo-style variable selection for `current`, `hourly`, and `daily` data
- **Geocoding support** - resolve place names like "Berlin" or "Tokyo" automatically with `--city`
- **Variable discovery** - built-in help for all available weather variables
- **Compact output** - requests only the data you ask for, keeping responses lean for AI context windows
- **Product limits** - capped at 48 hours for hourly data and 14 days for daily data

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

- `forecast` - Get weather forecast with current, hourly, and/or daily data (primary command)

## Usage

### Primary Command: forecast

The `forecast` command is the primary interface. You must specify at least one of `--current`, `--hourly`, or `--daily`.

```bash
# By location name (geocoding)
openmeteo-cli forecast --city "Berlin" --country "Germany" --current default --hourly default --forecast-hours 24

# By coordinates
openmeteo-cli forecast --latitude 52.52 --longitude 13.41 --daily default --forecast-days 7

# Current conditions only
openmeteo-cli forecast --city "Tokyo" --current temperature_2m,weather_code

# Combined request
openmeteo-cli forecast \
  --city "New York" \
  --current temperature_2m,weather_code \
  --hourly temperature_2m,precipitation_probability \
  --forecast-hours 24 \
  --daily weather_code,temperature_2m_max,temperature_2m_min \
  --forecast-days 7
```

### Variable Discovery

List all available variables:

```bash
# All variables
openmeteo-cli forecast variables

# Current weather variables
openmeteo-cli forecast variables current

# Hourly forecast variables
openmeteo-cli forecast variables hourly

# Daily forecast variables
openmeteo-cli forecast variables daily
```

### Compatibility Aliases

The `hourly` and `daily` commands are maintained as compatibility aliases. They map to the new `forecast` command with legacy default variable sets.

```bash
# Hourly (1-2 days) - compatibility alias
openmeteo-cli hourly --latitude <latitude> --longitude <longitude> --forecast-days <1-2>

# Daily (1-14 days) - compatibility alias
openmeteo-cli daily --latitude <latitude> --longitude <longitude> --forecast-days <1-14>
```

**Migration Note:** These aliases will be maintained for one release cycle to allow downstream tools to update. New code should use the `forecast` command directly.

## Help

Show help message:

```bash
openmeteo-cli -h
openmeteo-cli --help
openmeteo-cli forecast --help
openmeteo-cli forecast variables
openmeteo-cli hourly --help
openmeteo-cli daily --help
```

## Options

### Forecast Command Options

- `--city <name>` - City name for geocoding (e.g., "Berlin", "Tokyo, Germany")
- `--country <name>` - Country name to narrow geocoding results (optional)
- `--latitude <float>` - Latitude coordinate (-90 to 90)
- `--longitude <float>` - Longitude coordinate (-180 to 180)
  - `--city` cannot be combined with `--latitude`/`--longitude`
  - `--latitude` and `--longitude` must be provided together when using coordinates

- `--current <vars>` - Current weather variables as comma-separated list or `default`
- `--hourly <vars>` - Hourly forecast variables as comma-separated list or `default`
- `--daily <vars>` - Daily forecast variables as comma-separated list or `default`
  - At least one of `--current`, `--hourly`, or `--daily` is required
  - Use `default` for a sensible set of common variables
  - Variable names use Open-Meteo API field names

- `--forecast-hours <int>` - Number of hours for hourly forecast (required with `--hourly`, 1-48)
- `--forecast-days <int>` - Number of days for daily forecast (required with `--daily`, 1-14)

- `--units metric|imperial` - Units: metric (default) or imperial
- `--format toon|json` - Output format: toon (default) or json

## Default Variable Sets

### Current Default
`temperature_2m`, `apparent_temperature`, `precipitation`, `wind_speed_10m`, `weather_code`

### Hourly Default
`temperature_2m`, `precipitation_probability`, `precipitation`, `wind_speed_10m`, `weather_code`

### Daily Default
`weather_code`, `temperature_2m_min`, `temperature_2m_max`, `precipitation_sum`, `precipitation_probability_max`, `wind_speed_10m_max`, `wind_gusts_10m_max`, `uv_index_max`, `sunrise`, `sunset`

## Available Variables

### Current Weather Variables

| Variable | Description |
|----------|-------------|
| `temperature_2m` | Air temperature at 2 meters above ground |
| `relative_humidity_2m` | Relative humidity at 2 meters above ground |
| `dew_point_2m` | Dew point temperature at 2 meters above ground |
| `apparent_temperature` | Apparent temperature feels like |
| `pressure_msl` | Sea level air pressure |
| `cloud_cover` | Total cloud coverage as percentage |
| `cloud_cover_low` | Low-level cloud coverage as percentage |
| `cloud_cover_mid` | Mid-level cloud coverage as percentage |
| `cloud_cover_high` | High-level cloud coverage as percentage |
| `wind_speed_10m` | Wind speed at 10 meters above ground |
| `wind_speed_80m` | Wind speed at 80 meters above ground |
| `wind_direction_10m` | Wind direction at 10 meters above ground |
| `wind_direction_80m` | Wind direction at 80 meters above ground |
| `wind_gusts_10m` | Wind gust speed at 10 meters above ground |
| `shortwave_radiation` | Shortwave solar radiation |
| `direct_radiation` | Direct solar radiation |
| `diffuse_radiation` | Diffuse solar radiation |
| `direct_normal_irradiance` | Direct normal irradiance |
| `global_tilted_irradiance` | Global tilted irradiance for panels |
| `weather_code` | Weather condition code |
| `snow_depth` | Snow depth on ground |
| `precipitation` | Total precipitation (rain, showers, snow) |
| `showers` | Precipitation from showers in mm |
| `rain` | Precipitation from rain in mm |
| `precipitation_probability` | Probability of precipitation as percentage |
| `visibility` | Horizontal view distance in km |
| `evapotranspiration` | Reference evapotranspiration |
| `uv_index` | UV index |
| `is_day` | 1 for daytime, 0 for nighttime |
| `soil_temperature_0_to_7cm` | Soil temperature at 0-7cm depth |
| `soil_temperature_7_to_28cm` | Soil temperature at 7-28cm depth |
| `soil_temperature_28_to_100cm` | Soil temperature at 28-100cm depth |
| `soil_temperature_100_to_255cm` | Soil temperature at 100-255cm depth |
| `soil_moisture_0_to_7cm` | Soil moisture content at 0-7cm depth |
| `soil_moisture_7_to_28cm` | Soil moisture content at 7-28cm depth |
| `soil_moisture_28_to_100cm` | Soil moisture content at 28-100cm depth |
| `soil_moisture_100_to_255cm` | Soil moisture content at 100-255cm depth |

### Hourly Forecast Variables

Includes all current weather variables plus:
| Variable | Description |
|----------|-------------|
| `soil_temperature_0_to_7cm` | Soil temperature at 0-7cm depth |
| `soil_temperature_7_to_28cm` | Soil temperature at 7-28cm depth |
| `soil_temperature_28_to_100cm` | Soil temperature at 28-100cm depth |
| `soil_temperature_100_to_255cm` | Soil temperature at 100-255cm depth |
| `soil_moisture_0_to_7cm` | Soil moisture content at 0-7cm depth |
| `soil_moisture_7_to_28cm` | Soil moisture content at 7-28cm depth |
| `soil_moisture_28_to_100cm` | Soil moisture content at 28-100cm depth |
| `soil_moisture_100_to_255cm` | Soil moisture content at 100-255cm depth |

### Daily Forecast Variables

| Variable | Description |
|----------|-------------|
| `weather_code` | Weather condition code (noon) |
| `temperature_2m_max` | Maximum air temperature |
| `temperature_2m_min` | Minimum air temperature |
| `apparent_temperature_max` | Maximum apparent temperature |
| `apparent_temperature_min` | Minimum apparent temperature |
| `precipitation_sum` | Total precipitation sum |
| `precipitation_probability_max` | Maximum daily precipitation probability |
| `precipitation_hours` | Number of hours with precipitation |
| `showers_sum` | Precipitation from showers sum |
| `rain_sum` | Precipitation from rain sum |
| `snowfall_sum` | Snowfall sum |
| `snow_depth` | Snow depth on ground |
| `wind_speed_10m_max` | Maximum wind speed at 10m |
| `wind_gusts_10m_max` | Maximum wind gust speed at 10m |
| `wind_direction_10m_dominant` | Dominant wind direction at 10m |
| `shortwave_radiation_sum` | Total shortwave radiation |
| `et0_fao_evapotranspiration` | Reference evapotranspiration |
| `uv_index_max` | Maximum UV index |
| `uv_index_clear_sky_max` | Maximum clear sky UV index |
| `sunrise` | Sunrise time |
| `sunset` | Sunset time |
| `daylight_duration` | Daylight duration in seconds |
| `sunshine_duration` | Sunshine duration in seconds |

## Examples

### Using location names (geocoding)

```bash
# Current conditions for Berlin
openmeteo-cli forecast --city "Berlin" --country "Germany" --current default

# 24-hour hourly forecast for Tokyo
openmeteo-cli forecast --city "Tokyo" --country "Japan" --hourly default --forecast-hours 24

# 7-day daily forecast for New York
openmeteo-cli forecast --city "New York" --country "USA" --daily default --forecast-days 7

# Combined request with custom variables for London
openmeteo-cli forecast --city "London" --country "UK" \
  --current temperature_2m,weather_code \
  --hourly temperature_2m,precipitation_probability \
  --forecast-hours 24 \
  --daily weather_code,temperature_2m_max,temperature_2m_min \
  --forecast-days 5
```

### Using coordinates

```bash
# Current-only request
openmeteo-cli forecast --latitude 52.52 --longitude 13.41 --current temperature_2m,weather_code

# Hourly-only request (max 48 hours)
openmeteo-cli forecast --latitude 40.7128 --longitude -74.006 \
  --hourly temperature_2m,precipitation_probability,wind_speed_10m \
  --forecast-hours 48

# Daily-only request (max 14 days)
openmeteo-cli forecast --latitude 35.6762 --longitude 139.6503 \
  --daily weather_code,temperature_2m_max,temperature_2m_min,precipitation_sum \
  --forecast-days 14

# Combined request with all sections
openmeteo-cli forecast \
  --latitude 51.5074 --longitude -0.1278 \
  --current default \
  --hourly default --forecast-hours 24 \
  --daily default --forecast-days 7
```

### Imperial units and JSON output

```bash
# Imperial units
openmeteo-cli forecast --city "Los Angeles" --country "USA" --current default --units imperial

# JSON output for automation
openmeteo-cli forecast --city "Paris" --country "France" --daily default --forecast-days 7 --format json

# Combined
openmeteo-cli forecast --city "Sydney" --country "Australia" \
  --current temperature_2m,wind_speed_10m \
  --hourly temperature_2m,precipitation_probability \
  --forecast-hours 24 \
  --units imperial \
  --format json
```

### Variable discovery

```bash
# Show all available variables
openmeteo-cli forecast variables

# Show current weather variables
openmeteo-cli forecast variables current

# Show hourly forecast variables
openmeteo-cli forecast variables hourly

# Show daily forecast variables
openmeteo-cli forecast variables daily
```

### Using the default keyword

```bash
# Current weather with default variables
openmeteo-cli forecast --city "Berlin" --country "Germany" --current default

# Hourly with default variables
openmeteo-cli forecast --city "Tokyo" --country "Japan" --hourly default --forecast-hours 24

# Daily with default variables
openmeteo-cli forecast --city "New York" --country "USA" --daily default --forecast-days 7

# Combined: default for some sections, custom for others
openmeteo-cli forecast --city "London" --country "UK" \
  --current default \
  --hourly temperature_2m,precipitation_probability \
  --forecast-hours 24 \
  --daily default \
  --forecast-days 5
```

## Defaults

- Units: `metric`
- Format: `toon`
- Forecast hours limit: 48
- Forecast days limit: 14

## Validation Rules

- At least one of `--current`, `--hourly`, or `--daily` is required
- `--forecast-hours` is required with `--hourly` and must be between 1 and 48
- `--forecast-days` is required with `--daily` and must be between 1 and 14
- `--city` cannot be combined with `--latitude` or `--longitude`
- `--latitude` and `--longitude` must be provided together when using coordinates
- Variable names must be valid Open-Meteo API field names
- Duplicate variable names are rejected
- `default` is a valid explicit value for `--current`, `--hourly`, and `--daily`

## Geocoding

When using `--city`, the CLI queries the Open-Meteo geocoding API to resolve the place name to coordinates. The resolved location metadata is included in the output.

- If the query returns no results, a validation error is returned
- If the query returns multiple plausible results, a validation error is returned asking for refinement (use `--country` to narrow)
- The resolved place name and country are included in the output metadata

## Output Format Notes

### TOON Format (default)

TOON is a compact, human-readable text format designed for structured data and low token usage in agent workflows. Key features:
- Uses YAML-like structure with key-value pairs
- Numeric values remain numeric (not quoted strings)
- Sections for `meta`, `current`, `hourly`, and `daily`

Example:

```toon
meta:
  units:
    temperature: C
    humidity: %
    wind_speed: km/h
    wind_direction: deg
    precipitation: mm
    precipitation_probability: %
    uv_index: index
  timezone: Europe/Berlin
  latitude: 52.52
  longitude: 13.419998
  location:
    name: Berlin
    country: Germany
current:
  time: "2026-03-25T12:00"
  temperature_2m: 10.3
  apparent_temperature: 4.5
  precipitation: 0.1
  wind_speed_10m: 27
  weather: Slight rain showers
daily[3]:
  - date: 2026-03-25
    weekday: Wed
    weather: Slight rain showers
    temperature_2m_min: 5.4
    temperature_2m_max: 13.1
    precipitation_sum: 0.7
    precipitation_probability_max: 100
    wind_speed_10m_max: 27
    wind_gusts_10m_max: 65.9
    uv_index_max: 3.55
    sunrise: "05:56"
    sunset: "18:27"
  - date: 2026-03-26
    weekday: Thu
    weather: Overcast
    temperature_2m_min: 1.9
    temperature_2m_max: 8.4
    precipitation_sum: 0.1
    precipitation_probability_max: 53
    ...

### JSON Format

Standard JSON output with full pretty-printing. Use for programmatic consumption or debugging.

Example:

```json
{
  "meta": {
    "units": {
      "temperature": "C",
      "humidity": "%",
      "wind_speed": "km/h",
      "wind_direction": "deg",
      "precipitation": "mm",
      "precipitation_probability": "%",
      "uv_index": "index"
    },
    "timezone": "Europe/Berlin",
    "latitude": 52.52,
    "longitude": 13.419998,
    "location": {
      "name": "Berlin",
      "country": "Germany"
    }
  },
  "current": {
    "time": "2026-03-25T12:00",
    "temperature_2m": 10.3,
    "apparent_temperature": 4.5,
    "precipitation": 0.1,
    "wind_speed_10m": 27,
    "weather": "Slight rain showers"
  },
  "daily": [
    {
      "date": "2026-03-25",
      "weekday": "Wed",
      "weather": "Slight rain showers",
      "temperature_2m_min": 5.4,
      "temperature_2m_max": 13.1,
      "precipitation_sum": 0.7,
      "precipitation_probability_max": 100,
      "wind_speed_10m_max": 27,
      "wind_gusts_10m_max": 65.9,
      "uv_index_max": 3.55,
      "sunrise": "05:56",
      "sunset": "18:27"
    }
  ]
}
```

## Public Contract

### Command-Line Interface

#### Syntax

```bash
openmeteo-cli forecast --city <place> --country <country> | --latitude <lat> --longitude <lon> \
  (--current <vars> | --hourly <vars> --forecast-hours <n> | --daily <vars> --forecast-days <n>)... \
  [options]
```

#### Commands

- `forecast`: Get weather forecast with current, hourly, and/or daily data (primary)

#### Required Options

At least one section must be specified:
- `--current <vars>`: Current weather variables (comma-separated or `default`)
- `--hourly <vars>`: Hourly forecast variables (comma-separated or `default`), requires `--forecast-hours`
- `--daily <vars>`: Daily forecast variables (comma-separated or `default`), requires `--forecast-days`

Location (one of):
- `--city <place>`: City name for geocoding
- `--country <country>`: Country name to narrow results (optional, recommended)
- `--latitude <lat> --longitude <lon>`: Explicit coordinates

Range limits:
- `--forecast-hours <n>`: Required with `--hourly`, 1-48
- `--forecast-days <n>`: Required with `--daily`, 1-14

#### Optional Options

- `--units metric|imperial`: Weather units (default: `metric`)
- `--format toon|json`: Output format (default: `toon`)

### Output Schema

#### Meta Object (all outputs)

Metadata about the forecast generation, location, and units.

```json
{
  "units": {
    "temperature": "C",
    "humidity": "%",
    "wind_speed": "km/h",
    "wind_direction": "deg",
    "precipitation": "mm",
    "precipitation_probability": "%",
    "uv_index": "index"
  },
  "timezone": "Europe/Berlin",
  "latitude": 52.52,
  "longitude": 13.419998,
  "location": {
    "name": "Berlin",
    "country": "Germany"
  }
}
```

#### Current Section

Present when `--current` is specified. Always includes `time`.

```json
{
  "time": "2026-03-25T12:00",
  "temperature_2m": 10.3,
  "apparent_temperature": 4.5,
  "precipitation": 0.1,
  "wind_speed_10m": 27,
  "weather": "Slight rain showers"
}
```

#### Hourly Section

Present when `--hourly` is specified. Grouped by day with fixed leading fields `date`, `weekday`, and `time`.

```json
[
  {
    "date": "2026-03-25",
    "weekday": "Wed",
    "hours": [
      {
        "time": "00:00",
        "temperature_2m": 6.1,
        "precipitation_probability": 0,
        "weather": "Clear"
      }
    ]
  }
]
```

#### Daily Section

Present when `--daily` is specified. Each row has fixed leading fields `date` and `weekday`.

```json
[
  {
    "date": "2026-03-25",
    "weekday": "Wed",
    "weather": "Overcast",
    "temperature_2m_min": 5.4,
    "temperature_2m_max": 13.1,
    "precipitation_sum": 0.7,
    "precipitation_probability_max": 100,
    "wind_speed_10m_max": 27,
    "wind_gusts_10m_max": 65.9,
    "uv_index_max": 3.55,
    "sunrise": "05:56",
    "sunset": "18:27"
  }
]
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

- `time`: Local time as `YYYY-MM-DDTHH:MM` (e.g., `2026-03-25T12:00`)
- `date`: Date as `YYYY-MM-DD` (e.g., `2026-03-25`)
- `weekday`: Abbreviated weekday name (e.g., `Mon`, `Tue`, `Wed`)
- `sunrise`/`sunset` (daily): Local time as `HH:MM` (e.g., `06:30`, `19:45`)

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 2 | Invalid arguments (missing required options, invalid flag format) |
| 3 | Validation error (lat/lon out of range, invalid units/format, missing sections, bad variable names, geocoding failures) |
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
