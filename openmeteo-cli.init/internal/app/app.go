// Package app contains the application logic for openmeteo-cli.
package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"openmeteo-cli/internal/cli"
	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/openmeteo"
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

	cfg, err := cli.Parse(command, commandArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	weatherMapper := weathercode.NewMapper()
	httpClient := &openmeteo.RealHTTPClient{}
	omClient := openmeteo.NewClient(httpClient)
	fcService := forecast.NewService(omClient, weatherMapper)

	var result interface{}
	var errResult error
	switch command {
	case "today":
		result, errResult = fcService.Today(cfg.Lat, cfg.Lon, cfg.Units)
	case "day":
		result, errResult = fcService.Day(cfg.Lat, cfg.Lon, cfg.Date, cfg.Units)
	case "week":
		result, errResult = fcService.Week(cfg.Lat, cfg.Lon, cfg.Units)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "Valid commands: today, day, week")
		return 2
	}

	if errResult != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", errResult)
		return handleError(errResult)
	}

	return writeOutput(result, cfg.Format)
}

func handleError(err error) int {
	if err == forecast.ErrDateUnavailable {
		return 5
	}
	return 6
}

func writeOutput(data interface{}, format string) int {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return 6
	}
	fmt.Print(buf.String())
	return 0
}
