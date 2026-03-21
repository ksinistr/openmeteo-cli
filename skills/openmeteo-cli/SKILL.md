---
name: openmeteo-cli
description: Fetch concise weather forecasts with openmeteo-cli for AI-agent workflows.
---

# openmeteo-cli

Use this skill when you need a fast weather forecast for a specific latitude and longitude.

Prefer the default `toon` format for agent-facing output because it is compact and minimizes token usage. Use `json` only when another tool needs machine-readable output.

## Commands

```bash
openmeteo-cli today --lat <latitude> --lon <longitude>
openmeteo-cli day --lat <latitude> --lon <longitude> --date YYYY-MM-DD
openmeteo-cli week --lat <latitude> --lon <longitude>
```

## Guidance

- Use `today` for current conditions plus hourly rows for the local date.
- Use `day` when the caller already knows the exact date.
- Use `week` for a compact seven-day summary.
- Add `--units imperial` only when the caller explicitly wants Fahrenheit and mph.
- Add `--format json` only for downstream automation.

## Examples

```bash
openmeteo-cli today --lat 40.7128 --lon -74.0060
openmeteo-cli day --lat 34.0522 --lon -118.2437 --date 2026-04-15
openmeteo-cli week --lat 51.5074 --lon -0.1278 --units metric
```
