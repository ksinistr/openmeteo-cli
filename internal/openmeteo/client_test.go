package openmeteo

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockClient is a fake HTTP client for testing.
type MockClient struct {
	responseBody string
	statusCode   int
	err          error
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	resp := &http.Response{
		StatusCode: m.statusCode,
		Body:       &mockReadCloser{s: m.responseBody},
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

type mockReadCloser struct {
	s   string
	pos int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.s) {
		return 0, io.EOF
	}
	n = copy(p, m.s[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestClient_FetchForecast(t *testing.T) {
	mockResponse := `{
		"latitude": 40.0,
		"longitude": -74.0,
		"elevation": 10.0,
		"generationtime_ms": 10.5,
		"utc_offset_seconds": -18000,
		"timezone": "America/New_York",
		"timezone_abbreviation": "EST",
		"current": {
			"time": "2026-03-21T12:00",
			"temperature_2m": 15.5,
			"apparent_temperature": 14.0,
			"relative_humidity_2m": 65,
			"precipitation": 0.0,
			"precipitation_probability": 0,
			"wind_speed_10m": 5.5,
			"wind_gusts_10m": 8.0,
			"wind_direction_10m": 180,
			"uv_index": 3.0,
			"weather_code": 0
		},
		"hourly": {
			"time": ["2026-03-21T12:00", "2026-03-21T13:00"],
			"temperature_2m": [15.5, 16.0],
			"apparent_temperature": [14.0, 14.5],
			"relative_humidity_2m": [65, 64],
			"precipitation": [0.0, 0.0],
			"precipitation_probability": [0, 0],
			"wind_speed_10m": [5.5, 5.0],
			"wind_gusts_10m": [8.0, 7.5],
			"wind_direction_10m": [180, 185],
			"uv_index": [3.0, 3.5],
			"weather_code": [0, 0]
		},
		"daily": {
			"time": ["2026-03-21", "2026-03-22"],
			"weather_code": [0, 1],
			"temperature_2m_min": [10.0, 11.0],
			"temperature_2m_max": [20.0, 21.0],
			"precipitation_sum": [0.0, 0.5],
			"precipitation_probability_max": [0, 10],
			"wind_speed_10m_max": [5.5, 6.0],
			"wind_gusts_10m_max": [8.0, 8.5],
			"uv_index_max": [3.0, 4.0],
			"sunrise": ["2026-03-21T07:00", "2026-03-22T07:01"],
			"sunset": ["2026-03-21T19:00", "2026-03-22T19:01"]
		}
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	resp, err := client.FetchForecast(40.0, -74.0, "metric", "auto", 1)
	if err != nil {
		t.Fatalf("FetchForecast() returned error: %v", err)
	}

	// Verify response structure
	if resp.Latitude != 40.0 {
		t.Errorf("Latitude = %f, want 40.0", resp.Latitude)
	}
	if resp.Longitude != -74.0 {
		t.Errorf("Longitude = %f, want -74.0", resp.Longitude)
	}
	if resp.Timezone != "America/New_York" {
		t.Errorf("Timezone = %q, want America/New_York", resp.Timezone)
	}
	if resp.Current.Temperature2M != 15.5 {
		t.Errorf("Current.Temperature2M = %f, want 15.5", resp.Current.Temperature2M)
	}
	if len(resp.Hourly.Time) != 2 {
		t.Errorf("Hourly.Time length = %d, want 2", len(resp.Hourly.Time))
	}
	if len(resp.Daily.Time) != 2 {
		t.Errorf("Daily.Time length = %d, want 2", len(resp.Daily.Time))
	}
}

func TestClient_FetchForecast_HTTPError(t *testing.T) {
	mockClient := &MockClient{
		err:        errors.New("connection refused"),
		statusCode: http.StatusBadGateway,
	}

	client := NewClient(mockClient)

	_, err := client.FetchForecast(40.0, -74.0, "metric", "auto", 1)
	if err == nil {
		t.Error("FetchForecast() expected error, got nil")
	}
}

func TestClient_FetchForecast_BadRequest(t *testing.T) {
	mockClient := &MockClient{
		responseBody: `{"error": "Bad Request"}`,
		statusCode:   http.StatusBadRequest,
	}

	client := NewClient(mockClient)

	_, err := client.FetchForecast(40.0, -74.0, "metric", "auto", 1)
	if err == nil {
		t.Error("FetchForecast() expected error for bad status, got nil")
	}
}

func TestClient_FetchForecast_InvalidJSON(t *testing.T) {
	mockClient := &MockClient{
		responseBody: `not valid json`,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	_, err := client.FetchForecast(40.0, -74.0, "metric", "auto", 1)
	if err == nil {
		t.Error("FetchForecast() expected error for invalid JSON, got nil")
	}
}

func TestClient_buildRequest(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req, err := client.buildRequest(40.7128, -74.0060, "metric", "auto", 7)
	if err != nil {
		t.Fatalf("buildRequest() returned error: %v", err)
	}

	// Verify the URL starts with the expected base
	expectedBase := "https://api.open-meteo.com/v1/forecast?latitude=40.7128&longitude=-74.0060"
	if !strings.HasPrefix(req.URL.String(), expectedBase) {
		t.Errorf("URL = %q, prefix should be %q", req.URL.String(), expectedBase)
	}

	// Check user agent
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "openmeteo-cli/1.0.0" {
		t.Errorf("User-Agent = %q, want openmeteo-cli/1.0.0", userAgent)
	}
}

func TestClient_buildRequest_Imperial(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req, err := client.buildRequest(40.7128, -74.0060, "imperial", "auto", 7)
	if err != nil {
		t.Fatalf("buildRequest() returned error: %v", err)
	}

	if req.URL.String() == "" {
		t.Error("buildRequest() returned empty URL")
	}
}

func TestClient_buildRequest_CustomTimezone(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req, err := client.buildRequest(51.5074, -0.1278, "metric", "Europe/London", 7)
	if err != nil {
		t.Fatalf("buildRequest() returned error: %v", err)
	}

	// Check that timezone is in the URL (check that it was encoded in the URL)
	url := req.URL.String()
	if !strings.Contains(url, "timezone=") {
		t.Errorf("URL should contain timezone parameter, got: %q", url)
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	client := NewClient(nil)

	// Verify client is created with default HTTP client
	if client.httpClient == nil {
		t.Error("NewClient(nil) should set default HTTP client")
	}
}

func TestCurrent_StructFields(t *testing.T) {
	current := Current{
		Time:                     "2026-03-21T12:00",
		Temperature2M:            15.5,
		ApparentTemperature:      14.0,
		RelativeHumidity2M:       65,
		Precipitation:            0.0,
		PrecipitationProbability: 0,
		WindSpeed10M:             5.5,
		WindGusts10M:             8.0,
		WindDirection10M:         180,
		UVIndex:                  3.0,
		WeatherCode:              0,
	}

	if current.Time == "" {
		t.Error("Current.Time should not be empty")
	}
	if current.Temperature2M == 0 {
		t.Error("Current.Temperature2M should be set")
	}
}

func TestHourly_StructFields(t *testing.T) {
	hourly := Hourly{
		Time:                     []string{"2026-03-21T12:00"},
		Temperature2M:            []float64{15.5},
		ApparentTemperature:      []float64{14.0},
		RelativeHumidity2M:       []int{65},
		Precipitation:            []float64{0.0},
		PrecipitationProbability: []int{0},
		WindSpeed10M:             []float64{5.5},
		WindGusts10M:             []float64{8.0},
		WindDirection10M:         []int{180},
		UVIndex:                  []float64{3.0},
		WeatherCode:              []int{0},
	}

	if len(hourly.Time) == 0 {
		t.Error("Hourly.Time should not be empty")
	}
}

func TestDaily_StructFields(t *testing.T) {
	daily := Daily{
		Time:                        []string{"2026-03-21"},
		WeatherCode:                 []int{0},
		Temperature2MMin:            []float64{10.0},
		Temperature2MMax:            []float64{20.0},
		PrecipitationSum:            []float64{0.0},
		PrecipitationProbabilityMax: []int{0},
		WindSpeed10MMax:             []float64{5.5},
		WindGusts10MMax:             []float64{8.0},
		UVIndexMax:                  []float64{3.0},
		Sunrise:                     []string{"2026-03-21T07:00"},
		Sunset:                      []string{"2026-03-21T19:00"},
	}

	if len(daily.Time) == 0 {
		t.Error("Daily.Time should not be empty")
	}
}
