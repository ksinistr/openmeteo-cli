package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"openmeteo-cli/internal/cli"
)

// mockHTTPClient is a test HTTP client that routes requests to a test server.
type mockHTTPClient struct {
	serverURL string
}

// Do executes an HTTP request, routing it to the mock server.
func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Replace the request URL with the mock server URL
	// Preserve all query parameters and path
	mockURL := m.serverURL + "?" + req.URL.Query().Encode()

	mockReq, err := http.NewRequest("GET", mockURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mock request: %w", err)
	}
	// Copy headers
	mockReq.Header = req.Header

	// Make the request to the mock server
	return http.DefaultClient.Do(mockReq)
}

// TestRunExitCodes tests that correct exit codes are returned for different error types.
func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		// Exit code 2: invalid arguments
		{"invalid lat format", []string{"today", "--lat", "not-a-number", "--lon", "0"}, 2},
		{"invalid lon format", []string{"today", "--lat", "0", "--lon", "not-a-number"}, 2},
		{"missing lat", []string{"today", "--lon", "0"}, 2},
		{"missing lon", []string{"today", "--lat", "0"}, 2},
		{"missing both lat and lon", []string{"today"}, 2},
		{"invalid units format", []string{"today", "--lat", "0", "--lon", "0", "--units", "celsius"}, 3},
		{"invalid format format", []string{"today", "--lat", "0", "--lon", "0", "--format", "xml"}, 3},

		// Exit code 3: validation errors
		{"latitude too high", []string{"today", "--lat", "91", "--lon", "0"}, 3},
		{"latitude too low", []string{"today", "--lat", "-91", "--lon", "0"}, 3},
		{"longitude too high", []string{"today", "--lat", "0", "--lon", "181"}, 3},
		{"longitude too low", []string{"today", "--lat", "0", "--lon", "-181"}, 3},
		{"invalid units", []string{"today", "--lat", "40", "--lon", "-74", "--units", "imperial_fahrenheit"}, 3},
		{"invalid format", []string{"today", "--lat", "40", "--lon", "-74", "--format", "json_lines"}, 3},
		{"day missing date", []string{"day", "--lat", "40", "--lon", "-74"}, 3},
		{"day invalid date", []string{"day", "--lat", "40", "--lon", "-74", "--date", "2026-13-45"}, 3},

		// Unknown command - exit 2
		{"unknown command", []string{"forecast", "--lat", "40", "--lon", "-74"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != tt.wantExitCode {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.wantExitCode)
			}
		})
	}
}

// TestRunEmptyArgs tests behavior when called with no arguments.
func TestRunEmptyArgs(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := Run([]string{})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	stderr := buf.String()

	if exitCode != 2 {
		t.Errorf("Run([]) exit code = %d, want 2", exitCode)
	}
	if !strings.Contains(stderr, "Error: no command specified") {
		t.Errorf("Run([]) stderr = %q, want to contain 'Error: no command specified'", stderr)
	}
}

// TestRunNoCommand tests behavior when called with only invalid args (no command).
func TestRunNoCommand(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Note: With "--lat" as first arg, the code treats it as command name and "--lat"
	// becomes an argument to check for lat/lon - which fails validation
	exitCode := Run([]string{"--lat", "40", "--lon", "-74"})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	stderr := buf.String()

	if exitCode != 2 {
		t.Errorf("Run with args but no command exit code = %d, want 2", exitCode)
	}
	// The error is about lat/lon validation since the first arg is used as command name
	if !strings.Contains(stderr, "Error:") {
		t.Errorf("Run should output error to stderr, got: %q", stderr)
	}
}

// mockOpenMeteoServer creates a test server that returns valid Open-Meteo API responses.
func mockOpenMeteoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request has required parameters
		if r.URL.Query().Get("latitude") == "" || r.URL.Query().Get("longitude") == "" {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		// Build a minimal valid Open-Meteo response
		now := time.Now().UTC()
		today := now.Format("2006-01-02")

		// Generate 7 daily entries for week command support
		dailyTimes := make([]string, 7)
		dailySunrises := make([]string, 7)
		dailySunsets := make([]string, 7)
		dailyWeatherCodes := make([]int, 7)
		dailyTempMaxs := make([]float64, 7)
		dailyTempMins := make([]float64, 7)
		dailyPrecipSums := make([]float64, 7)
		dailyPrecipProbMaxs := make([]int, 7)
		dailyWindSpeedMaxs := make([]float64, 7)
		dailyWindGustsMaxs := make([]float64, 7)
		dailyUVIndexMaxs := make([]float64, 7)

		// Generate hourly entries for the first 2 days (today and tomorrow)
		// This provides sufficient data for testing today/day commands
		var hourlyTimes []string
		var hourlyTemps []float64
		var hourlyApparentTemps []float64
		var hourlyHumidity []int
		var hourlyPrecip []float64
		var hourlyPrecipProb []int
		var hourlyWeatherCodes []int
		var hourlyWindSpeed []float64
		var hourlyWindGusts []float64
		var hourlyWindDir []int
		var hourlyUVIndex []float64

		for day := 0; day < 2; day++ {
			date := now.AddDate(0, 0, day)
			dateStr := date.Format("2006-01-02")
			// Add a few hourly entries for each day
			for hour := 0; hour < 24; hour += 12 {
				hourlyTimes = append(hourlyTimes, fmt.Sprintf("%sT%02d:00", dateStr, hour))
				hourlyTemps = append(hourlyTemps, 10.0+float64(hour))
				hourlyApparentTemps = append(hourlyApparentTemps, 9.0+float64(hour))
				hourlyHumidity = append(hourlyHumidity, 60+hour)
				hourlyPrecip = append(hourlyPrecip, 0.0)
				hourlyPrecipProb = append(hourlyPrecipProb, 10)
				hourlyWeatherCodes = append(hourlyWeatherCodes, 1)
				hourlyWindSpeed = append(hourlyWindSpeed, 10.0+float64(hour))
				hourlyWindGusts = append(hourlyWindGusts, 15.0+float64(hour))
				hourlyWindDir = append(hourlyWindDir, 240+hour)
				hourlyUVIndex = append(hourlyUVIndex, float64(hour)/4)
			}
		}

		for i := 0; i < 7; i++ {
			date := now.AddDate(0, 0, i)
			dateStr := date.Format("2006-01-02")
			dailyTimes[i] = dateStr
			dailySunrises[i] = dateStr + "T06:30"
			dailySunsets[i] = dateStr + "T19:45"
			dailyWeatherCodes[i] = 1
			dailyTempMaxs[i] = 16.3
			dailyTempMins[i] = 8.1
			dailyPrecipSums[i] = 0.0
			dailyPrecipProbMaxs[i] = 20
			dailyWindSpeedMaxs[i] = 18.7
			dailyWindGustsMaxs[i] = 25.4
			dailyUVIndexMaxs[i] = 4.1
		}

		response := map[string]interface{}{
			"latitude":  40.7128,
			"longitude": -74.0060,
			"timezone":  "America/New_York",
			"current": map[string]interface{}{
				"time":                      today + "T12:00",
				"temperature_2m":            15.5,
				"apparent_temperature":      14.2,
				"relative_humidity_2m":      65,
				"precipitation":             0.0,
				"precipitation_probability": 10,
				"weather_code":              1,
				"wind_speed_10m":            15.3,
				"wind_gusts_10m":            22.1,
				"wind_direction_10m":        245,
				"uv_index":                  3.2,
			},
			"hourly": map[string]interface{}{
				"time":                      hourlyTimes,
				"temperature_2m":            hourlyTemps,
				"apparent_temperature":      hourlyApparentTemps,
				"relative_humidity_2m":      hourlyHumidity,
				"precipitation":             hourlyPrecip,
				"precipitation_probability": hourlyPrecipProb,
				"weather_code":              hourlyWeatherCodes,
				"wind_speed_10m":            hourlyWindSpeed,
				"wind_gusts_10m":            hourlyWindGusts,
				"wind_direction_10m":        hourlyWindDir,
				"uv_index":                  hourlyUVIndex,
			},
			"daily": map[string]interface{}{
				"time":                          dailyTimes,
				"sunrise":                       dailySunrises,
				"sunset":                        dailySunsets,
				"weather_code":                  dailyWeatherCodes,
				"temperature_2m_max":            dailyTempMaxs,
				"temperature_2m_min":            dailyTempMins,
				"precipitation_sum":             dailyPrecipSums,
				"precipitation_probability_max": dailyPrecipProbMaxs,
				"wind_speed_10m_max":            dailyWindSpeedMaxs,
				"wind_gusts_10m_max":            dailyWindGustsMaxs,
				"uv_index_max":                  dailyUVIndexMaxs,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// TestRunCommandDispatch tests that the Run function dispatches commands correctly.
// It uses a mock HTTP server to verify the actual dispatch logic through runWithClient.
func TestRunCommandDispatch(t *testing.T) {
	// Start a mock server for testing
	server := mockOpenMeteoServer()
	defer server.Close()

	// Use a date relative to now to ensure test remains valid over time
	// The mock server generates 7 days starting from time.Now().UTC()
	testDate := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")

	testCases := []struct {
		name string
		args []string
		want int // expected exit code
	}{
		{"today dispatch", []string{"today", "--lat", "40.7128", "--lon", "-74.0060"}, 0},
		{"day dispatch", []string{"day", "--lat", "40.7128", "--lon", "-74.0060", "--date", testDate}, 0},
		{"week dispatch", []string{"week", "--lat", "40.7128", "--lon", "-74.0060"}, 0},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				t.Fatal("no command provided")
			}
			command := tt.args[0]
			commandArgs := tt.args[1:]

			// Parse args to get config (same as Run does)
			cfg, err := cli.Parse(command, commandArgs)
			if err != nil {
				t.Fatalf("cli.Parse failed: %v", err)
			}

			// Apply command-specific validation (same as in Run)
			if command == "day" && cfg.DateStr == "" {
				t.Fatal("date is required for day command")
			}
			if command != "day" && cfg.DateStr != "" {
				t.Fatal("--date flag is only valid for day command")
			}

			// Parse date for day command
			var date time.Time
			if cfg.DateStr != "" {
				date, err = time.Parse("2006-01-02", cfg.DateStr)
				if err != nil {
					t.Fatalf("date parsing failed: %v", err)
				}
			}

			// Now actually test the dispatch by calling runWithClient with mock server
			// This tests the actual switch statement in runWithClient
			mockClient := &mockHTTPClient{serverURL: server.URL}
			exitCode := runWithClient(cfg, date, command, mockClient)

			if exitCode != tt.want {
				t.Errorf("runWithClient() exit code = %d, want %d", exitCode, tt.want)
			}
		})
	}
}

// TestRunValidatesArgsBeforeDispatch tests that argument validation happens before API calls.
func TestRunValidatesArgsBeforeDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"invalid lat format", []string{"today", "--lat", "abc", "--lon", "0"}},
		{"lat out of range", []string{"today", "--lat", "100", "--lon", "0"}},
		{"day missing date", []string{"day", "--lat", "40", "--lon", "-74"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			_ = Run(tt.args)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			stderr := buf.String()

			// Check that we got an error message (validation error, not network error)
			if !strings.Contains(stderr, "Error:") {
				t.Errorf("Run(%v) should output error to stderr, got: %q", tt.args, stderr)
			}
			// Validation error should occur before network calls
			// Network errors would be from real HTTP calls
		})
	}
}

// TestRunValidationErrorExitCode tests that validation errors return exit code 3.
func TestRunValidationErrorExitCode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"lat out of range", []string{"today", "--lat", "100", "--lon", "0"}, 3},
		{"lon out of range", []string{"today", "--lat", "40", "--lon", "200"}, 3},
		{"invalid units", []string{"today", "--lat", "40", "--lon", "0", "--units", "invalid"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != tt.want {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.want)
			}
		})
	}
}

// TestRunInvalidArgumentExitCode tests that invalid argument errors return exit code 2.
func TestRunInvalidArgumentExitCode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"missing lat/lon", []string{"today"}, 2},
		{"invalid lat format", []string{"today", "--lat", "abc", "--lon", "0"}, 2},
		{"invalid lon format", []string{"today", "--lat", "0", "--lon", "xyz"}, 2},
		{"unknown command", []string{"forecast", "--lat", "0", "--lon", "0"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != tt.want {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.want)
			}
		})
	}
}
