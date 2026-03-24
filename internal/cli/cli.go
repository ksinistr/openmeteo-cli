// Package cli handles command-line argument parsing and validation.
package cli

import (
	"fmt"
	"math"
	"strconv"
)

// Config holds all configuration for a CLI command execution.
type Config struct {
	Hourly      bool
	Daily       bool
	ForecastDays int
	Lat         float64
	Lon         float64
	Units       string
	Format      string
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

func HasHelpFlag(args []string) bool {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			return true
		case "--lat", "--lon", "--units", "--format", "--forecast-days":
			if i+1 < len(args) {
				i++
			}
		}
	}
	return false
}

// Parse parses command-line arguments for the forecast command.
func Parse(args []string) (*Config, error) {
	if HasHelpFlag(args) {
		cfg := &Config{
			Hourly:       false,
			Daily:        false,
			ForecastDays: 1,
			Lat:          0,
			Lon:          0,
			Units:        "metric",
			Format:       "toon",
		}
		return cfg, nil
	}

	// Default values
	var lat, lon float64
	var units, format string
	var forecastDays int
	var hourly, daily bool
	var hasLat, hasLon, hasUnits, hasFormat, hasForecastDays bool
	var hasHourly, hasDaily bool

	units = "metric"
	format = "toon"
	forecastDays = 1

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
		} else if args[i] == "--hourly" {
			if hasHourly {
				return nil, &InvalidArgumentError{message: "duplicate flag: --hourly"}
			}
			hourly = true
			hasHourly = true
		} else if args[i] == "--daily" {
			if hasDaily {
				return nil, &InvalidArgumentError{message: "duplicate flag: --daily"}
			}
			daily = true
			hasDaily = true
		} else if args[i] == "--forecast-days" && i+1 < len(args) {
			if hasForecastDays {
				return nil, &InvalidArgumentError{message: "duplicate flag: --forecast-days"}
			}
			tmp, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid forecast-days value"}
			}
			if tmp < 1 {
				return nil, &InvalidArgumentError{message: "forecast-days must be at least 1"}
			}
			forecastDays = tmp
			hasForecastDays = true
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

	// Validate that exactly one of --hourly or --daily is specified
	if hasHourly && hasDaily {
		return nil, &ValidationError{message: "cannot use both --hourly and --daily"}
	}
	if !hasHourly && !hasDaily {
		return nil, &ValidationError{message: "must specify either --hourly or --daily"}
	}

	// Validate forecast-days limits
	if hourly && forecastDays > 2 {
		return nil, &ValidationError{message: "--hourly supports maximum 2 days"}
	}
	if daily && forecastDays > 14 {
		return nil, &ValidationError{message: "--daily supports maximum 14 days"}
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
		Hourly:       hourly,
		Daily:        daily,
		ForecastDays: forecastDays,
		Lat:          lat,
		Lon:          lon,
		Units:        units,
		Format:       format,
	}

	return cfg, nil
}
