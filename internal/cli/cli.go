// Package cli handles command-line argument parsing and validation.
package cli

import (
	"fmt"
	"math"
	"strconv"
)

// Config holds all configuration for a CLI command execution.
type Config struct {
	Command string
	Lat     float64
	Lon     float64
	Units   string
	Format  string
	DateStr string
}

// ValidationError represents a validation error with exit code 3.
type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}

// InvalidArgumentError represents an invalid argument error with exit code 2.
type InvalidArgumentError struct {
	message string
}

func (e *InvalidArgumentError) Error() string {
	return e.message
}

// Parse parses command-line arguments for the given command.
func Parse(command string, args []string) (*Config, error) {
	// Default values
	var lat, lon float64
	var units, format, dateStr string
	var hasLat, hasLon, hasUnits, hasFormat, hasDate bool

	units = "metric"
	format = "toon"

	// Parse arguments manually
	for i := 0; i < len(args); i++ {
		if args[i] == "--lat" && i+1 < len(args) {
			tmp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid lat value"}
			}
			if math.IsNaN(tmp) || math.IsInf(tmp, 0) {
				return nil, &InvalidArgumentError{message: "lat must be a valid finite number"}
			}
			if hasLat {
				return nil, &InvalidArgumentError{message: "duplicate flag: --lat"}
			}
			lat = tmp
			hasLat = true
			i++
		} else if args[i] == "--lon" && i+1 < len(args) {
			tmp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid lon value"}
			}
			if math.IsNaN(tmp) || math.IsInf(tmp, 0) {
				return nil, &InvalidArgumentError{message: "lon must be a valid finite number"}
			}
			if hasLon {
				return nil, &InvalidArgumentError{message: "duplicate flag: --lon"}
			}
			lon = tmp
			hasLon = true
			i++
		} else if args[i] == "--units" && i+1 < len(args) {
			if hasUnits {
				return nil, &InvalidArgumentError{message: "duplicate flag: --units"}
			}
			units = args[i+1]
			hasUnits = true
			i++
		} else if args[i] == "--format" && i+1 < len(args) {
			if hasFormat {
				return nil, &InvalidArgumentError{message: "duplicate flag: --format"}
			}
			format = args[i+1]
			hasFormat = true
			i++
		} else if args[i] == "--date" && i+1 < len(args) {
			if hasDate {
				return nil, &InvalidArgumentError{message: "duplicate flag: --date"}
			}
			dateStr = args[i+1]
			hasDate = true
			i++
		} else if len(args[i]) > 0 && args[i][0] == '-' {
			// Unknown flag - return error to help catch typos
			return nil, &InvalidArgumentError{message: fmt.Sprintf("unknown flag: %s", args[i])}
		} else {
			// Unexpected positional argument
			return nil, &InvalidArgumentError{message: fmt.Sprintf("unexpected argument: %s", args[i])}
		}
	}

	// Validate required parameters
	if !hasLat || !hasLon {
		return nil, &InvalidArgumentError{message: "lat and lon are required"}
	}

	if lat < -90 || lat > 90 {
		return nil, &ValidationError{message: "latitude must be between -90 and 90"}
	}

	if lon < -180 || lon > 180 {
		return nil, &ValidationError{message: "longitude must be between -180 and 180"}
	}

	if units != "metric" && units != "imperial" {
		return nil, &ValidationError{message: "units must be 'metric' or 'imperial'"}
	}

	if format != "toon" && format != "json" {
		return nil, &ValidationError{message: "format must be 'toon' or 'json'"}
	}

	cfg := &Config{
		Command: command,
		Lat:     lat,
		Lon:     lon,
		Units:   units,
		Format:  format,
		DateStr: dateStr,
	}

	return cfg, nil
}
