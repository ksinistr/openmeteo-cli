package openmeteo

import (
	"errors"
	"io"
	"net/http"
	"net/url"
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
			"precipitation_probability": [0, 0]
		},
		"daily": {
			"time": ["2026-03-21", "2026-03-22"],
			"weather_code": [0, 1],
			"temperature_2m_min": [10.0, 11.0],
			"temperature_2m_max": [20.0, 21.0]
		}
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	req := ForecastRequest{
		Latitude:            40.0,
		Longitude:           -74.0,
		CurrentVars:         []string{"temperature_2m", "weather_code"},
		HourlyVars:          []string{"temperature_2m", "precipitation_probability"},
		DailyVars:           []string{"weather_code", "temperature_2m_min", "temperature_2m_max"},
		HourlyForecastHours: 2,
		DailyForecastDays:   2,
		Units:               "metric",
	}

	resp, err := client.FetchForecast(req)
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

	req := ForecastRequest{
		Latitude:            40.0,
		Longitude:           -74.0,
		CurrentVars:         []string{"temperature_2m"},
		HourlyForecastHours: 1,
		DailyForecastDays:   1,
		Units:               "metric",
	}

	_, err := client.FetchForecast(req)
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

	req := ForecastRequest{
		Latitude:            40.0,
		Longitude:           -74.0,
		CurrentVars:         []string{"temperature_2m"},
		HourlyForecastHours: 1,
		DailyForecastDays:   1,
		Units:               "metric",
	}

	_, err := client.FetchForecast(req)
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

	req := ForecastRequest{
		Latitude:            40.0,
		Longitude:           -74.0,
		CurrentVars:         []string{"temperature_2m"},
		HourlyForecastHours: 1,
		DailyForecastDays:   1,
		Units:               "metric",
	}

	_, err := client.FetchForecast(req)
	if err == nil {
		t.Error("FetchForecast() expected error for invalid JSON, got nil")
	}
}

func TestClient_validateForecastRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ForecastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid current-only request",
			req: ForecastRequest{
				Latitude:    40.0,
				Longitude:   -74.0,
				CurrentVars: []string{"temperature_2m"},
				Units:       "metric",
			},
			wantErr: false,
		},
		{
			name: "valid hourly-only request",
			req: ForecastRequest{
				Latitude:            40.0,
				Longitude:           -74.0,
				HourlyVars:          []string{"temperature_2m"},
				HourlyForecastHours: 24,
				Units:               "metric",
			},
			wantErr: false,
		},
		{
			name: "valid daily-only request",
			req: ForecastRequest{
				Latitude:          40.0,
				Longitude:         -74.0,
				DailyVars:         []string{"temperature_2m_max"},
				DailyForecastDays: 7,
				Units:             "metric",
			},
			wantErr: false,
		},
		{
			name: "valid combined request",
			req: ForecastRequest{
				Latitude:            40.0,
				Longitude:           -74.0,
				CurrentVars:         []string{"temperature_2m"},
				HourlyVars:          []string{"temperature_2m"},
				DailyVars:           []string{"temperature_2m_max"},
				HourlyForecastHours: 24,
				DailyForecastDays:   7,
				Units:               "metric",
			},
			wantErr: false,
		},
		{
			name: "error: no sections requested",
			req: ForecastRequest{
				Latitude:  40.0,
				Longitude: -74.0,
				Units:     "metric",
			},
			wantErr: true,
			errMsg:  "at least one section",
		},
		{
			name: "error: latitude out of range",
			req: ForecastRequest{
				Latitude:    95.0,
				Longitude:   -74.0,
				CurrentVars: []string{"temperature_2m"},
				Units:       "metric",
			},
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: "error: longitude out of range",
			req: ForecastRequest{
				Latitude:    40.0,
				Longitude:   -185.0,
				CurrentVars: []string{"temperature_2m"},
				Units:       "metric",
			},
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name: "error: hourly hours too low",
			req: ForecastRequest{
				Latitude:            40.0,
				Longitude:           -74.0,
				HourlyVars:          []string{"temperature_2m"},
				HourlyForecastHours: 0,
				Units:               "metric",
			},
			wantErr: true,
			errMsg:  "hourly forecast hours must be between 1 and 48",
		},
		{
			name: "error: hourly hours too high",
			req: ForecastRequest{
				Latitude:            40.0,
				Longitude:           -74.0,
				HourlyVars:          []string{"temperature_2m"},
				HourlyForecastHours: 49,
				Units:               "metric",
			},
			wantErr: true,
			errMsg:  "hourly forecast hours must be between 1 and 48",
		},
		{
			name: "error: daily days too low",
			req: ForecastRequest{
				Latitude:          40.0,
				Longitude:         -74.0,
				DailyVars:         []string{"temperature_2m_max"},
				DailyForecastDays: 0,
				Units:             "metric",
			},
			wantErr: true,
			errMsg:  "daily forecast days must be between 1 and 14",
		},
		{
			name: "error: daily days too high",
			req: ForecastRequest{
				Latitude:          40.0,
				Longitude:         -74.0,
				DailyVars:         []string{"temperature_2m_max"},
				DailyForecastDays: 15,
				Units:             "metric",
			},
			wantErr: true,
			errMsg:  "daily forecast days must be between 1 and 14",
		},
		{
			name: "error: invalid units",
			req: ForecastRequest{
				Latitude:    40.0,
				Longitude:   -74.0,
				CurrentVars: []string{"temperature_2m"},
				Units:       "invalid",
			},
			wantErr: true,
			errMsg:  "units must be 'metric' or 'imperial'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateForecastRequest(&tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateForecastRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestClient_buildForecastRequest_CurrentOnly(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req := &ForecastRequest{
		Latitude:    52.52,
		Longitude:   13.405,
		CurrentVars: []string{"temperature_2m", "weather_code"},
		Units:       "metric",
	}

	httpReq, err := client.buildForecastRequest(req)
	if err != nil {
		t.Fatalf("buildForecastRequest() returned error: %v", err)
	}

	url := httpReq.URL.String()

	// Check that current is included (comma is URL encoded as %2C)
	if !strings.Contains(url, "current=") {
		t.Errorf("URL should contain current parameter, got: %q", url)
	}
	if !strings.Contains(url, "temperature_2m") {
		t.Errorf("URL should contain temperature_2m variable, got: %q", url)
	}
	if !strings.Contains(url, "weather_code") {
		t.Errorf("URL should contain weather_code variable, got: %q", url)
	}

	// Check that hourly and daily are NOT included
	if strings.Contains(url, "hourly=") {
		t.Errorf("URL should not contain hourly parameter, got: %q", url)
	}
	if strings.Contains(url, "daily=") {
		t.Errorf("URL should not contain daily parameter, got: %q", url)
	}

	// Check that timezone is NOT included for current-only (no daily data)
	if strings.Contains(url, "timezone=") {
		t.Errorf("URL should not contain timezone parameter for current-only, got: %q", url)
	}
}

func TestClient_buildForecastRequest_HourlyOnly(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req := &ForecastRequest{
		Latitude:            52.52,
		Longitude:           13.405,
		HourlyVars:          []string{"temperature_2m", "precipitation_probability"},
		HourlyForecastHours: 24,
		Units:               "metric",
	}

	httpReq, err := client.buildForecastRequest(req)
	if err != nil {
		t.Fatalf("buildForecastRequest() returned error: %v", err)
	}

	url := httpReq.URL.String()

	// Check that hourly is included (comma is URL encoded as %2C)
	if !strings.Contains(url, "hourly=") {
		t.Errorf("URL should contain hourly parameter, got: %q", url)
	}
	if !strings.Contains(url, "temperature_2m") {
		t.Errorf("URL should contain temperature_2m variable, got: %q", url)
	}
	if !strings.Contains(url, "precipitation_probability") {
		t.Errorf("URL should contain precipitation_probability variable, got: %q", url)
	}

	// Check that forecast_hours is set correctly
	if !strings.Contains(url, "forecast_hours=24") {
		t.Errorf("URL should contain forecast_hours=24, got: %q", url)
	}

	// Check that current and daily are NOT included
	if strings.Contains(url, "current=") {
		t.Errorf("URL should not contain current parameter, got: %q", url)
	}
	if strings.Contains(url, "daily=") {
		t.Errorf("URL should not contain daily parameter, got: %q", url)
	}

	// Check that timezone is NOT included for hourly-only (no daily data)
	if strings.Contains(url, "timezone=") {
		t.Errorf("URL should not contain timezone parameter for hourly-only, got: %q", url)
	}
}

func TestClient_buildForecastRequest_DailyOnly(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req := &ForecastRequest{
		Latitude:          52.52,
		Longitude:         13.405,
		DailyVars:         []string{"weather_code", "temperature_2m_max", "temperature_2m_min"},
		DailyForecastDays: 7,
		Units:             "metric",
	}

	httpReq, err := client.buildForecastRequest(req)
	if err != nil {
		t.Fatalf("buildForecastRequest() returned error: %v", err)
	}

	url := httpReq.URL.String()

	// Check that daily is included (comma is URL encoded as %2C)
	if !strings.Contains(url, "daily=") {
		t.Errorf("URL should contain daily parameter, got: %q", url)
	}
	if !strings.Contains(url, "weather_code") {
		t.Errorf("URL should contain weather_code variable, got: %q", url)
	}
	if !strings.Contains(url, "temperature_2m_max") {
		t.Errorf("URL should contain temperature_2m_max variable, got: %q", url)
	}
	if !strings.Contains(url, "temperature_2m_min") {
		t.Errorf("URL should contain temperature_2m_min variable, got: %q", url)
	}

	// Check that forecast_days is set correctly
	if !strings.Contains(url, "forecast_days=7") {
		t.Errorf("URL should contain forecast_days=7, got: %q", url)
	}

	// Check that timezone=auto is included when daily is requested
	if !strings.Contains(url, "timezone=auto") {
		t.Errorf("URL should contain timezone=auto for daily request, got: %q", url)
	}

	// Check that current and hourly are NOT included
	if strings.Contains(url, "current=") {
		t.Errorf("URL should not contain current parameter, got: %q", url)
	}
	if strings.Contains(url, "hourly=") {
		t.Errorf("URL should not contain hourly parameter, got: %q", url)
	}
}

func TestClient_buildForecastRequest_Combined(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req := &ForecastRequest{
		Latitude:            52.52,
		Longitude:           13.405,
		CurrentVars:         []string{"temperature_2m", "weather_code"},
		HourlyVars:          []string{"temperature_2m", "precipitation_probability"},
		DailyVars:           []string{"weather_code", "temperature_2m_max"},
		HourlyForecastHours: 24,
		DailyForecastDays:   7,
		Units:               "metric",
	}

	httpReq, err := client.buildForecastRequest(req)
	if err != nil {
		t.Fatalf("buildForecastRequest() returned error: %v", err)
	}

	url := httpReq.URL.String()

	// Check that all sections are included
	if !strings.Contains(url, "current=") {
		t.Errorf("URL should contain current parameter, got: %q", url)
	}
	if !strings.Contains(url, "hourly=") {
		t.Errorf("URL should contain hourly parameter, got: %q", url)
	}
	if !strings.Contains(url, "daily=") {
		t.Errorf("URL should contain daily parameter, got: %q", url)
	}
	if !strings.Contains(url, "forecast_hours=24") {
		t.Errorf("URL should contain forecast_hours=24, got: %q", url)
	}
	if !strings.Contains(url, "forecast_days=7") {
		t.Errorf("URL should contain forecast_days=7, got: %q", url)
	}
	if !strings.Contains(url, "timezone=auto") {
		t.Errorf("URL should contain timezone=auto for daily request, got: %q", url)
	}
}

func TestClient_buildForecastRequest_Imperial(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req := &ForecastRequest{
		Latitude:    40.7128,
		Longitude:   -74.0060,
		CurrentVars: []string{"temperature_2m"},
		Units:       "imperial",
	}

	httpReq, err := client.buildForecastRequest(req)
	if err != nil {
		t.Fatalf("buildForecastRequest() returned error: %v", err)
	}

	url := httpReq.URL.String()

	// Check that imperial unit parameters are set
	if !strings.Contains(url, "temperature_unit=fahrenheit") {
		t.Errorf("URL should contain temperature_unit=fahrenheit, got: %q", url)
	}
	if !strings.Contains(url, "wind_speed_unit=mph") {
		t.Errorf("URL should contain wind_speed_unit=mph, got: %q", url)
	}
	if !strings.Contains(url, "precipitation_unit=inch") {
		t.Errorf("URL should contain precipitation_unit=inch, got: %q", url)
	}
}

func TestBuildForecastURL(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		req        ForecastRequest
		wantParams map[string]string // query params to check (key -> expected substring in value)
		wantErr    bool
	}{
		{
			name:    "current-only request",
			baseURL: "https://api.open-meteo.com/v1/forecast",
			req: ForecastRequest{
				Latitude:    52.52,
				Longitude:   13.405,
				CurrentVars: []string{"temperature_2m", "weather_code"},
				Units:       "metric",
			},
			wantParams: map[string]string{
				"current": "temperature_2m",
			},
			wantErr: false,
		},
		{
			name:    "hourly-only request",
			baseURL: "https://api.open-meteo.com/v1/forecast",
			req: ForecastRequest{
				Latitude:            40.7128,
				Longitude:           -74.0060,
				HourlyVars:          []string{"temperature_2m"},
				HourlyForecastHours: 48,
				Units:               "metric",
			},
			wantParams: map[string]string{
				"hourly":         "temperature_2m",
				"forecast_hours": "48",
			},
			wantErr: false,
		},
		{
			name:    "daily-only request",
			baseURL: "https://api.open-meteo.com/v1/forecast",
			req: ForecastRequest{
				Latitude:          51.5074,
				Longitude:         -0.1278,
				DailyVars:         []string{"temperature_2m_max", "temperature_2m_min"},
				DailyForecastDays: 14,
				Units:             "metric",
			},
			wantParams: map[string]string{
				"daily":         "temperature_2m_max",
				"forecast_days": "14",
				"timezone":      "auto",
			},
			wantErr: false,
		},
		{
			name:    "imperial units",
			baseURL: "https://api.open-meteo.com/v1/forecast",
			req: ForecastRequest{
				Latitude:    35.6762,
				Longitude:   139.6503,
				CurrentVars: []string{"temperature_2m"},
				Units:       "imperial",
			},
			wantParams: map[string]string{
				"current":            "temperature_2m",
				"temperature_unit":   "fahrenheit",
				"wind_speed_unit":    "mph",
				"precipitation_unit": "inch",
			},
			wantErr: false,
		},
		{
			name:    "combined request",
			baseURL: "https://api.open-meteo.com/v1/forecast",
			req: ForecastRequest{
				Latitude:            52.52,
				Longitude:           13.405,
				CurrentVars:         []string{"temperature_2m"},
				HourlyVars:          []string{"precipitation"},
				DailyVars:           []string{"weather_code"},
				HourlyForecastHours: 24,
				DailyForecastDays:   7,
				Units:               "metric",
			},
			wantParams: map[string]string{
				"current":        "temperature_2m",
				"hourly":         "precipitation",
				"forecast_hours": "24",
				"daily":          "weather_code",
				"forecast_days":  "7",
				"timezone":       "auto",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := BuildForecastURL(tt.baseURL, &tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildForecastURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Parse the URL and check query params
			u, parseErr := url.Parse(gotURL)
			if parseErr != nil {
				t.Fatalf("Failed to parse generated URL: %v", parseErr)
			}

			query := u.Query()
			for key, wantSubstr := range tt.wantParams {
				gotVal := query.Get(key)
				if gotVal == "" {
					t.Errorf("BuildForecastURL() missing query parameter %q, got URL: %q", key, gotURL)
				}
				if !strings.Contains(gotVal, wantSubstr) {
					t.Errorf("BuildForecastURL() parameter %q = %q, want containing %q", key, gotVal, wantSubstr)
				}
			}
		})
	}
}

func TestClient_buildForecastRequest_UserAgent(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}

	req := &ForecastRequest{
		Latitude:    40.7128,
		Longitude:   -74.0060,
		CurrentVars: []string{"temperature_2m"},
		Units:       "metric",
	}

	httpReq, err := client.buildForecastRequest(req)
	if err != nil {
		t.Fatalf("buildForecastRequest() returned error: %v", err)
	}

	userAgent := httpReq.Header.Get("User-Agent")
	if userAgent != "openmeteo-cli/1.0.0" {
		t.Errorf("User-Agent = %q, want openmeteo-cli/1.0.0", userAgent)
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

func TestClient_FetchLocation_Success(t *testing.T) {
	mockResponse := `{
		"results": [
			{
				"name": "Berlin",
				"country": "Germany",
				"latitude": 52.52,
				"longitude": 13.405,
				"admin1": "Berlin",
				"population": 3645000,
				"feature_code": "PPLC",
				"timezone": "Europe/Berlin"
			}
		]
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	location, err := client.FetchLocation("Berlin", 1)
	if err != nil {
		t.Fatalf("FetchLocation() returned error: %v", err)
	}

	if location.Name != "Berlin" {
		t.Errorf("Name = %q, want Berlin", location.Name)
	}
	if location.Country != "Germany" {
		t.Errorf("Country = %q, want Germany", location.Country)
	}
	if location.Latitude != 52.52 {
		t.Errorf("Latitude = %f, want 52.52", location.Latitude)
	}
	if location.Longitude != 13.405 {
		t.Errorf("Longitude = %f, want 13.405", location.Longitude)
	}
	if location.Admin1 != "Berlin" {
		t.Errorf("Admin1 = %q, want Berlin", location.Admin1)
	}
}

func TestClient_FetchLocation_NotFound(t *testing.T) {
	mockResponse := `{
		"results": []
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocation("NonexistentPlace12345", 1)
	if err == nil {
		t.Fatal("FetchLocation() expected error for empty results, got nil")
	}
	if !errors.Is(err, ErrLocationNotFound) {
		t.Errorf("FetchLocation() error = %v, want ErrLocationNotFound", err)
	}
}

func TestClient_FetchLocation_Ambiguous(t *testing.T) {
	mockResponse := `{
		"results": [
			{
				"name": "Springfield",
				"country": "United States",
				"latitude": 37.22,
				"longitude": -93.29,
				"admin1": "Missouri",
				"population": 17000
			},
			{
				"name": "Springfield",
				"country": "United States",
				"latitude": 42.11,
				"longitude": -72.54,
				"admin1": "Massachusetts",
				"population": 9000
			},
			{
				"name": "Springfield",
				"country": "United States",
				"latitude": 39.78,
				"longitude": -89.65,
				"admin1": "Illinois",
				"population": 116000
			}
		]
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocation("Springfield", 5)
	if err == nil {
		t.Fatal("FetchLocation() expected error for ambiguous results, got nil")
	}
	if !errors.Is(err, ErrLocationAmbiguous) {
		t.Errorf("FetchLocation() error = %v, want ErrLocationAmbiguous", err)
	}
}

func TestClient_FetchLocation_EmptyName(t *testing.T) {
	mockClient := &MockClient{
		statusCode: http.StatusOK,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocation("", 1)
	if err == nil {
		t.Fatal("FetchLocation() expected error for empty name, got nil")
	}
}

func TestClient_FetchLocation_HTTPError(t *testing.T) {
	mockClient := &MockClient{
		err:        errors.New("connection refused"),
		statusCode: http.StatusBadGateway,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocation("Berlin", 1)
	if err == nil {
		t.Fatal("FetchLocation() expected error for HTTP failure, got nil")
	}
	if !errors.Is(err, ErrUpstreamAPI) {
		t.Errorf("FetchLocation() error = %v, want ErrUpstreamAPI", err)
	}
}

func TestClient_FetchLocation_InvalidStatus(t *testing.T) {
	mockClient := &MockClient{
		responseBody: `{"error": "Bad Request"}`,
		statusCode:   http.StatusBadRequest,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocation("Berlin", 1)
	if err == nil {
		t.Fatal("FetchLocation() expected error for bad status, got nil")
	}
}

func TestClient_FetchLocation_InvalidJSON(t *testing.T) {
	mockClient := &MockClient{
		responseBody: `not valid json`,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocation("Berlin", 1)
	if err == nil {
		t.Fatal("FetchLocation() expected error for invalid JSON, got nil")
	}
}

func TestClient_FetchLocationRaw_Success(t *testing.T) {
	mockResponse := `{
		"results": [
			{
				"name": "Portland",
				"country": "United States",
				"latitude": 45.52,
				"longitude": -122.68,
				"admin1": "Oregon",
				"population": 650000
			},
			{
				"name": "Portland",
				"country": "United States",
				"latitude": 43.66,
				"longitude": -70.26,
				"admin1": "Maine",
				"population": 67000
			}
		]
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	results, err := client.FetchLocationRaw("Portland", 5)
	if err != nil {
		t.Fatalf("FetchLocationRaw() returned error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("FetchLocationRaw() returned %d results, want 2", len(results))
	}
	if results[0].Name != "Portland" {
		t.Errorf("Results[0].Name = %q, want Portland", results[0].Name)
	}
	if results[0].Admin1 != "Oregon" {
		t.Errorf("Results[0].Admin1 = %q, want Oregon", results[0].Admin1)
	}
}

func TestClient_FetchLocationRaw_NotFound(t *testing.T) {
	mockResponse := `{
		"results": []
	}`

	mockClient := &MockClient{
		responseBody: mockResponse,
		statusCode:   http.StatusOK,
	}

	client := NewClient(mockClient)

	_, err := client.FetchLocationRaw("NonexistentPlace12345", 1)
	if err == nil {
		t.Fatal("FetchLocationRaw() expected error for empty results, got nil")
	}
	if !errors.Is(err, ErrLocationNotFound) {
		t.Errorf("FetchLocationRaw() error = %v, want ErrLocationNotFound", err)
	}
}

func TestClient_buildGeocodeRequest(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		geocodeURL: GeocodeURL,
	}

	req, err := client.buildGeocodeRequest("Berlin", 1)
	if err != nil {
		t.Fatalf("buildGeocodeRequest() returned error: %v", err)
	}

	// Verify the URL starts with the expected base
	expectedBase := "https://geocoding-api.open-meteo.com/v1/search?name=Berlin&count=1"
	if !strings.HasPrefix(req.URL.String(), expectedBase) {
		t.Errorf("URL = %q, prefix should start with %q", req.URL.String(), expectedBase)
	}

	// Check user agent
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "openmeteo-cli/1.0.0" {
		t.Errorf("User-Agent = %q, want openmeteo-cli/1.0.0", userAgent)
	}
}

func TestClient_buildGeocodeRequest_Escaping(t *testing.T) {
	client := &Client{
		httpClient: &MockClient{},
		geocodeURL: GeocodeURL,
	}

	req, err := client.buildGeocodeRequest("San Francisco", 1)
	if err != nil {
		t.Fatalf("buildGeocodeRequest() returned error: %v", err)
	}

	url := req.URL.String()
	if !strings.Contains(url, "San+Francisco") && !strings.Contains(url, "San%20Francisco") {
		t.Errorf("URL should contain escaped space, got: %q", url)
	}
}

func TestResolvedLocation_StructFields(t *testing.T) {
	loc := ResolvedLocation{
		Latitude:  52.52,
		Longitude: 13.405,
		Name:      "Berlin",
		Country:   "Germany",
		Admin1:    "Berlin",
	}

	if loc.Name != "Berlin" {
		t.Error("ResolvedLocation.Name should be Berlin")
	}
	if loc.Country != "Germany" {
		t.Error("ResolvedLocation.Country should be Germany")
	}
	if loc.Latitude == 0 {
		t.Error("ResolvedLocation.Latitude should be set")
	}
	if loc.Longitude == 0 {
		t.Error("ResolvedLocation.Longitude should be set")
	}
}
