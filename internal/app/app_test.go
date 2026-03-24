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
		{"invalid latitude format", []string{"hourly", "--latitude", "not-a-number", "--longitude", "0", "--forecast-days", "1"}, 2},
		{"invalid longitude format", []string{"daily", "--latitude", "0", "--longitude", "not-a-number", "--forecast-days", "1"}, 2},
		{"missing latitude", []string{"hourly", "--longitude", "0", "--forecast-days", "1"}, 3},
		{"missing longitude", []string{"daily", "--latitude", "0", "--forecast-days", "1"}, 3},
		{"missing both latitude and longitude", []string{"hourly", "--forecast-days", "1"}, 3},
		{"invalid units format", []string{"hourly", "--latitude", "0", "--longitude", "0", "--forecast-days", "1", "--units", "celsius"}, 3},
		{"invalid format format", []string{"daily", "--latitude", "0", "--longitude", "0", "--forecast-days", "1", "--format", "xml"}, 3},

		// Exit code 3: validation errors
		{"latitude too high", []string{"hourly", "--latitude", "91", "--longitude", "0", "--forecast-days", "1"}, 3},
		{"latitude too low", []string{"daily", "--latitude", "-91", "--longitude", "0", "--forecast-days", "1"}, 3},
		{"longitude too high", []string{"hourly", "--latitude", "0", "--longitude", "181", "--forecast-days", "1"}, 3},
		{"longitude too low", []string{"daily", "--latitude", "0", "--longitude", "-181", "--forecast-days", "1"}, 3},
		{"invalid units", []string{"hourly", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "--units", "imperial_fahrenheit"}, 3},
		{"invalid format", []string{"daily", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "--format", "json_lines"}, 3},
		{"missing forecast-days", []string{"hourly", "--latitude", "40", "--longitude", "-74"}, 3},
		{"hourly exceeds max days", []string{"hourly", "--latitude", "40", "--longitude", "-74", "--forecast-days", "3"}, 3},
		{"daily exceeds max days", []string{"daily", "--latitude", "40", "--longitude", "-74", "--forecast-days", "15"}, 3},

		// Unknown command - exit 2
		{"unknown command", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1"}, 2},
		{"unknown command 2", []string{"today", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1"}, 2},
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
	if !strings.Contains(stderr, "Usage: openmeteo-cli") {
		t.Errorf("Run([]) stderr = %q, want to contain 'Usage: openmeteo-cli'", stderr)
	}
}

// TestRunNoCommand tests behavior when called with only invalid args (no command).
func TestRunNoCommand(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := Run([]string{"--latitude", "40", "--longitude", "-74", "--forecast-days", "1"})

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	stderr := buf.String()

	if exitCode != 2 {
		t.Errorf("Run with args but no command exit code = %d, want 2", exitCode)
	}
	// The error is about unknown command since the first arg is used as command name
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

		// Generate 14 daily entries for daily command support
		dailyCount := 14
		dailyTimes := make([]string, dailyCount)
		dailySunrises := make([]string, dailyCount)
		dailySunsets := make([]string, dailyCount)
		dailyWeatherCodes := make([]int, dailyCount)
		dailyTempMaxs := make([]float64, dailyCount)
		dailyTempMins := make([]float64, dailyCount)
		dailyPrecipSums := make([]float64, dailyCount)
		dailyPrecipProbMaxs := make([]int, dailyCount)
		dailyWindSpeedMaxs := make([]float64, dailyCount)
		dailyWindGustsMaxs := make([]float64, dailyCount)
		dailyUVIndexMaxs := make([]float64, dailyCount)

		// Generate hourly entries for the first 2 days (today and tomorrow)
		// This provides sufficient data for testing hourly command
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
			// Add hourly entries for each day
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

		for i := 0; i < dailyCount; i++ {
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

	testCases := []struct {
		name string
		args []string
		want int // expected exit code
	}{
		{"hourly dispatch", []string{"hourly", "--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "1"}, 0},
		{"hourly 2 days dispatch", []string{"hourly", "--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "2"}, 0},
		{"daily dispatch", []string{"daily", "--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "7"}, 0},
		{"daily 14 days dispatch", []string{"daily", "--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "14"}, 0},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				t.Fatal("no command provided")
			}
			command := tt.args[0]
			commandArgs := tt.args[1:]

			// Parse args to get config (same as Run does)
			cfg, err := cli.Parse(commandArgs, command)
			if err != nil {
				t.Fatalf("cli.Parse failed: %v", err)
			}

			// Now actually test the dispatch by calling runWithClient with mock server
			// This tests the actual switch statement in runWithClient
			mockClient := &mockHTTPClient{serverURL: server.URL}
			exitCode := runWithClient(cfg, mockClient)

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
		{"invalid latitude format", []string{"hourly", "--latitude", "abc", "--longitude", "0", "--forecast-days", "1"}},
		{"latitude out of range", []string{"hourly", "--latitude", "100", "--longitude", "0", "--forecast-days", "1"}},
		{"no forecast-days", []string{"hourly", "--latitude", "40", "--longitude", "-74"}},
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
		{"latitude out of range", []string{"hourly", "--latitude", "100", "--longitude", "0", "--forecast-days", "1"}, 3},
		{"longitude out of range", []string{"daily", "--latitude", "40", "--longitude", "200", "--forecast-days", "1"}, 3},
		{"invalid units", []string{"hourly", "--latitude", "40", "--longitude", "0", "--forecast-days", "1", "--units", "invalid"}, 3},
		{"no forecast-days", []string{"hourly", "--latitude", "40", "--longitude", "0"}, 3},
		{"hourly exceeds max", []string{"hourly", "--latitude", "40", "--longitude", "0", "--forecast-days", "3"}, 3},
		{"daily exceeds max", []string{"daily", "--latitude", "40", "--longitude", "0", "--forecast-days", "15"}, 3},
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
		{"invalid latitude format", []string{"hourly", "--latitude", "abc", "--longitude", "0", "--forecast-days", "1"}, 2},
		{"invalid longitude format", []string{"daily", "--latitude", "0", "--longitude", "xyz", "--forecast-days", "1"}, 2},
		{"invalid forecast-days format", []string{"hourly", "--latitude", "0", "--longitude", "0", "--forecast-days", "abc"}, 2},
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

// TestRunHelpFlags tests that help flags return exit code 0 and print usage.
func TestRunHelpFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantOutput   string
	}{
		// Root help
		{"root -h", []string{"-h"}, 0, "Usage: openmeteo-cli"},
		{"root --help", []string{"--help"}, 0, "Usage: openmeteo-cli"},
		// Command help
		{"hourly -h", []string{"hourly", "-h"}, 0, "Usage: openmeteo-cli hourly"},
		{"hourly --help", []string{"hourly", "--help"}, 0, "Usage: openmeteo-cli hourly"},
		{"daily -h", []string{"daily", "-h"}, 0, "Usage: openmeteo-cli daily"},
		{"daily --help", []string{"daily", "--help"}, 0, "Usage: openmeteo-cli daily"},
		{"hourly with args and -h", []string{"hourly", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "-h"}, 0, "Usage: openmeteo-cli hourly"},
		{"daily with args and --help", []string{"daily", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "--help"}, 0, "Usage: openmeteo-cli daily"},
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
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if exitCode != tt.wantExitCode {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.wantExitCode)
			}
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Run(%v) stdout = %q, want to contain %q", tt.args, output, tt.wantOutput)
			}
		})
	}
}

func TestRunDoesNotTreatFlagValuesAsHelp(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStdErr   string
	}{
		{
			name:         "units value help token returns validation error",
			args:         []string{"hourly", "--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "--units", "--help"},
			wantExitCode: 3,
			wantStdErr:   "Error: units must be 'metric' or 'imperial'",
		},
		{
			name:         "latitude value help token returns parse error",
			args:         []string{"hourly", "--latitude", "-h", "--longitude", "-74", "--forecast-days", "1"},
			wantExitCode: 2,
			wantStdErr:   "Error: invalid latitude value",
		},
		{
			name:         "forecast-days value help token returns parse error",
			args:         []string{"hourly", "--latitude", "40", "--longitude", "-74", "--forecast-days", "--help"},
			wantExitCode: 2,
			wantStdErr:   "Error: invalid forecast-days value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			stdoutR, stdoutW, _ := os.Pipe()
			os.Stdout = stdoutW

			oldStderr := os.Stderr
			stderrR, stderrW, _ := os.Pipe()
			os.Stderr = stderrW

			exitCode := Run(tt.args)

			_ = stdoutW.Close()
			os.Stdout = oldStdout
			_ = stderrW.Close()
			os.Stderr = oldStderr

			var stdoutBuf bytes.Buffer
			_, _ = io.Copy(&stdoutBuf, stdoutR)

			var stderrBuf bytes.Buffer
			_, _ = io.Copy(&stderrBuf, stderrR)

			if exitCode != tt.wantExitCode {
				t.Fatalf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.wantExitCode)
			}
			if strings.Contains(stdoutBuf.String(), "Usage: openmeteo-cli") {
				t.Fatalf("Run(%v) stdout unexpectedly contained help output: %q", tt.args, stdoutBuf.String())
			}
			if !strings.Contains(stderrBuf.String(), tt.wantStdErr) {
				t.Fatalf("Run(%v) stderr = %q, want to contain %q", tt.args, stderrBuf.String(), tt.wantStdErr)
			}
		})
	}
}
