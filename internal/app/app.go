package app

import (
	"errors"
	"fmt"
	"os"
	"time"

	"openmeteo-cli/internal/cli"
	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/output"
	"openmeteo-cli/internal/weathercode"
)

// Run is the main entrypoint for the application.
func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no command specified")
		fmt.Fprintln(os.Stderr, "Usage: openmeteo-cli (today|day|week) [options]")
		return 2
	}

	command := args[0]
	commandArgs := args[1:]

	// Validate command before doing any other parsing
	validCommands := map[string]bool{"today": true, "day": true, "week": true}
	if !validCommands[command] {
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "Valid commands: today, day, week")
		return 2
	}

	cfg, err := cli.Parse(command, commandArgs)
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

	// Command-specific validation after parsing
	// First check if --date is allowed for the command (before validating format)
	if command != "day" && cfg.DateStr != "" {
		fmt.Fprintln(os.Stderr, "Error: --date flag is only valid for day command")
		return 3
	}
	if command == "day" && cfg.DateStr == "" {
		fmt.Fprintln(os.Stderr, "Error: date is required for day command")
		return 3
	}

	// Now validate date format (only after checking --date is allowed)
	var date time.Time
	if cfg.DateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", cfg.DateStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date format, use YYYY-MM-DD: %v\n", err)
			return 3
		}

		// Verify the date string round-trips correctly (catches invalid dates like Feb 31)
		if date.Format("2006-01-02") != cfg.DateStr {
			fmt.Fprintln(os.Stderr, "Error: invalid date (not a valid calendar date)")
			return 3
		}

		// Validate date is within forecast range (today through 16 days from today)
		today := time.Now().UTC().Truncate(24 * time.Hour)
		minDate := today
		maxDate := today.AddDate(0, 0, 16)
		if date.Before(minDate) || date.After(maxDate) {
			fmt.Fprintln(os.Stderr, "Error: date must be between today and 16 days from today")
			return 3
		}
	}

	// Use real HTTP client
	return runWithClient(cfg, date, command, openmeteo.NewRealHTTPClient())
}

// runWithClient executes the command with the given HTTP client.
// This is primarily used for testing with mock HTTP clients.
func runWithClient(cfg *cli.Config, date time.Time, command string, httpClient openmeteo.HTTPClient) int {
	weatherMapper := weathercode.NewMapper()
	omClient := openmeteo.NewClient(httpClient)
	fcService := forecast.NewService(omClient, weatherMapper)

	var result interface{}
	var errResult error
	switch command {
	case "today":
		result, errResult = fcService.Today(cfg.Lat, cfg.Lon, cfg.Units)
	case "day":
		result, errResult = fcService.Day(cfg.Lat, cfg.Lon, date, cfg.Units)
	case "week":
		result, errResult = fcService.Week(cfg.Lat, cfg.Lon, cfg.Units)
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
