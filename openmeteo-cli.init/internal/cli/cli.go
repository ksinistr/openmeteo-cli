// Package cli handles command-line argument parsing and validation.
package cli

import (
	"fmt"
	"time"
)

// Config holds all configuration for a CLI command execution.
type Config struct {
	Command string
	Lat     float64
	Lon     float64
	Units   string
	Format  string
	Date    time.Time
	DateStr string
}

// Parse parses command-line arguments for the given command.
func Parse(command string, args []string) (*Config, error) {
	// Default values
	var lat, lon float64
	var units, format, dateStr string
	var hasLat, hasLon bool

	units = "metric"
	format = "toon"

	// Parse arguments manually
	for i := 0; i < len(args); i++ {
		if args[i] == "--lat" && i+1 < len(args) {
			var tmp float64
			if _, err := fmt.Sscanf(args[i+1], "%f", &tmp); err != nil {
				return nil, fmt.Errorf("invalid lat value")
			}
			lat = tmp
			hasLat = true
			i++
		} else if args[i] == "--lon" && i+1 < len(args) {
			var tmp float64
			if _, err := fmt.Sscanf(args[i+1], "%f", &tmp); err != nil {
				return nil, fmt.Errorf("invalid lon value")
			}
			lon = tmp
			hasLon = true
			i++
		} else if args[i] == "--units" && i+1 < len(args) {
			units = args[i+1]
			i++
		} else if args[i] == "--format" && i+1 < len(args) {
			format = args[i+1]
			i++
		} else if args[i] == "--date" && i+1 < len(args) {
			dateStr = args[i+1]
			i++
		}
	}

	// Validate required parameters
	if !hasLat || !hasLon {
		return nil, fmt.Errorf("lat and lon are required")
	}

	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("latitude must be between -90 and 90")
	}

	if lon < -180 || lon > 180 {
		return nil, fmt.Errorf("longitude must be between -180 and 180")
	}

	if units != "metric" && units != "imperial" {
		return nil, fmt.Errorf("units must be 'metric' or 'imperial'")
	}

	if format != "toon" && format != "json" {
		return nil, fmt.Errorf("format must be 'toon' or 'json'")
	}

	if command == "day" && dateStr == "" {
		return nil, fmt.Errorf("date is required for day command")
	}

	var date time.Time
	if dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
		}
	}

	cfg := &Config{
		Command: command,
		Lat:     lat,
		Lon:     lon,
		Units:   units,
		Format:  format,
		Date:    date,
		DateStr: dateStr,
	}

	return cfg, nil
}
