// Package cli handles command-line argument parsing and validation.
package cli

import (
	"fmt"
	"math"
	"strconv"
)

// Config holds all configuration for a CLI command execution.
type Config struct {
	Mode         string
	ForecastDays int
	Latitude     float64
	Longitude    float64
	Units        string
	Format       string
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
		case "--latitude", "--longitude", "--units", "--format", "--forecast-days":
			if i+1 < len(args) {
				i++
			}
		}
	}
	return false
}

// Parse parses command-line arguments for the hourly or daily command.
// mode is "hourly" or "daily" depending on the subcommand.
func Parse(args []string, mode string) (*Config, error) {
	if HasHelpFlag(args) {
		cfg := &Config{
			Mode:         mode,
			ForecastDays: 1,
			Latitude:     0,
			Longitude:    0,
			Units:        "metric",
			Format:       "toon",
		}
		return cfg, nil
	}

	// Default values
	var latitude, longitude float64
	var units, format string
	var forecastDays int
	var hasLatitude, hasLongitude, hasUnits, hasFormat, hasForecastDays bool

	units = "metric"
	format = "toon"

	// Parse arguments manually
	for i := 0; i < len(args); i++ {
		if args[i] == "--latitude" && i+1 < len(args) {
			tmp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid latitude value"}
			}
			if math.IsNaN(tmp) || math.IsInf(tmp, 0) {
				return nil, &InvalidArgumentError{message: "latitude must be a valid finite number"}
			}
			if hasLatitude {
				return nil, &InvalidArgumentError{message: "duplicate flag: --latitude"}
			}
			latitude = tmp
			hasLatitude = true
			i++
		} else if args[i] == "--longitude" && i+1 < len(args) {
			tmp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid longitude value"}
			}
			if math.IsNaN(tmp) || math.IsInf(tmp, 0) {
				return nil, &InvalidArgumentError{message: "longitude must be a valid finite number"}
			}
			if hasLongitude {
				return nil, &InvalidArgumentError{message: "duplicate flag: --longitude"}
			}
			longitude = tmp
			hasLongitude = true
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
	if !hasLatitude || !hasLongitude {
		return nil, &ValidationError{message: "--latitude and --longitude are required"}
	}

	// Validate forecast-days is required
	if !hasForecastDays {
		return nil, &ValidationError{message: "--forecast-days is required"}
	}

	// Validate forecast-days limits based on mode
	if mode == "hourly" && forecastDays > 2 {
		return nil, &ValidationError{message: "--forecast-days must be between 1 and 2 for hourly forecast"}
	}
	if mode == "daily" && forecastDays > 14 {
		return nil, &ValidationError{message: "--forecast-days must be between 1 and 14 for daily forecast"}
	}

	if latitude < -90 || latitude > 90 {
		return nil, &ValidationError{message: "latitude must be between -90 and 90"}
	}

	if longitude < -180 || longitude > 180 {
		return nil, &ValidationError{message: "longitude must be between -180 and 180"}
	}

	if units != "metric" && units != "imperial" {
		return nil, &ValidationError{message: "units must be 'metric' or 'imperial'"}
	}

	if format != "toon" && format != "json" {
		return nil, &ValidationError{message: "format must be 'toon' or 'json'"}
	}

	cfg := &Config{
		Mode:         mode,
		ForecastDays: forecastDays,
		Latitude:     latitude,
		Longitude:    longitude,
		Units:        units,
		Format:       format,
	}

	return cfg, nil
}
