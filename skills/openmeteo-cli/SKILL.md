---
name: openmeteo-cli
description: Fetch concise weather forecasts with openmeteo-cli for AI-agent workflows.
---

# openmeteo-cli

Use this skill when you need a fast weather forecast for a specific latitude and longitude.

Use the default output unless another tool needs machine-readable JSON. Add `--format json` only for downstream automation.

## Commands

```bash
# Hourly forecast (max 2 days)
openmeteo-cli forecast --lat <latitude> --lon <longitude> --hourly [--forecast-days 1|2]

# Daily forecast (max 14 days)
openmeteo-cli forecast --lat <latitude> --lon <longitude> --daily [--forecast-days 1-14]
```

## Guidance

- Use `--hourly` for hourly forecast data (maximum 2 days).
- Use `--daily` for daily forecast data (maximum 14 days).
- Add `--units imperial` only when the caller explicitly wants Fahrenheit and mph.
- Add `--format json` only when another tool needs machine-readable output.

## Examples

```bash
openmeteo-cli forecast --lat 40.7128 --lon -74.0060 --hourly
openmeteo-cli forecast --lat 40.7128 --lon -74.0060 --hourly --forecast-days 2
openmeteo-cli forecast --lat 34.0522 --lon -118.2437 --daily --forecast-days 7
openmeteo-cli forecast --lat 51.5074 --lon -0.1278 --daily --forecast-days 14 --units metric
```

## Validation Rules

- Exactly one of `--hourly` or `--daily` is required
- `--hourly` supports maximum 2 days
- `--daily` supports maximum 14 days
- `--forecast-days` defaults to 1 if not specified