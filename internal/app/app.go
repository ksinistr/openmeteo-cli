package app

import (
	"errors"
	"fmt"
	"os"

	"openmeteo-cli/internal/cli"
	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/output"
	"openmeteo-cli/internal/weathercode"
)

const usageRoot = `Usage: openmeteo-cli forecast [options]

Commands:
  forecast  Get weather forecast

Options:
  --lat <float>           Latitude coordinate (required, -90 to 90)
  --lon <float>           Longitude coordinate (required, -180 to 180)
  --hourly                Get hourly forecast (max 2 days)
  --daily                 Get daily forecast (max 14 days)
  --forecast-days <int>   Number of days to forecast (default: 1)
  --units metric|imperial Units: metric (default) or imperial
  --format toon|json      Output format: toon (default) or json
  -h, --help              Show this help message

Examples:
  openmeteo-cli forecast --lat 40.7128 --lon -74.0060 --hourly --forecast-days 1
  openmeteo-cli forecast --lat 40.7128 --lon -74.0060 --daily --forecast-days 7
`

const usageHourly = `Usage: openmeteo-cli forecast --lat <float> --lon <float> --hourly [options]

Get hourly weather forecast

Options:
  --lat <float>           Latitude coordinate (required, -90 to 90)
  --lon <float>           Longitude coordinate (required, -180 to 180)
  --hourly                Get hourly forecast (max 2 days)
  --forecast-days <int>   Number of days to forecast (default: 1, max: 2)
  --units metric|imperial Units: metric (default) or imperial
  --format toon|json      Output format: toon (default) or json
  -h, --help              Show this help message
`

const usageDaily = `Usage: openmeteo-cli forecast --lat <float> --lon <float> --daily [options]

Get daily weather forecast

Options:
  --lat <float>           Latitude coordinate (required, -90 to 90)
  --lon <float>           Longitude coordinate (required, -180 to 180)
  --daily                 Get daily forecast (max 14 days)
  --forecast-days <int>   Number of days to forecast (default: 1, max: 14)
  --units metric|imperial Units: metric (default) or imperial
  --format toon|json      Output format: toon (default) or json
  -h, --help              Show this help message
`

// Run is the main entrypoint for the application.
func Run(args []string) int {
	// Check for help flags before any other processing
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usageRoot)
		return 0
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "Error: no command specified\n")
		fmt.Fprint(os.Stderr, usageRoot)
		return 2
	}

	command := args[0]
	commandArgs := args[1:]

	// Check for help flags on command-specific paths
	if command == "forecast" && cli.HasHelpFlag(commandArgs) {
		fmt.Print(usageRoot)
		return 0
	}

	// Validate command before doing any other parsing
	if command != "forecast" {
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "Valid command: forecast")
		return 2
	}

	cfg, err := cli.Parse(commandArgs)
	if err != nil {
		// Check for validation errors (exit 3) vs invalid argument errors (exit 2)
		var ve *cli.ValidationError
		var ia *cli.InvalidArgumentError
		if errors.As(err, &ia) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
		if errors.As(err, &ve) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 3
		}
		// Fallback for other errors
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Use real HTTP client
	return runWithClient(cfg, openmeteo.NewRealHTTPClient())
}

// runWithClient executes the forecast command with the given HTTP client.
// This is primarily used for testing with mock HTTP clients.
func runWithClient(cfg *cli.Config, httpClient openmeteo.HTTPClient) int {
	weatherMapper := weathercode.NewMapper()
	omClient := openmeteo.NewClient(httpClient)
	fcService := forecast.NewService(omClient, weatherMapper)

	var result interface{}
	var errResult error

	// Route based on --hourly/--daily flag
	if cfg.Hourly {
		result, errResult = fcService.Forecast(cfg.Lat, cfg.Lon, cfg.Units, true, cfg.ForecastDays)
	} else if cfg.Daily {
		result, errResult = fcService.Forecast(cfg.Lat, cfg.Lon, cfg.Units, false, cfg.ForecastDays)
	}

	if errResult != nil {
		// Check for upstream API errors (exit 4)
		if errors.Is(errResult, openmeteo.ErrUpstreamAPI) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", errResult)
			return 4
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", errResult)
		return handleError(errResult)
	}

	return writeOutput(cfg.Format, result)
}

func handleError(err error) int {
	if errors.Is(err, forecast.ErrDateUnavailable) {
		return 5
	}
	return 6
}

func writeOutput(format string, data interface{}) int {
	w := output.NewWriter()
	if err := w.Write(data, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 6
	}
	return 0
}
