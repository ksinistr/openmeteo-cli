// Package cli handles command-line argument parsing and validation.
package cli

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ForecastConfig holds configuration for the forecast command.
type ForecastConfig struct {
	// Location input
	City      string  // --city (city name)
	Country   string  // --country (country name, optional)
	Latitude  float64 // --latitude (explicit coordinate)
	Longitude float64 // --longitude (explicit coordinate)

	// Request sections
	CurrentVars []string // --current (CSV list or "default")
	HourlyVars  []string // --hourly (CSV list or "default")
	DailyVars   []string // --daily (CSV list or "default")

	// Range limits
	ForecastHours int // --forecast-hours (required with --hourly, 1..48)
	ForecastDays  int // --forecast-days (required with --daily, 1..14)

	// Output options
	Units  string // --units (metric or imperial)
	Format string // --format (toon or json)

	// Special command
	VariablesQuery   string // "variables", "variables current", "variables hourly", or "variables daily"
	IsVariablesQuery bool   // true when this is a variables query command
}

// VariableDefinitions holds the default variable sets and all valid variables.
type VariableDefinitions struct {
	CurrentDefaults []string
	HourlyDefaults  []string
	DailyDefaults   []string
	CurrentVars     map[string]string // name -> description
	HourlyVars      map[string]string // name -> description
	DailyVars       map[string]string // name -> description
}

// GetVariableDefinitions returns the built-in variable definitions.
func GetVariableDefinitions() VariableDefinitions {
	return VariableDefinitions{
		CurrentDefaults: []string{"temperature_2m", "apparent_temperature", "precipitation", "wind_speed_10m", "weather_code"},
		HourlyDefaults:  []string{"temperature_2m", "precipitation_probability", "precipitation", "wind_speed_10m", "weather_code"},
		DailyDefaults:   []string{"weather_code", "temperature_2m_min", "temperature_2m_max", "precipitation_sum", "precipitation_probability_max", "wind_speed_10m_max", "wind_gusts_10m_max", "uv_index_max", "sunrise", "sunset"},
		CurrentVars: map[string]string{
			"temperature_2m":            "Air temperature at 2 meters above ground",
			"relative_humidity_2m":      "Relative humidity at 2 meters above ground",
			"apparent_temperature":      "Apparent temperature feels like",
			"precipitation":             "Total precipitation (rain, showers, snow)",
			"precipitation_probability": "Probability of precipitation as percentage",
			"wind_speed_10m":            "Wind speed at 10 meters above ground",
			"wind_gusts_10m":            "Wind gust speed at 10 meters above ground",
			"wind_direction_10m":        "Wind direction at 10 meters above ground",
			"uv_index":                  "UV index",
			"weather_code":              "Weather condition code",
		},
		HourlyVars: map[string]string{
			"temperature_2m":            "Air temperature at 2 meters above ground",
			"relative_humidity_2m":      "Relative humidity at 2 meters above ground",
			"apparent_temperature":      "Apparent temperature feels like",
			"precipitation":             "Total precipitation (rain, showers, snow)",
			"precipitation_probability": "Probability of precipitation as percentage",
			"wind_speed_10m":            "Wind speed at 10 meters above ground",
			"wind_gusts_10m":            "Wind gust speed at 10 meters above ground",
			"wind_direction_10m":        "Wind direction at 10 meters above ground",
			"uv_index":                  "UV index",
			"weather_code":              "Weather condition code",
		},
		DailyVars: map[string]string{
			"weather_code":                  "Weather condition code (noon)",
			"temperature_2m_max":            "Maximum air temperature",
			"temperature_2m_min":            "Minimum air temperature",
			"precipitation_sum":             "Total precipitation sum",
			"precipitation_probability_max": "Maximum daily precipitation probability",
			"wind_speed_10m_max":            "Maximum wind speed at 10m",
			"wind_gusts_10m_max":            "Maximum wind gust speed at 10m",
			"uv_index_max":                  "Maximum UV index",
			"sunrise":                       "Sunrise time",
			"sunset":                        "Sunset time",
		},
	}
}

// ExpandDefaultVars expands the "default" keyword to the actual default variable set.
func (defs VariableDefinitions) ExpandDefaultVars(vars []string, section string) []string {
	result := make([]string, 0, len(vars))
	for _, v := range vars {
		if v == "default" {
			switch section {
			case "current":
				result = append(result, defs.CurrentDefaults...)
			case "hourly":
				result = append(result, defs.HourlyDefaults...)
			case "daily":
				result = append(result, defs.DailyDefaults...)
			}
		} else {
			result = append(result, v)
		}
	}
	return result
}

// ValidateVars checks if all variables are valid for the given section.
func (v VariableDefinitions) ValidateVars(vars []string, section string) error {
	var validVars map[string]string
	switch section {
	case "current":
		validVars = v.CurrentVars
	case "hourly":
		validVars = v.HourlyVars
	case "daily":
		validVars = v.DailyVars
	default:
		return fmt.Errorf("invalid section: %s", section)
	}

	for _, variable := range vars {
		if variable == "default" {
			continue
		}
		if _, ok := validVars[variable]; !ok {
			return fmt.Errorf("unknown %s variable: %s", section, variable)
		}
	}
	return nil
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
		case "--latitude", "--longitude", "--units", "--format", "--forecast-days", "--current", "--hourly", "--daily", "--forecast-hours", "--city", "--country":
			if i+1 < len(args) {
				i++
			}
		}
	}
	return false
}

// ParseForecast parses command-line arguments for the forecast command.
func ParseForecast(args []string) (*ForecastConfig, error) {
	if HasHelpFlag(args) {
		cfg := &ForecastConfig{
			Latitude:     0,
			Longitude:    0,
			ForecastDays: 1,
			Units:        "metric",
			Format:       "toon",
		}
		return cfg, nil
	}

	// Check for variables query (e.g., "forecast variables" or "forecast variables current")
	if len(args) > 0 && args[0] == "variables" {
		var query string
		if len(args) > 1 {
			query = strings.Join(args[1:], " ")
		}
		cfg := &ForecastConfig{
			VariablesQuery:   query,
			IsVariablesQuery: true, // Mark that this is a variables query regardless of the query string
			Units:            "metric",
			Format:           "toon",
		}
		return cfg, nil
	}

	var cfg ForecastConfig

	// Default values
	cfg.Units = "metric"
	cfg.Format = "toon"

	// Track which flags have been seen
	var hasCity, hasCountry, hasLatitude, hasLongitude bool
	var hasCurrent, hasHourly, hasDaily bool
	var hasForecastHours, hasForecastDays bool
	var hasUnits, hasFormat bool

	// Parse arguments manually
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--city" && i+1 < len(args):
			if hasCity {
				return nil, &InvalidArgumentError{message: "duplicate flag: --city"}
			}
			cfg.City = args[i+1]
			if cfg.City == "" {
				return nil, &InvalidArgumentError{message: "--city requires a non-empty value"}
			}
			hasCity = true
			i++

		case args[i] == "--city":
			return nil, &InvalidArgumentError{message: "flag --city requires a value"}

		case args[i] == "--country" && i+1 < len(args):
			if hasCountry {
				return nil, &InvalidArgumentError{message: "duplicate flag: --country"}
			}
			cfg.Country = args[i+1]
			hasCountry = true
			i++

		case args[i] == "--country":
			return nil, &InvalidArgumentError{message: "flag --country requires a value"}

		case args[i] == "--latitude" && i+1 < len(args):
			if hasLatitude {
				return nil, &InvalidArgumentError{message: "duplicate flag: --latitude"}
			}
			tmp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid latitude value"}
			}
			if math.IsNaN(tmp) || math.IsInf(tmp, 0) {
				return nil, &InvalidArgumentError{message: "latitude must be a valid finite number"}
			}
			cfg.Latitude = tmp
			hasLatitude = true
			i++

		case args[i] == "--latitude":
			return nil, &InvalidArgumentError{message: "flag --latitude requires a value"}

		case args[i] == "--longitude" && i+1 < len(args):
			if hasLongitude {
				return nil, &InvalidArgumentError{message: "duplicate flag: --longitude"}
			}
			tmp, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid longitude value"}
			}
			if math.IsNaN(tmp) || math.IsInf(tmp, 0) {
				return nil, &InvalidArgumentError{message: "longitude must be a valid finite number"}
			}
			cfg.Longitude = tmp
			hasLongitude = true
			i++

		case args[i] == "--longitude":
			return nil, &InvalidArgumentError{message: "flag --longitude requires a value"}

		case args[i] == "--current" && i+1 < len(args):
			if hasCurrent {
				return nil, &InvalidArgumentError{message: "duplicate flag: --current"}
			}
			vars := strings.Split(args[i+1], ",")
			for i, v := range vars {
				vars[i] = strings.TrimSpace(v)
			}
			cfg.CurrentVars = vars
			hasCurrent = true
			i++

		case args[i] == "--current":
			return nil, &InvalidArgumentError{message: "flag --current requires a value (e.g., --current temperature_2m,weather_code)"}

		case args[i] == "--hourly" && i+1 < len(args):
			if hasHourly {
				return nil, &InvalidArgumentError{message: "duplicate flag: --hourly"}
			}
			vars := strings.Split(args[i+1], ",")
			for i, v := range vars {
				vars[i] = strings.TrimSpace(v)
			}
			cfg.HourlyVars = vars
			hasHourly = true
			i++

		case args[i] == "--hourly":
			return nil, &InvalidArgumentError{message: "flag --hourly requires a value"}

		case args[i] == "--daily" && i+1 < len(args):
			if hasDaily {
				return nil, &InvalidArgumentError{message: "duplicate flag: --daily"}
			}
			vars := strings.Split(args[i+1], ",")
			for i, v := range vars {
				vars[i] = strings.TrimSpace(v)
			}
			cfg.DailyVars = vars
			hasDaily = true
			i++

		case args[i] == "--daily":
			return nil, &InvalidArgumentError{message: "flag --daily requires a value"}

		case args[i] == "--forecast-hours" && i+1 < len(args):
			if hasForecastHours {
				return nil, &InvalidArgumentError{message: "duplicate flag: --forecast-hours"}
			}
			tmp, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid forecast-hours value"}
			}
			if tmp < 1 {
				return nil, &ValidationError{message: "--forecast-hours must be at least 1"}
			}
			if tmp > 48 {
				return nil, &ValidationError{message: "--forecast-hours must be at most 48"}
			}
			cfg.ForecastHours = tmp
			hasForecastHours = true
			i++

		case args[i] == "--forecast-hours":
			return nil, &InvalidArgumentError{message: "flag --forecast-hours requires a value"}

		case args[i] == "--forecast-days" && i+1 < len(args):
			if hasForecastDays {
				return nil, &InvalidArgumentError{message: "duplicate flag: --forecast-days"}
			}
			tmp, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, &InvalidArgumentError{message: "invalid forecast-days value"}
			}
			if tmp < 1 {
				return nil, &ValidationError{message: "--forecast-days must be at least 1"}
			}
			if tmp > 14 {
				return nil, &ValidationError{message: "--forecast-days must be at most 14"}
			}
			cfg.ForecastDays = tmp
			hasForecastDays = true
			i++

		case args[i] == "--forecast-days":
			return nil, &InvalidArgumentError{message: "flag --forecast-days requires a value"}

		case args[i] == "--units" && i+1 < len(args):
			if hasUnits {
				return nil, &InvalidArgumentError{message: "duplicate flag: --units"}
			}
			cfg.Units = args[i+1]
			hasUnits = true
			i++

		case args[i] == "--units":
			return nil, &InvalidArgumentError{message: "flag --units requires a value (metric or imperial)"}

		case args[i] == "--format" && i+1 < len(args):
			if hasFormat {
				return nil, &InvalidArgumentError{message: "duplicate flag: --format"}
			}
			cfg.Format = args[i+1]
			hasFormat = true
			i++

		case args[i] == "--format":
			return nil, &InvalidArgumentError{message: "flag --format requires a value (toon or json)"}

		case len(args[i]) > 0 && args[i][0] == '-':
			return nil, &InvalidArgumentError{message: fmt.Sprintf("unknown flag: %s", args[i])}

		default:
			return nil, &InvalidArgumentError{message: fmt.Sprintf("unexpected argument: %s", args[i])}
		}
	}

	// Validation: at least one section required
	if !hasCurrent && !hasHourly && !hasDaily {
		return nil, &ValidationError{message: "at least one of --current, --hourly, or --daily is required"}
	}

	// Validation: --forecast-hours required with --hourly
	if hasHourly && !hasForecastHours {
		return nil, &ValidationError{message: "--forecast-hours is required when --hourly is specified"}
	}

	// Validation: --forecast-days required with --daily
	if hasDaily && !hasForecastDays {
		return nil, &ValidationError{message: "--forecast-days is required when --daily is specified"}
	}

	// Validation: city/country mutually exclusive with latitude/longitude
	if hasCity && (hasLatitude || hasLongitude) {
		return nil, &ValidationError{message: "--city cannot be combined with --latitude or --longitude"}
	}

	// Validation: latitude and longitude must be provided together
	if (hasLatitude && !hasLongitude) || (hasLongitude && !hasLatitude) {
		return nil, &ValidationError{message: "--latitude and --longitude must be provided together"}
	}

	// Validation: city or coordinates required
	if !hasCity && !hasLatitude && !hasLongitude {
		return nil, &ValidationError{message: "either --city or both --latitude and --longitude are required"}
	}

	// Validate coordinate ranges
	if hasLatitude && (cfg.Latitude < -90 || cfg.Latitude > 90) {
		return nil, &ValidationError{message: "latitude must be between -90 and 90"}
	}
	if hasLongitude && (cfg.Longitude < -180 || cfg.Longitude > 180) {
		return nil, &ValidationError{message: "longitude must be between -180 and 180"}
	}

	// Validate units
	if cfg.Units != "metric" && cfg.Units != "imperial" {
		return nil, &ValidationError{message: "units must be 'metric' or 'imperial'"}
	}

	// Validate format
	if cfg.Format != "toon" && cfg.Format != "json" {
		return nil, &ValidationError{message: "format must be 'toon' or 'json'"}
	}

	// Validate variables against variable definitions
	varDefs := GetVariableDefinitions()

	// Check for duplicate variables
	if hasCurrent {
		if err := validateNoDuplicates(cfg.CurrentVars, "current"); err != nil {
			return nil, err
		}
		if err := varDefs.ValidateVars(cfg.CurrentVars, "current"); err != nil {
			return nil, &ValidationError{message: err.Error()}
		}
	}

	if hasHourly {
		if err := validateNoDuplicates(cfg.HourlyVars, "hourly"); err != nil {
			return nil, err
		}
		if err := varDefs.ValidateVars(cfg.HourlyVars, "hourly"); err != nil {
			return nil, &ValidationError{message: err.Error()}
		}
	}

	if hasDaily {
		if err := validateNoDuplicates(cfg.DailyVars, "daily"); err != nil {
			return nil, err
		}
		if err := varDefs.ValidateVars(cfg.DailyVars, "daily"); err != nil {
			return nil, &ValidationError{message: err.Error()}
		}
	}

	return &cfg, nil
}

// validateNoDuplicates checks for duplicate variables in a list.
func validateNoDuplicates(vars []string, section string) error {
	seen := make(map[string]bool)
	for _, v := range vars {
		if seen[v] {
			return &ValidationError{message: fmt.Sprintf("duplicate %s variable: %s", section, v)}
		}
		seen[v] = true
	}
	return nil
}
