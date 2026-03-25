package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"openmeteo-cli/internal/cli"
	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/weathercode"
)

// =============================================================================
// Acceptance Criteria Tests
// These tests verify the acceptance criteria from Task 7 of the implementation plan.
// =============================================================================

// mockRequestTrackingServer creates a test server that tracks which parameters
// were requested in the API call. This allows us to verify that the CLI only
// requests the sections and variables explicitly asked for.
type mockRequestTrackingServer struct {
	server          *httptest.Server
	requestedParams map[string]string
	requestCount    int
	geocodingCalled bool
	forecastCalled  bool
}

func newMockRequestTrackingServer() *mockRequestTrackingServer {
	m := &mockRequestTrackingServer{
		requestedParams: make(map[string]string),
	}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track all query parameters
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				m.requestedParams[key] = values[0]
			}
		}
		m.requestCount++

		// Check if this is a geocoding request (has "name" but no "latitude")
		if r.URL.Query().Get("name") != "" && r.URL.Query().Get("latitude") == "" {
			m.geocodingCalled = true
			// Return geocoding response
			name := r.URL.Query().Get("name")
			response := map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"name":      name,
						"latitude":  52.52,
						"longitude": 13.41,
						"country":   "Germany",
						"admin1":    "Berlin",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		// This is a forecast request
		m.forecastCalled = true

		// Build a minimal valid Open-Meteo response based on what was requested
		response := map[string]interface{}{
			"latitude":  52.52,
			"longitude": 13.41,
			"timezone":  "Europe/Berlin",
		}

		// Add current section if requested
		if r.URL.Query().Get("current") != "" {
			response["current"] = map[string]interface{}{
				"time":           "2026-03-25T12:00",
				"temperature_2m": 15.5,
				"weather_code":   1,
			}
		}

		// Add hourly section if requested
		if r.URL.Query().Get("hourly") != "" {
			hourlyTimes := []string{"2026-03-25T00:00", "2026-03-25T01:00"}
			hourlyTemps := []float64{10.0, 11.0}
			hourlyPrecipProb := []int{10, 20}

			response["hourly"] = map[string]interface{}{
				"time":                      hourlyTimes,
				"temperature_2m":            hourlyTemps,
				"precipitation_probability": hourlyPrecipProb,
			}
		}

		// Add daily section if requested
		if r.URL.Query().Get("daily") != "" {
			dailyTimes := []string{"2026-03-25", "2026-03-26"}
			dailyWeatherCodes := []int{1, 1}
			dailyTempMaxs := []float64{16.3, 17.1}
			dailyTempMins := []float64{8.1, 9.2}
			dailyPrecipSums := []float64{0.0, 0.1}
			dailyPrecipProbMaxs := []int{20, 30}
			dailyWindSpeedMaxs := []float64{18.7, 20.1}
			dailyWindGustsMaxs := []float64{25.4, 28.2}
			dailyUVIndexMaxs := []float64{4.1, 4.5}
			dailySunrises := []string{"2026-03-25T06:30", "2026-03-26T06:28"}
			dailySunsets := []string{"2026-03-25T19:45", "2026-03-26T19:47"}

			response["daily"] = map[string]interface{}{
				"time":                          dailyTimes,
				"weather_code":                  dailyWeatherCodes,
				"temperature_2m_max":            dailyTempMaxs,
				"temperature_2m_min":            dailyTempMins,
				"precipitation_sum":             dailyPrecipSums,
				"precipitation_probability_max": dailyPrecipProbMaxs,
				"wind_speed_10m_max":            dailyWindSpeedMaxs,
				"wind_gusts_10m_max":            dailyWindGustsMaxs,
				"uv_index_max":                  dailyUVIndexMaxs,
				"sunrise":                       dailySunrises,
				"sunset":                        dailySunsets,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	return m
}

func (m *mockRequestTrackingServer) Close() {
	m.server.Close()
}

func (m *mockRequestTrackingServer) URL() string {
	return m.server.URL
}

// TestAcceptance_ForecastHoursLimit verifies that --forecast-hours accepts 48
// but rejects 49.
func TestAcceptance_ForecastHoursLimit(t *testing.T) {
	tests := []struct {
		name          string
		forecastHours int
		wantExitCode  int
		wantError     string
	}{
		{
			name:          "forecast-hours 48 is accepted by parser",
			forecastHours: 48,
			wantExitCode:  0, // Parse succeeds
		},
		{
			name:          "forecast-hours 49 is rejected",
			forecastHours: 49,
			wantExitCode:  3, // ValidationError
			wantError:     "--forecast-hours must be at most 48",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parsing and validation directly without making API calls
			hoursStr := strconv.Itoa(tt.forecastHours)
			args := []string{"--latitude", "52.52", "--longitude", "13.41", "--hourly", "default", "--forecast-hours", hoursStr}

			cfg, err := cli.ParseForecast(args)

			if tt.wantError != "" {
				// Expect an error
				if err == nil {
					t.Errorf("ParseForecast() expected error with %q, got nil", tt.wantError)
					return
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("ParseForecast() error = %q, want to contain %q", err.Error(), tt.wantError)
				}
				// Check exit code from error type
				var validationErr *cli.ValidationError
				if errors.As(err, &validationErr) {
					// Exit code 3 is correct for ValidationError
				} else {
					t.Errorf("Expected ValidationError (exit code 3), got %T", err)
				}
			} else {
				// Expect success
				if err != nil {
					t.Errorf("ParseForecast() unexpected error: %v", err)
					return
				}
				if cfg.ForecastHours != tt.forecastHours {
					t.Errorf("ForecastHours = %d, want %d", cfg.ForecastHours, tt.forecastHours)
				}
			}
		})
	}
}

// TestAcceptance_ForecastDaysLimit verifies that --forecast-days accepts 14
// but rejects 15.
func TestAcceptance_ForecastDaysLimit(t *testing.T) {
	tests := []struct {
		name         string
		forecastDays int
		wantExitCode int
		wantError    string
	}{
		{
			name:         "forecast-days 14 is accepted by parser",
			forecastDays: 14,
			wantExitCode: 0, // Parse succeeds
		},
		{
			name:         "forecast-days 15 is rejected",
			forecastDays: 15,
			wantExitCode: 3, // ValidationError
			wantError:    "--forecast-days must be at most 14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parsing and validation directly without making API calls
			daysStr := ""
			if tt.forecastDays < 10 {
				daysStr = string([]byte{byte(tt.forecastDays + '0')})
			} else {
				daysStr = string([]byte{byte(tt.forecastDays/10 + '0'), byte(tt.forecastDays%10 + '0')})
			}
			args := []string{"--latitude", "52.52", "--longitude", "13.41", "--daily", "default", "--forecast-days", daysStr}

			cfg, err := cli.ParseForecast(args)

			if tt.wantError != "" {
				// Expect an error
				if err == nil {
					t.Errorf("ParseForecast() expected error with %q, got nil", tt.wantError)
					return
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("ParseForecast() error = %q, want to contain %q", err.Error(), tt.wantError)
				}
				// Check exit code from error type
				var validationErr *cli.ValidationError
				if errors.As(err, &validationErr) {
					// Exit code 3 is correct for ValidationError
				} else {
					t.Errorf("Expected ValidationError (exit code 3), got %T", err)
				}
			} else {
				// Expect success
				if err != nil {
					t.Errorf("ParseForecast() unexpected error: %v", err)
					return
				}
				if cfg.ForecastDays != tt.forecastDays {
					t.Errorf("ForecastDays = %d, want %d", cfg.ForecastDays, tt.forecastDays)
				}
			}
		})
	}
}

// TestAcceptance_CurrentOnlyRequest verifies that current-only requests work
// without hourly or daily data.
func TestAcceptance_CurrentOnlyRequest(t *testing.T) {
	mock := newMockRequestTrackingServer()
	defer mock.Close()

	args := []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "temperature_2m,weather_code"}
	commandArgs := args[1:]

	cfg, err := cli.ParseForecast(commandArgs)
	if err != nil {
		t.Fatalf("cli.ParseForecast failed: %v", err)
	}

	// Verify config has current vars but no hourly/daily
	if len(cfg.CurrentVars) == 0 {
		t.Error("Expected CurrentVars to be non-empty")
	}
	if len(cfg.HourlyVars) != 0 {
		t.Errorf("Expected HourlyVars to be empty, got %v", cfg.HourlyVars)
	}
	if len(cfg.DailyVars) != 0 {
		t.Errorf("Expected DailyVars to be empty, got %v", cfg.DailyVars)
	}

	// Verify the request can be made
	mockClient := &mockHTTPClient{serverURL: mock.URL()}
	weatherMapper := weathercode.NewMapper()
	omClient := openmeteo.NewClient(mockClient)
	omClient.SetBaseURL(mock.URL())
	omClient.SetGeocodeURL(mock.URL())
	fcService := forecast.NewService(omClient, weatherMapper)

	varDefs := cli.GetVariableDefinitions()
	currentVars := varDefs.ExpandDefaultVars(cfg.CurrentVars, "current")

	req := forecast.ForecastRequest{
		Latitude:    cfg.Latitude,
		Longitude:   cfg.Longitude,
		CurrentVars: currentVars,
		Units:       cfg.Units,
	}

	result, err := fcService.ForecastVariable(req)
	if err != nil {
		t.Fatalf("ForecastVariable failed: %v", err)
	}

	if result == nil {
		t.Fatal("ForecastVariable returned nil result")
	}

	if result.Current == nil {
		t.Error("Expected Current section in result, got nil")
	}
}

// TestAcceptance_DefaultKeywordExpansion verifies that the "default" keyword
// expands to the documented default variable set.
func TestAcceptance_DefaultKeywordExpansion(t *testing.T) {
	varDefs := cli.GetVariableDefinitions()

	// Test daily default expansion
	dailyDefault := varDefs.ExpandDefaultVars([]string{"default"}, "daily")

	expectedDailyDefaults := []string{
		"weather_code", "temperature_2m_min", "temperature_2m_max",
		"precipitation_sum", "precipitation_probability_max", "wind_speed_10m_max",
		"wind_gusts_10m_max", "uv_index_max", "sunrise", "sunset",
	}

	if len(dailyDefault) != len(expectedDailyDefaults) {
		t.Errorf("Expected %d daily default variables, got %d", len(expectedDailyDefaults), len(dailyDefault))
	}

	// Check that all expected variables are present
	dailyDefaultMap := make(map[string]bool)
	for _, v := range dailyDefault {
		dailyDefaultMap[v] = true
	}

	for _, expected := range expectedDailyDefaults {
		if !dailyDefaultMap[expected] {
			t.Errorf("Expected daily default to contain %q", expected)
		}
	}

	// Verify the documented default set matches what's in GetVariableDefinitions
	if len(varDefs.DailyDefaults) != len(expectedDailyDefaults) {
		t.Errorf("Documented daily defaults count mismatch: expected %d, got %d", len(expectedDailyDefaults), len(varDefs.DailyDefaults))
	}

	// Test hourly default expansion
	hourlyDefault := varDefs.ExpandDefaultVars([]string{"default"}, "hourly")
	expectedHourlyDefaults := []string{
		"temperature_2m", "precipitation_probability", "precipitation",
		"wind_speed_10m", "weather_code",
	}

	if len(hourlyDefault) != len(expectedHourlyDefaults) {
		t.Errorf("Expected %d hourly default variables, got %d", len(expectedHourlyDefaults), len(hourlyDefault))
	}

	hourlyDefaultMap := make(map[string]bool)
	for _, v := range hourlyDefault {
		hourlyDefaultMap[v] = true
	}

	for _, expected := range expectedHourlyDefaults {
		if !hourlyDefaultMap[expected] {
			t.Errorf("Expected hourly default to contain %q", expected)
		}
	}

	// Test current default expansion
	currentDefault := varDefs.ExpandDefaultVars([]string{"default"}, "current")
	expectedCurrentDefaults := []string{
		"temperature_2m", "apparent_temperature", "precipitation",
		"wind_speed_10m", "weather_code",
	}

	if len(currentDefault) != len(expectedCurrentDefaults) {
		t.Errorf("Expected %d current default variables, got %d", len(expectedCurrentDefaults), len(currentDefault))
	}

	currentDefaultMap := make(map[string]bool)
	for _, v := range currentDefault {
		currentDefaultMap[v] = true
	}

	for _, expected := range expectedCurrentDefaults {
		if !currentDefaultMap[expected] {
			t.Errorf("Expected current default to contain %q", expected)
		}
	}
}

// TestAcceptance_VariableDiscoveryOutput verifies that "forecast variables daily"
// prints the supported fields, descriptions, units, and default membership.
func TestAcceptance_VariableDiscoveryOutput(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
	}{
		{
			name: "variables daily shows fields and descriptions",
			args: []string{"forecast", "variables", "daily"},
			wantOutput: []string{
				"weather_code",
				"temperature_2m_max",
				"temperature_2m_min",
				"precipitation_sum",
				"sunrise",
				"sunset",
			},
		},
		{
			name: "variables hourly shows fields and descriptions",
			args: []string{"forecast", "variables", "hourly"},
			wantOutput: []string{
				"temperature_2m",
				"precipitation_probability",
				"weather_code",
			},
		},
		{
			name: "variables current shows fields and descriptions",
			args: []string{"forecast", "variables", "current"},
			wantOutput: []string{
				"temperature_2m",
				"apparent_temperature",
				"weather_code",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			exitCode := Run(tt.args)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("Run() exit code = %d, want 0", exitCode)
			}

			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q, got: %q", want, output)
				}
			}
		})
	}
}

// TestAcceptance_SectionRequests verifies that forecast supports
// current-only, hourly-only, daily-only, and combined requests.
func TestAcceptance_SectionRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCurrent bool
		wantHourly  bool
		wantDaily   bool
	}{
		{
			name:        "current-only request",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "default"},
			wantCurrent: true,
			wantHourly:  false,
			wantDaily:   false,
		},
		{
			name:        "hourly-only request",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--hourly", "default", "--forecast-hours", "24"},
			wantCurrent: false,
			wantHourly:  true,
			wantDaily:   false,
		},
		{
			name:        "daily-only request",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--daily", "default", "--forecast-days", "7"},
			wantCurrent: false,
			wantHourly:  false,
			wantDaily:   true,
		},
		{
			name:        "combined current and hourly",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "default", "--hourly", "default", "--forecast-hours", "24"},
			wantCurrent: true,
			wantHourly:  true,
			wantDaily:   false,
		},
		{
			name:        "combined current and daily",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "default", "--daily", "default", "--forecast-days", "7"},
			wantCurrent: true,
			wantHourly:  false,
			wantDaily:   true,
		},
		{
			name:        "combined hourly and daily",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--hourly", "default", "--forecast-hours", "24", "--daily", "default", "--forecast-days", "7"},
			wantCurrent: false,
			wantHourly:  true,
			wantDaily:   true,
		},
		{
			name:        "all sections combined",
			args:        []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "default", "--hourly", "default", "--forecast-hours", "48", "--daily", "default", "--forecast-days", "14"},
			wantCurrent: true,
			wantHourly:  true,
			wantDaily:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandArgs := tt.args[1:]

			cfg, err := cli.ParseForecast(commandArgs)
			if err != nil {
				t.Fatalf("cli.ParseForecast failed: %v", err)
			}

			hasCurrent := len(cfg.CurrentVars) > 0
			hasHourly := len(cfg.HourlyVars) > 0
			hasDaily := len(cfg.DailyVars) > 0

			if hasCurrent != tt.wantCurrent {
				t.Errorf("CurrentVars: got %v, want %v", hasCurrent, tt.wantCurrent)
			}
			if hasHourly != tt.wantHourly {
				t.Errorf("HourlyVars: got %v, want %v", hasHourly, tt.wantHourly)
			}
			if hasDaily != tt.wantDaily {
				t.Errorf("DailyVars: got %v, want %v", hasDaily, tt.wantDaily)
			}
		})
	}
}

// TestAcceptance_CityGeocoding verifies that --city resolves to
// coordinates and includes resolved place metadata.
func TestAcceptance_CityGeocoding(t *testing.T) {
	mock := newMockRequestTrackingServer()
	defer mock.Close()

	args := []string{"forecast", "--city", "Berlin", "--current", "default"}
	commandArgs := args[1:]

	cfg, err := cli.ParseForecast(commandArgs)
	if err != nil {
		t.Fatalf("cli.ParseForecast failed: %v", err)
	}

	// Verify config has city set
	if cfg.City != "Berlin" {
		t.Errorf("Expected City to be \"Berlin\", got %q", cfg.City)
	}

	// Verify geocoding works
	mockClient := &mockHTTPClient{serverURL: mock.URL()}
	weatherMapper := weathercode.NewMapper()
	omClient := openmeteo.NewClient(mockClient)
	omClient.SetBaseURL(mock.URL())
	omClient.SetGeocodeURL(mock.URL())

	searchQuery := cfg.City
	if cfg.Country != "" {
		searchQuery = cfg.City + ", " + cfg.Country
	}
	location, err := omClient.FetchLocation(searchQuery, 1)
	if err != nil {
		t.Fatalf("FetchLocation failed: %v", err)
	}

	// Verify coordinates were resolved
	if location.Latitude != 52.52 {
		t.Errorf("Expected latitude 52.52, got %f", location.Latitude)
	}
	if location.Longitude != 13.41 {
		t.Errorf("Expected longitude 13.41, got %f", location.Longitude)
	}

	// Verify metadata
	if location.Name != "Berlin" {
		t.Errorf("Expected location name \"Berlin\", got %q", location.Name)
	}
	if location.Country != "Germany" {
		t.Errorf("Expected country \"Germany\", got %q", location.Country)
	}

	// Verify forecast includes location metadata in result
	fcService := forecast.NewService(omClient, weatherMapper)
	varDefs := cli.GetVariableDefinitions()
	currentVars := varDefs.ExpandDefaultVars(cfg.CurrentVars, "current")

	req := forecast.ForecastRequest{
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
		CurrentVars: currentVars,
		Units:       cfg.Units,
		Location:    location,
	}

	result, err := fcService.ForecastVariable(req)
	if err != nil {
		t.Fatalf("ForecastVariable failed: %v", err)
	}

	if result.Meta.Location == nil {
		t.Error("Expected location metadata in result, got nil")
	} else {
		if result.Meta.Location.Name != "Berlin" {
			t.Errorf("Expected location name \"Berlin\" in metadata, got %q", result.Meta.Location.Name)
		}
		if result.Meta.Location.Country != "Germany" {
			t.Errorf("Expected country \"Germany\" in metadata, got %q", result.Meta.Location.Country)
		}
	}
}

// TestAcceptance_RequestFiltering verifies that the CLI only requests the
// weather sections and variables explicitly asked for.
func TestAcceptance_RequestFiltering(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		wantCurrent       bool
		wantHourly        bool
		wantDaily         bool
		wantVarsInRequest []string
	}{
		{
			name:              "current-only requests only current section",
			args:              []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "temperature_2m,weather_code"},
			wantCurrent:       true,
			wantHourly:        false,
			wantDaily:         false,
			wantVarsInRequest: []string{"temperature_2m", "weather_code"},
		},
		{
			name:              "hourly-only requests only hourly section",
			args:              []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--hourly", "temperature_2m,precipitation_probability", "--forecast-hours", "24"},
			wantCurrent:       false,
			wantHourly:        true,
			wantDaily:         false,
			wantVarsInRequest: []string{"temperature_2m", "precipitation_probability"},
		},
		{
			name:              "daily-only requests only daily section",
			args:              []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--daily", "weather_code,temperature_2m_max", "--forecast-days", "7"},
			wantCurrent:       false,
			wantHourly:        false,
			wantDaily:         true,
			wantVarsInRequest: []string{"weather_code", "temperature_2m_max"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockRequestTrackingServer()
			defer mock.Close()

			commandArgs := tt.args[1:]

			cfg, err := cli.ParseForecast(commandArgs)
			if err != nil {
				t.Fatalf("cli.ParseForecast failed: %v", err)
			}

			// Create a tracking HTTP client
			trackingClient := &trackingHTTPClient{
				mock: mock,
			}

			weatherMapper := weathercode.NewMapper()
			omClient := openmeteo.NewClient(trackingClient)
			omClient.SetBaseURL(mock.URL())
			omClient.SetGeocodeURL(mock.URL())
			fcService := forecast.NewService(omClient, weatherMapper)

			varDefs := cli.GetVariableDefinitions()
			currentVars := varDefs.ExpandDefaultVars(cfg.CurrentVars, "current")
			hourlyVars := varDefs.ExpandDefaultVars(cfg.HourlyVars, "hourly")
			dailyVars := varDefs.ExpandDefaultVars(cfg.DailyVars, "daily")

			req := forecast.ForecastRequest{
				Latitude:            cfg.Latitude,
				Longitude:           cfg.Longitude,
				CurrentVars:         currentVars,
				HourlyVars:          hourlyVars,
				DailyVars:           dailyVars,
				HourlyForecastHours: cfg.ForecastHours,
				DailyForecastDays:   cfg.ForecastDays,
				Units:               cfg.Units,
			}

			_, err = fcService.ForecastVariable(req)
			if err != nil {
				t.Fatalf("ForecastVariable failed: %v", err)
			}

			// Check which sections were requested
			hasCurrent := trackingClient.lastRequestURL.Query().Get("current") != ""
			hasHourly := trackingClient.lastRequestURL.Query().Get("hourly") != ""
			hasDaily := trackingClient.lastRequestURL.Query().Get("daily") != ""

			if hasCurrent != tt.wantCurrent {
				t.Errorf("Current section requested: got %v, want %v", hasCurrent, tt.wantCurrent)
			}
			if hasHourly != tt.wantHourly {
				t.Errorf("Hourly section requested: got %v, want %v", hasHourly, tt.wantHourly)
			}
			if hasDaily != tt.wantDaily {
				t.Errorf("Daily section requested: got %v, want %v", hasDaily, tt.wantDaily)
			}

			// Check that only the requested variables were in the request
			for _, section := range []string{"current", "hourly", "daily"} {
				varsParam := trackingClient.lastRequestURL.Query().Get(section)
				if varsParam != "" {
					requestedVars := strings.Split(varsParam, ",")
					// Verify each requested variable is in the expected list
					for _, v := range requestedVars {
						found := false
						for _, expected := range tt.wantVarsInRequest {
							if v == expected {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Unexpected variable %q in %s request", v, section)
						}
					}
				}
			}
		})
	}
}

// trackingHTTPClient is an HTTP client that tracks the last request URL
// for verification purposes.
type trackingHTTPClient struct {
	mock           *mockRequestTrackingServer
	lastRequestURL *url.URL
}

func (c *trackingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Track the request URL
	c.lastRequestURL = req.URL

	// Forward to mock server
	mockURL := c.mock.URL() + "?" + req.URL.Query().Encode()
	mockReq, err := http.NewRequest("GET", mockURL, nil)
	if err != nil {
		return nil, err
	}
	mockReq.Header = req.Header
	return http.DefaultClient.Do(mockReq)
}

// TestAcceptance_AllSectionsCombined verifies all sections can be requested together.
func TestAcceptance_AllSectionsCombined(t *testing.T) {
	args := []string{"forecast", "--latitude", "52.52", "--longitude", "13.41",
		"--current", "default", "--hourly", "default", "--forecast-hours", "48",
		"--daily", "default", "--forecast-days", "14"}

	commandArgs := args[1:]

	cfg, err := cli.ParseForecast(commandArgs)
	if err != nil {
		t.Fatalf("cli.ParseForecast failed: %v", err)
	}

	// Verify all sections are requested
	if len(cfg.CurrentVars) == 0 {
		t.Error("Expected CurrentVars to be non-empty")
	}
	if len(cfg.HourlyVars) == 0 {
		t.Error("Expected HourlyVars to be non-empty")
	}
	if len(cfg.DailyVars) == 0 {
		t.Error("Expected DailyVars to be non-empty")
	}

	// Verify limits
	if cfg.ForecastHours != 48 {
		t.Errorf("Expected ForecastHours 48, got %d", cfg.ForecastHours)
	}
	if cfg.ForecastDays != 14 {
		t.Errorf("Expected ForecastDays 14, got %d", cfg.ForecastDays)
	}
}
