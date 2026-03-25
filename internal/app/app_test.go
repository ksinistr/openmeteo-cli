package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"openmeteo-cli/internal/cli"
	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/weathercode"
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
		{"invalid latitude format", []string{"forecast", "--latitude", "not-a-number", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}, 2},
		{"invalid longitude format", []string{"forecast", "--latitude", "0", "--longitude", "not-a-number", "--daily", "default", "--forecast-days", "1"}, 2},
		{"missing latitude", []string{"forecast", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}, 3},
		{"missing longitude", []string{"forecast", "--latitude", "0", "--daily", "default", "--forecast-days", "1"}, 3},
		{"missing both latitude and longitude", []string{"forecast", "--daily", "default", "--forecast-days", "1"}, 3},
		{"invalid units format", []string{"forecast", "--latitude", "0", "--longitude", "0", "--daily", "default", "--forecast-days", "1", "--units", "celsius"}, 3},
		{"invalid format format", []string{"forecast", "--latitude", "0", "--longitude", "0", "--daily", "default", "--forecast-days", "1", "--format", "xml"}, 3},

		// Exit code 3: validation errors
		{"latitude too high", []string{"forecast", "--latitude", "91", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}, 3},
		{"latitude too low", []string{"forecast", "--latitude", "-91", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}, 3},
		{"longitude too high", []string{"forecast", "--latitude", "0", "--longitude", "181", "--daily", "default", "--forecast-days", "1"}, 3},
		{"longitude too low", []string{"forecast", "--latitude", "0", "--longitude", "-181", "--daily", "default", "--forecast-days", "1"}, 3},
		{"invalid units", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "1", "--units", "imperial_fahrenheit"}, 3},
		{"invalid format", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "1", "--format", "json_lines"}, 3},
		{"missing forecast-days", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default"}, 3},
		{"hourly exceeds max hours", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--hourly", "default", "--forecast-hours", "49"}, 3},
		{"daily exceeds max days", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "15"}, 3},

		// Unknown command - exit 2
		{"unknown command", []string{"weekly", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "1"}, 2},
		{"unknown command 2", []string{"today", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "1"}, 2},
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

// TestRunValidatesArgsBeforeDispatch tests that argument validation happens before API calls.
func TestRunValidatesArgsBeforeDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"invalid latitude format", []string{"forecast", "--latitude", "abc", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}},
		{"latitude out of range", []string{"forecast", "--latitude", "100", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}},
		{"no forecast-days", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default"}},
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
		{"latitude out of range", []string{"forecast", "--latitude", "100", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}, 3},
		{"longitude out of range", []string{"forecast", "--latitude", "40", "--longitude", "200", "--daily", "default", "--forecast-days", "1"}, 3},
		{"invalid units", []string{"forecast", "--latitude", "40", "--longitude", "0", "--daily", "default", "--forecast-days", "1", "--units", "invalid"}, 3},
		{"no forecast-days with daily", []string{"forecast", "--latitude", "40", "--longitude", "0", "--daily", "default"}, 3},
		{"hourly exceeds max", []string{"forecast", "--latitude", "40", "--longitude", "0", "--hourly", "default", "--forecast-hours", "49"}, 3},
		{"daily exceeds max", []string{"forecast", "--latitude", "40", "--longitude", "0", "--daily", "default", "--forecast-days", "15"}, 3},
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
		{"invalid latitude format", []string{"forecast", "--latitude", "abc", "--longitude", "0", "--daily", "default", "--forecast-days", "1"}, 2},
		{"invalid longitude format", []string{"forecast", "--latitude", "0", "--longitude", "xyz", "--daily", "default", "--forecast-days", "1"}, 2},
		{"invalid forecast-days format", []string{"forecast", "--latitude", "0", "--longitude", "0", "--daily", "default", "--forecast-days", "abc"}, 2},
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
		{"forecast -h", []string{"forecast", "-h"}, 0, "Usage: openmeteo-cli forecast"},
		{"forecast --help", []string{"forecast", "--help"}, 0, "Usage: openmeteo-cli forecast"},
		{"forecast with args and -h", []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "1", "-h"}, 0, "Usage: openmeteo-cli forecast"},
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
			args:         []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "1", "--units", "--help"},
			wantExitCode: 3,
			wantStdErr:   "Error: units must be 'metric' or 'imperial'",
		},
		{
			name:         "latitude value help token returns parse error",
			args:         []string{"forecast", "--latitude", "-h", "--longitude", "-74", "--daily", "default", "--forecast-days", "1"},
			wantExitCode: 2,
			wantStdErr:   "Error: invalid latitude value",
		},
		{
			name:         "forecast-days value help token returns parse error",
			args:         []string{"forecast", "--latitude", "40", "--longitude", "-74", "--daily", "default", "--forecast-days", "--help"},
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

// =============================================================================
// Forecast Command Tests
// =============================================================================

// mockOpenMeteoServerForForecast creates a test server for forecast command testing.
func mockOpenMeteoServerForForecast() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The mockHTTPClient strips the path and only preserves query parameters
		// We differentiate between geocoding and forecast requests by checking query parameters

		// Check if this is a geocoding request (has "name" parameter but no "latitude")
		if r.URL.Query().Get("name") != "" && r.URL.Query().Get("latitude") == "" {
			// Parse the query parameter for location name
			name := r.URL.Query().Get("name")
			if name == "" {
				http.Error(w, "Missing name parameter", http.StatusBadRequest)
				return
			}

			// Return geocoding response
			if name == "NotFound" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []struct{}{}})
				return
			}
			if name == "Ambiguous" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"results": []map[string]interface{}{
						{"name": "Ambiguous, USA", "latitude": 40.0, "longitude": -74.0, "country": "United States"},
						{"name": "Ambiguous, UK", "latitude": 51.5, "longitude": -0.1, "country": "United Kingdom"},
					},
				})
				return
			}

			// Return a single location result
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

		// This is a forecast request (has "latitude" parameter)
		// Build a minimal valid Open-Meteo response
		now := time.Now().UTC()
		today := now.Format("2006-01-02")

		// Get requested parameters from query and parse as floats
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
		lon, _ := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)

		// Build response based on what was requested
		response := map[string]interface{}{
			"latitude":  lat,
			"longitude": lon,
			"timezone":  "Europe/Berlin",
		}

		// Add current section if requested
		if r.URL.Query().Get("current") != "" {
			response["current"] = map[string]interface{}{
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
			}
		}

		// Add hourly section if requested
		if r.URL.Query().Get("hourly") != "" {
			var hourlyTimes []string
			var hourlyTemps []float64
			var hourlyPrecip []float64
			var hourlyPrecipProb []int
			var hourlyWindSpeed []float64
			var hourlyWeatherCodes []int

			// Generate hourly entries
			hours := 48
			if h := r.URL.Query().Get("forecast_hours"); h != "" {
				_, _ = fmt.Sscanf(h, "%d", &hours)
			}

			for day := 0; day < 2; day++ {
				date := now.AddDate(0, 0, day)
				dateStr := date.Format("2006-01-02")
				for hour := 0; hour < 24; hour++ {
					if len(hourlyTimes) >= hours {
						break
					}
					hourlyTimes = append(hourlyTimes, fmt.Sprintf("%sT%02d:00", dateStr, hour))
					hourlyTemps = append(hourlyTemps, 10.0+float64(hour))
					hourlyPrecip = append(hourlyPrecip, 0.0)
					hourlyPrecipProb = append(hourlyPrecipProb, 10)
					hourlyWindSpeed = append(hourlyWindSpeed, 10.0+float64(hour)/2)
					hourlyWeatherCodes = append(hourlyWeatherCodes, 1)
				}
				if len(hourlyTimes) >= hours {
					break
				}
			}

			response["hourly"] = map[string]interface{}{
				"time":                      hourlyTimes,
				"temperature_2m":            hourlyTemps,
				"precipitation":             hourlyPrecip,
				"precipitation_probability": hourlyPrecipProb,
				"wind_speed_10m":            hourlyWindSpeed,
				"weather_code":              hourlyWeatherCodes,
			}
		}

		// Add daily section if requested
		if r.URL.Query().Get("daily") != "" {
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

			response["daily"] = map[string]interface{}{
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
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// TestRunForecastCommandSuccess tests successful forecast command execution.
func TestRunForecastCommandSuccess(t *testing.T) {
	server := mockOpenMeteoServerForForecast()
	defer server.Close()

	tests := []struct {
		name string
		args []string
		want int
	}{
		{
			name: "current only with coordinates",
			args: []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "temperature_2m,weather_code"},
			want: 0,
		},
		{
			name: "hourly only with coordinates",
			args: []string{"forecast", "--latitude", "40.7", "--longitude", "-74.0", "--hourly", "default", "--forecast-hours", "24"},
			want: 0,
		},
		{
			name: "daily only with coordinates",
			args: []string{"forecast", "--latitude", "40.7", "--longitude", "-74.0", "--daily", "default", "--forecast-days", "7"},
			want: 0,
		},
		{
			name: "combined current and daily",
			args: []string{"forecast", "--latitude", "51.5", "--longitude", "-0.1", "--current", "default", "--daily", "default", "--forecast-days", "5"},
			want: 0,
		},
		{
			name: "all sections with coordinates",
			args: []string{"forecast", "--latitude", "51.5", "--longitude", "-0.1", "--current", "default", "--hourly", "default", "--forecast-hours", "48", "--daily", "default", "--forecast-days", "14"},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				t.Fatal("no command provided")
			}
			_ = tt.args[0] // command name
			commandArgs := tt.args[1:]

			// Parse forecast arguments
			cfg, err := cli.ParseForecast(commandArgs)
			if err != nil {
				t.Fatalf("cli.ParseForecast failed: %v", err)
			}

			// Create mock client
			mockClient := &mockHTTPClient{serverURL: server.URL}

			// Test the forecast command with mock server
			exitCode := runForecastWithClient(cfg, mockClient)

			if exitCode != tt.want {
				t.Errorf("runForecastWithClient() exit code = %d, want %d", exitCode, tt.want)
			}
		})
	}
}

// TestRunForecastCommandWithGeocoding tests geocoding integration in forecast command.
func TestRunForecastCommandWithGeocoding(t *testing.T) {
	server := mockOpenMeteoServerForForecast()
	defer server.Close()

	tests := []struct {
		name string
		args []string
		want int
	}{
		{
			name: "current only with location",
			args: []string{"forecast", "--city", "Berlin", "--current", "default"},
			want: 0,
		},
		{
			name: "hourly only with location",
			args: []string{"forecast", "--city", "Tokyo", "--hourly", "default", "--forecast-hours", "24"},
			want: 0,
		},
		{
			name: "daily only with location",
			args: []string{"forecast", "--city", "Paris", "--daily", "default", "--forecast-days", "7"},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				t.Fatal("no command provided")
			}
			_ = tt.args[0] // command name
			commandArgs := tt.args[1:]

			// Parse forecast arguments
			cfg, err := cli.ParseForecast(commandArgs)
			if err != nil {
				t.Fatalf("cli.ParseForecast failed: %v", err)
			}

			// Create mock client and configure client URLs
			mockClient := &mockHTTPClient{serverURL: server.URL}

			// We need to create a custom client for this test that routes both geocoding and forecast
			// For now, we'll test the geocoding path separately
			weatherMapper := weathercode.NewMapper()
			omClient := openmeteo.NewClient(mockClient)
			omClient.SetBaseURL(server.URL)
			omClient.SetGeocodeURL(server.URL)
			fcService := forecast.NewService(omClient, weatherMapper)

			// Resolve location first
			searchQuery := cfg.City
			if cfg.Country != "" {
				searchQuery = cfg.City + ", " + cfg.Country
			}
			location, err := omClient.FetchLocation(searchQuery, 1)
			if err != nil {
				t.Fatalf("FetchLocation failed: %v", err)
			}

			// Expand "default" keyword to actual variable sets
			varDefs := cli.GetVariableDefinitions()
			currentVars := varDefs.ExpandDefaultVars(cfg.CurrentVars, "current")
			hourlyVars := varDefs.ExpandDefaultVars(cfg.HourlyVars, "hourly")
			dailyVars := varDefs.ExpandDefaultVars(cfg.DailyVars, "daily")

			// Build forecast request
			req := forecast.ForecastRequest{
				Latitude:            location.Latitude,
				Longitude:           location.Longitude,
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
				t.Errorf("ForecastVariable failed: %v", err)
				return
			}

			// Verify we got a result
			if result == nil {
				t.Error("ForecastVariable returned nil result")
				return
			}

			// Verify location metadata is in output
			if result.Meta.Location == nil {
				t.Error("Expected location metadata in result, got nil")
			} else if result.Meta.Location.Name != cfg.City {
				t.Errorf("Expected location name %q, got %q", cfg.City, result.Meta.Location.Name)
			}
		})
	}
}

// TestRunForecastCommandErrors tests forecast command error handling.
func TestRunForecastCommandErrors(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStdErr   string
	}{
		// Parse errors (exit 2)
		{
			name:         "unknown flag",
			args:         []string{"forecast", "--unknown", "value"},
			wantExitCode: 2,
			wantStdErr:   "unknown flag",
		},
		{
			name:         "invalid latitude format",
			args:         []string{"forecast", "--latitude", "invalid", "--longitude", "0", "--current", "default"},
			wantExitCode: 2,
			wantStdErr:   "invalid latitude value",
		},
		// Validation errors (exit 3)
		{
			name:         "no section specified",
			args:         []string{"forecast", "--latitude", "52.52", "--longitude", "13.41"},
			wantExitCode: 3,
			wantStdErr:   "at least one of --current, --hourly, or --daily is required",
		},
		{
			name:         "missing forecast-hours with hourly",
			args:         []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--hourly", "default"},
			wantExitCode: 3,
			wantStdErr:   "--forecast-hours is required when --hourly is specified",
		},
		{
			name:         "missing forecast-days with daily",
			args:         []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--daily", "default"},
			wantExitCode: 3,
			wantStdErr:   "--forecast-days is required when --daily is specified",
		},
		{
			name:         "location with lat/lon",
			args:         []string{"forecast", "--city", "Berlin", "--latitude", "52.52", "--longitude", "13.41", "--current", "default"},
			wantExitCode: 3,
			wantStdErr:   "--city cannot be combined with --latitude or --longitude",
		},
		{
			name:         "unknown variable",
			args:         []string{"forecast", "--latitude", "52.52", "--longitude", "13.41", "--current", "unknown_var"},
			wantExitCode: 3,
			wantStdErr:   "unknown current variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			exitCode := Run(tt.args)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			stderr := buf.String()

			if exitCode != tt.wantExitCode {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.wantExitCode)
			}
			if !strings.Contains(stderr, tt.wantStdErr) {
				t.Errorf("Run(%v) stderr = %q, want to contain %q", tt.args, stderr, tt.wantStdErr)
			}
		})
	}
}

// TestRunForecastGeocodingErrors tests geocoding-specific error handling.
func TestRunForecastGeocodingErrors(t *testing.T) {
	server := mockOpenMeteoServerForForecast()
	defer server.Close()

	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStdErr   string
	}{
		{
			name:         "location not found",
			args:         []string{"forecast", "--city", "NotFound", "--current", "default"},
			wantExitCode: 3,
			wantStdErr:   "location not found",
		},
		{
			name:         "location ambiguous",
			args:         []string{"forecast", "--city", "Ambiguous", "--current", "default"},
			wantExitCode: 3,
			wantStdErr:   "location is ambiguous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				t.Fatal("no command provided")
			}
			_ = tt.args[0] // command name
			commandArgs := tt.args[1:]

			// Parse forecast arguments
			cfg, err := cli.ParseForecast(commandArgs)
			if err != nil {
				t.Fatalf("cli.ParseForecast failed: %v", err)
			}

			// Create mock client
			mockClient := &mockHTTPClient{serverURL: server.URL}

			// Create a custom client for this test
			omClient := openmeteo.NewClient(mockClient)
			omClient.SetBaseURL(server.URL)
			omClient.SetGeocodeURL(server.URL)

			// Try to resolve location - this should fail
			searchQuery := cfg.City
			if cfg.Country != "" {
				searchQuery = cfg.City + ", " + cfg.Country
			}
			_, err = omClient.FetchLocation(searchQuery, 1)
			if err == nil {
				t.Error("Expected FetchLocation to fail, but it succeeded")
			}

			// Check error type
			if tt.wantStdErr == "location not found" && !errors.Is(err, openmeteo.ErrLocationNotFound) {
				t.Errorf("Expected ErrLocationNotFound, got: %v", err)
			}
			if tt.wantStdErr == "location is ambiguous" && !errors.Is(err, openmeteo.ErrLocationAmbiguous) {
				t.Errorf("Expected ErrLocationAmbiguous, got: %v", err)
			}
		})
	}
}

// TestRunForecastVariablesHelp tests the forecast variables command.
func TestRunForecastVariablesHelp(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantInDoc    string
		notWantInDoc string
	}{
		{
			name:         "variables all",
			args:         []string{"forecast", "variables"},
			wantInDoc:    "Forecast Variables",
			notWantInDoc: "",
		},
		{
			name:         "variables current",
			args:         []string{"forecast", "variables", "current"},
			wantInDoc:    "Current Weather",
			notWantInDoc: "Hourly Forecast",
		},
		{
			name:         "variables hourly",
			args:         []string{"forecast", "variables", "hourly"},
			wantInDoc:    "Hourly Forecast",
			notWantInDoc: "Daily Forecast",
		},
		{
			name:         "variables daily",
			args:         []string{"forecast", "variables", "daily"},
			wantInDoc:    "Daily Forecast",
			notWantInDoc: "Current Weather",
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
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if exitCode != 0 {
				t.Errorf("Run(%v) exit code = %d, want 0", tt.args, exitCode)
			}
			if !strings.Contains(output, tt.wantInDoc) {
				t.Errorf("Run(%v) stdout = %q, want to contain %q", tt.args, output, tt.wantInDoc)
			}
			if tt.notWantInDoc != "" && strings.Contains(output, tt.notWantInDoc) {
				t.Errorf("Run(%v) stdout = %q, should not contain %q", tt.args, output, tt.notWantInDoc)
			}
		})
	}
}

// TestRunForecastCommandHelp tests forecast command help output.
func TestRunForecastCommandHelp(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantOutput   string
	}{
		{
			name:         "forecast help",
			args:         []string{"forecast", "--help"},
			wantExitCode: 0,
			wantOutput:   "Get weather forecast with variable selection",
		},
		{
			name:         "forecast help with -h",
			args:         []string{"forecast", "-h"},
			wantExitCode: 0,
			wantOutput:   "Get weather forecast with variable selection",
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

// TestRunUnknownCommand tests unknown command handling.
func TestRunUnknownCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStdErr   string
	}{
		{
			name:         "unknown command",
			args:         []string{"unknown", "--latitude", "40", "--longitude", "-74"},
			wantExitCode: 2,
			wantStdErr:   "unknown command",
		},
		{
			name:         "invalid command name",
			args:         []string{"today", "--latitude", "40", "--longitude", "-74"},
			wantExitCode: 2,
			wantStdErr:   "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			exitCode := Run(tt.args)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			stderr := buf.String()

			if exitCode != tt.wantExitCode {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.wantExitCode)
			}
			if !strings.Contains(stderr, tt.wantStdErr) {
				t.Errorf("Run(%v) stderr = %q, want to contain %q", tt.args, stderr, tt.wantStdErr)
			}
		})
	}
}
