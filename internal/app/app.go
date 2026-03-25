package app

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"openmeteo-cli/internal/cli"
	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/output"
	"openmeteo-cli/internal/weathercode"
)

const usageRoot = `Usage: openmeteo-cli <command> [options]

Commands:
  forecast  Get weather forecast with variable selection

Options:
  --city <name>            City name to geocode (e.g., "Berlin")
  --country <name>         Country name to narrow results (optional)
  --latitude <float>       Latitude coordinate (-90 to 90)
  --longitude <float>      Longitude coordinate (-180 to 180)
  --units metric|imperial  Units: metric (default) or imperial
  --format toon|json       Output format: toon (default) or json
  -h, --help               Show this help message

Examples:
  openmeteo-cli forecast --city Berlin --country Germany --current default --daily default --forecast-days 7
  openmeteo-cli forecast --latitude 52.52 --longitude 13.41 --hourly temperature_2m --forecast-hours 24
`

const usageForecast = `Usage: openmeteo-cli forecast --city <name>|--latitude <lat> --longitude <lon> [options]

Get weather forecast with variable selection

At least one section (--current, --hourly, or --daily) is required.
Use "default" to get the recommended variable set for each section.
Run "forecast variables" for available variables and their descriptions.

Section Options:
  --current <vars>       Current conditions: CSV list or "default"
  --hourly <vars>        Hourly data: CSV list or "default" (requires --forecast-hours)
  --daily <vars>         Daily data: CSV list or "default" (requires --forecast-days)

Location Options (one required):
  --city <name>          City name to geocode (e.g., "Berlin")
  --country <name>       Country name to narrow results (optional, recommended)
  --latitude <float>     Latitude coordinate (-90 to 90)
  --longitude <float>    Longitude coordinate (-180 to 180)

Range Limits:
  --forecast-hours <int> Number of hours to return (1-48, required with --hourly)
  --forecast-days <int>  Number of days to return (1-14, required with --daily)

Output Options:
  --units metric|imperial Units: metric (default) or imperial
  --format toon|json      Output format: toon (default) or json

Variable Discovery:
  forecast variables       Show all available variables
  forecast variables current  Show current weather variables
  forecast variables hourly   Show hourly forecast variables
  forecast variables daily    Show daily forecast variables

Examples:
  # Current conditions for a city
  openmeteo-cli forecast --city Berlin --country Germany --current default

  # Hourly temperature for 24 hours
  openmeteo-cli forecast --city Tokyo --hourly temperature_2m --forecast-hours 24

  # 7-day daily forecast with default variables
  openmeteo-cli forecast --latitude 40.7 --longitude -74.0 --daily default --forecast-days 7

  # Combined current and daily forecast
  openmeteo-cli forecast --city Paris --country France --current default --daily default --forecast-days 5

  # Show available variables
  openmeteo-cli forecast variables
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
	if cli.HasHelpFlag(commandArgs) {
		switch command {
		case "forecast":
			fmt.Print(usageForecast)
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
			fmt.Fprintln(os.Stderr, "Valid commands: forecast")
			return 2
		}
		return 0
	}

	// Route to command handler
	switch command {
	case "forecast":
		return runForecast(commandArgs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "Valid commands: forecast")
		return 2
	}
}

// runForecast executes the forecast command with variable selection.
func runForecast(args []string) int {
	// Parse forecast arguments
	cfg, err := cli.ParseForecast(args)
	if err != nil {
		return handleParseError(err)
	}

	// Handle variables query command
	if cfg.IsVariablesQuery {
		return printVariableHelp(cfg.VariablesQuery)
	}

	// Use real HTTP client
	return runForecastWithClient(cfg, openmeteo.NewRealHTTPClient())
}

// runForecastWithClient executes the forecast command with the given HTTP client.
func runForecastWithClient(cfg *cli.ForecastConfig, httpClient openmeteo.HTTPClient) int {
	weatherMapper := weathercode.NewMapper()
	omClient := openmeteo.NewClient(httpClient)
	fcService := forecast.NewService(omClient, weatherMapper)

	// Resolve location if provided
	var location *openmeteo.ResolvedLocation
	var lat, lon float64

	if cfg.City != "" {
		// Build search query from city and optional country
		searchQuery := cfg.City
		if cfg.Country != "" {
			searchQuery = cfg.City + ", " + cfg.Country
		}

		// Geocode the location - request up to 5 results to detect ambiguity
		loc, err := omClient.FetchLocation(searchQuery, 5)
		if err != nil {
			if errors.Is(err, openmeteo.ErrLocationNotFound) {
				fmt.Fprintf(os.Stderr, "Error: location not found: %s\n", searchQuery)
				return 3
			}
			if errors.Is(err, openmeteo.ErrLocationAmbiguous) {
				fmt.Fprintf(os.Stderr, "Error: location is ambiguous: %s\n", cfg.City)
				fmt.Fprintln(os.Stderr, "Matching locations:")
				// Fetch all results to show options
				results, fetchErr := omClient.FetchLocationRaw(searchQuery, 10)
				if fetchErr == nil && len(results) > 0 {
					for _, r := range results {
						fmt.Fprintf(os.Stderr, "  --city %s --country %s", r.Name, r.Country)
						if r.Admin1 != "" {
							fmt.Fprintf(os.Stderr, " (region: %s)", r.Admin1)
						}
						fmt.Fprintln(os.Stderr)
					}
				}
				return 3
			}
			// Upstream API error
			if errors.Is(err, openmeteo.ErrUpstreamAPI) {
				fmt.Fprintf(os.Stderr, "Error: geocoding failed: %v\n", err)
				return 4
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 4
		}
		location = loc
		lat = loc.Latitude
		lon = loc.Longitude
	} else {
		lat = cfg.Latitude
		lon = cfg.Longitude
	}

	// Expand "default" keyword to actual variable sets
	varDefs := cli.GetVariableDefinitions()
	currentVars := varDefs.ExpandDefaultVars(cfg.CurrentVars, "current")
	hourlyVars := varDefs.ExpandDefaultVars(cfg.HourlyVars, "hourly")
	dailyVars := varDefs.ExpandDefaultVars(cfg.DailyVars, "daily")

	// Build forecast request
	req := forecast.ForecastRequest{
		Latitude:            lat,
		Longitude:           lon,
		CurrentVars:         currentVars,
		HourlyVars:          hourlyVars,
		DailyVars:           dailyVars,
		HourlyForecastHours: cfg.ForecastHours,
		DailyForecastDays:   cfg.ForecastDays,
		Units:               cfg.Units,
		Location:            location,
	}

	// Get forecast
	result, err := fcService.ForecastVariable(req)
	if err != nil {
		return handleServiceError(err)
	}

	// Write output
	return writeOutput(cfg.Format, result)
}

// handleParseError converts parsing errors to appropriate exit codes.
func handleParseError(err error) int {
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

// handleServiceError converts service errors to appropriate exit codes.
func handleServiceError(err error) int {
	if errors.Is(err, openmeteo.ErrUpstreamAPI) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 4
	}
	if errors.Is(err, forecast.ErrDateUnavailable) {
		return 5
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return 6
}

// printVariableHelp prints variable help based on the query.
func printVariableHelp(query string) int {
	varDefs := cli.GetVariableDefinitions()

	// Determine which section to show
	section := ""
	if query != "" {
		section = query
	}

	fmt.Println("Forecast Variables")
	fmt.Println()
	fmt.Println("Use the 'default' keyword to get recommended variables:")
	fmt.Println("  --current default")
	fmt.Println("  --hourly default")
	fmt.Println("  --daily default")
	fmt.Println()
	fmt.Println("Or specify variables as a comma-separated list:")
	fmt.Println("  --current temperature_2m,weather_code")
	fmt.Println("  --hourly temperature_2m,precipitation_probability")
	fmt.Println("  --daily weather_code,temperature_2m_max,temperature_2m_min")
	fmt.Println()

	switch section {
	case "current":
		printVariableSection("Current Weather", varDefs.CurrentVars, varDefs.CurrentDefaults)
	case "hourly":
		printVariableSection("Hourly Forecast", varDefs.HourlyVars, varDefs.HourlyDefaults)
	case "daily":
		printVariableSection("Daily Forecast", varDefs.DailyVars, varDefs.DailyDefaults)
	default:
		// Show all sections
		printVariableSection("Current Weather", varDefs.CurrentVars, varDefs.CurrentDefaults)
		fmt.Println()
		printVariableSection("Hourly Forecast", varDefs.HourlyVars, varDefs.HourlyDefaults)
		fmt.Println()
		printVariableSection("Daily Forecast", varDefs.DailyVars, varDefs.DailyDefaults)
	}

	return 0
}

// printVariableSection prints a variable section with name, description, and default membership.
func printVariableSection(title string, vars map[string]string, defaults []string) {
	defaultSet := make(map[string]bool, len(defaults))
	for _, v := range defaults {
		defaultSet[v] = true
	}

	// Sort variable names for consistent output
	sortedNames := make([]string, 0, len(vars))
	for name := range vars {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	fmt.Println(title)
	fmt.Println(strings.Repeat("-", len(title)))
	fmt.Printf("%-30s %-60s %s\n", "Variable", "Description", "Default")
	fmt.Println(strings.Repeat("-", 110))

	for _, name := range sortedNames {
		desc := vars[name]
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		defaultMark := " "
		if defaultSet[name] {
			defaultMark = "*"
		}
		fmt.Printf("%-30s %-60s %s\n", name, desc, defaultMark)
	}
	fmt.Println()
	fmt.Println("* = included in default set")
}

func writeOutput(format string, data interface{}) int {
	w := output.NewWriter()
	if err := w.Write(data, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 6
	}
	return 0
}
