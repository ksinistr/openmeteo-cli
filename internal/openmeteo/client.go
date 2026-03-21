package openmeteo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Version is the current version of the openmeteo-cli tool.
const Version = "1.0.0"

// ErrUpstreamAPI is returned when the Open-Meteo API returns an error.
var ErrUpstreamAPI = errors.New("upstream API error")

// HTTPClient is an interface for HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RealHTTPClient is a simple HTTP client implementation with timeout.
type RealHTTPClient struct {
	client *http.Client
}

// NewRealHTTPClient creates a RealHTTPClient with default timeout.
func NewRealHTTPClient() *RealHTTPClient {
	return &RealHTTPClient{
		client: &http.Client{Timeout: 35 * time.Second},
	}
}

// Do executes an HTTP request.
func (c *RealHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// APIResponse represents the Open-Meteo API response.
type APIResponse struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Elevation        float64 `json:"elevation"`
	GenerationTimeMS float64 `json:"generationtime_ms"`
	UTCOffsetSeconds int     `json:"utc_offset_seconds"`
	Timezone         string  `json:"timezone"`
	TimezoneAbbrev   string  `json:"timezone_abbreviation"`
	Current          Current `json:"current"`
	Hourly           Hourly  `json:"hourly"`
	Daily            Daily   `json:"daily"`
}

// Current represents the current weather conditions.
type Current struct {
	Time                     string  `json:"time"`
	Temperature2M            float64 `json:"temperature_2m"`
	ApparentTemperature      float64 `json:"apparent_temperature"`
	RelativeHumidity2M       int     `json:"relative_humidity_2m"`
	Precipitation            float64 `json:"precipitation"`
	PrecipitationProbability int     `json:"precipitation_probability"`
	WindSpeed10M             float64 `json:"wind_speed_10m"`
	WindGusts10M             float64 `json:"wind_gusts_10m"`
	WindDirection10M         int     `json:"wind_direction_10m"`
	UVIndex                  float64 `json:"uv_index"`
	WeatherCode              int     `json:"weather_code"`
}

// Hourly represents hourly weather data.
type Hourly struct {
	Time                     []string  `json:"time"`
	Temperature2M            []float64 `json:"temperature_2m"`
	ApparentTemperature      []float64 `json:"apparent_temperature"`
	RelativeHumidity2M       []int     `json:"relative_humidity_2m"`
	Precipitation            []float64 `json:"precipitation"`
	PrecipitationProbability []int     `json:"precipitation_probability"`
	WindSpeed10M             []float64 `json:"wind_speed_10m"`
	WindGusts10M             []float64 `json:"wind_gusts_10m"`
	WindDirection10M         []int     `json:"wind_direction_10m"`
	UVIndex                  []float64 `json:"uv_index"`
	WeatherCode              []int     `json:"weather_code"`
}

// Daily represents daily weather data.
type Daily struct {
	Time                        []string  `json:"time"`
	WeatherCode                 []int     `json:"weather_code"`
	Temperature2MMin            []float64 `json:"temperature_2m_min"`
	Temperature2MMax            []float64 `json:"temperature_2m_max"`
	PrecipitationSum            []float64 `json:"precipitation_sum"`
	PrecipitationProbabilityMax []int     `json:"precipitation_probability_max"`
	WindSpeed10MMax             []float64 `json:"wind_speed_10m_max"`
	WindGusts10MMax             []float64 `json:"wind_gusts_10m_max"`
	UVIndexMax                  []float64 `json:"uv_index_max"`
	Sunrise                     []string  `json:"sunrise"`
	Sunset                      []string  `json:"sunset"`
}

// Client wraps HTTP client for Open-Meteo API calls.
type Client struct {
	httpClient HTTPClient
	baseURL    string
}

// NewClient creates a new Open-Meteo API client.
func NewClient(httpClient HTTPClient) *Client {
	if httpClient == nil {
		httpClient = NewRealHTTPClient()
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://api.open-meteo.com/v1/forecast",
	}
}

// FetchForecast fetches weather forecast for the given coordinates.
func (c *Client) FetchForecast(lat, lon float64, units, timezone string, forecastDays int) (*APIResponse, error) {
	if forecastDays < 1 || forecastDays > 16 {
		return nil, fmt.Errorf("forecastDays must be between 1 and 16, got %d", forecastDays)
	}

	// Validate timezone - "auto" is a special value for the Open-Meteo API
	if timezone != "auto" {
		if _, err := time.LoadLocation(timezone); err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", timezone, err)
		}
	}

	req, err := c.buildRequest(lat, lon, units, timezone, forecastDays)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch data: %v", ErrUpstreamAPI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: API returned status %d", ErrUpstreamAPI, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrUpstreamAPI, err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrUpstreamAPI, err)
	}

	return &apiResp, nil
}

// buildRequest constructs the HTTP request for the Open-Meteo API.
func (c *Client) buildRequest(lat, lon float64, units, timezone string, forecastDays int) (*http.Request, error) {
	currentParams := "temperature_2m,apparent_temperature,relative_humidity_2m,precipitation,precipitation_probability,wind_speed_10m,wind_gusts_10m,wind_direction_10m,uv_index,weather_code"
	hourlyParams := "temperature_2m,apparent_temperature,relative_humidity_2m,precipitation,precipitation_probability,wind_speed_10m,wind_gusts_10m,wind_direction_10m,uv_index,weather_code"
	dailyParams := "weather_code,temperature_2m_min,temperature_2m_max,precipitation_sum,precipitation_probability_max,wind_speed_10m_max,wind_gusts_10m_max,uv_index_max,sunrise,sunset"

	baseURL := fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f&current=%s&hourly=%s&daily=%s&timezone=%s&forecast_days=%d",
		c.baseURL, lat, lon,
		currentParams, hourlyParams, dailyParams,
		url.QueryEscape(timezone),
		forecastDays,
	)

	// Add units parameters if specified
	// Open-Meteo uses separate parameters for different unit types
	if units == "imperial" {
		baseURL += "&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch"
	}
	// Default is metric (celsius, km/h, mm) - no parameters needed

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "openmeteo-cli/"+Version)
	return req, nil
}
