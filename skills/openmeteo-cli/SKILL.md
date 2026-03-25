---
name: openmeteo-cli
description: Fetch concise weather forecasts with openmeteo-cli for AI-agent workflows.
---

# openmeteo-cli

Use this skill when you need a fast weather forecast for a location. The CLI supports both place names (via geocoding) and coordinates.

Use the default `toon` output format unless another tool needs `--format json`.

## Commands

### Primary: forecast

The `forecast` command is the primary interface. You must specify at least one of `--current`, `--hourly`, or `--daily`.

```bash
# By city and country (geocoding)
openmeteo-cli forecast --city "<city>" --country "<country>" --current default

# By coordinates
openmeteo-cli forecast --latitude <lat> --longitude <lon> --daily default --forecast-days 7

# Combined request
openmeteo-cli forecast --city "Berlin" --country "Germany" \
  --current default \
  --hourly default --forecast-hours 24 \
  --daily default --forecast-days 7
```

### Variable Discovery

List available variables:

```bash
openmeteo-cli forecast variables
openmeteo-cli forecast variables current
openmeteo-cli forecast variables hourly
openmeteo-cli forecast variables daily
```

## Guidance

- Prefer `--city` with `--country` for better geocoding UX (e.g., "Berlin", "Tokyo, Japan")
- Use `--current` for current conditions, `--hourly` for hourly forecast (max 48 hours), `--daily` for daily forecast (max 14 days)
- Use `default` for sensible variable sets, or specify comma-separated variables (e.g., `temperature_2m,weather_code`)
- `--forecast-hours` is required with `--hourly` (1-48)
- `--forecast-days` is required with `--daily` (1-14)
- Add `--units imperial` only when the caller explicitly wants Fahrenheit and mph
- Add `--format json` only when another tool needs machine-readable output

## Examples

```bash
# Current weather for a city
openmeteo-cli forecast --city "Berlin" --country "Germany" --current default

# 24-hour hourly forecast
openmeteo-cli forecast --city "Tokyo" --country "Japan" --hourly default --forecast-hours 24

# 7-day daily forecast
openmeteo-cli forecast --city "New York" --country "USA" --daily default --forecast-days 7

# Custom variables
openmeteo-cli forecast --city "London" --country "UK" --current temperature_2m,weather_code

# Combined request
openmeteo-cli forecast --city "Paris" --country "France" \
  --current default \
  --hourly temperature_2m,precipitation_probability --forecast-hours 24 \
  --daily default --forecast-days 5

# Using coordinates
openmeteo-cli forecast --latitude 40.7128 --longitude -74.006 --daily default --forecast-days 7
```

## Validation Rules

- At least one of `--current`, `--hourly`, or `--daily` is required
- `--forecast-hours` is required with `--hourly` (1-48)
- `--forecast-days` is required with `--daily` (1-14)
- `--city` cannot be combined with `--latitude`/`--longitude`
- `--latitude` and `--longitude` must be provided together when using coordinates

## Default Variable Sets

### Current Default
`temperature_2m`, `apparent_temperature`, `precipitation`, `wind_speed_10m`, `weather_code`

### Hourly Default
`temperature_2m`, `precipitation_probability`, `precipitation`, `wind_speed_10m`, `weather_code`

### Daily Default
`weather_code`, `temperature_2m_min`, `temperature_2m_max`, `precipitation_sum`, `precipitation_probability_max`, `wind_speed_10m_max`, `wind_gusts_10m_max`, `uv_index_max`, `sunrise`, `sunset`
