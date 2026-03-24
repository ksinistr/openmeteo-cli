---
name: openmeteo-cli
description: Fetch concise weather forecasts with openmeteo-cli for AI-agent workflows.
---

# openmeteo-cli

Use this skill when you need a fast weather forecast for a specific latitude and longitude.

Use the default output unless another tool needs machine-readable JSON. Add `--format json` only for downstream automation.

## Commands

```bash
# Hourly forecast (1-2 days)
openmeteo-cli hourly --latitude <latitude> --longitude <longitude> --forecast-days <1-2>

# Daily forecast (1-14 days)
openmeteo-cli daily --latitude <latitude> --longitude <longitude> --forecast-days <1-14>
```

## Guidance

- Use `hourly` for hourly forecast data (maximum 2 days).
- Use `daily` for daily forecast data (maximum 14 days).
- `--forecast-days` is required (no default).
- Add `--units imperial` only when the caller explicitly wants Fahrenheit and mph.
- Add `--format json` only when another tool needs machine-readable output.

## Examples

```bash
openmeteo-cli hourly --latitude 40.7128 --longitude -74.0060 --forecast-days 1
openmeteo-cli hourly --latitude 40.7128 --longitude -74.0060 --forecast-days 2
openmeteo-cli daily --latitude 34.0522 --longitude -118.2437 --forecast-days 7
openmeteo-cli daily --latitude 51.5074 --longitude -0.1278 --forecast-days 14 --units metric
```

## Validation Rules

- `hourly` command: `--forecast-days` must be 1-2
- `daily` command: `--forecast-days` must be 1-14
- `--forecast-days` is required (no default)