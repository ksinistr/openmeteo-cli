package openmeteo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Version is the current version of the openmeteo-cli tool.
const Version = "1.0.0"

// ErrUpstreamAPI is returned when the Open-Meteo API returns an error.
var ErrUpstreamAPI = errors.New("upstream API error")

// ErrLocationNotFound is returned when geocoding finds no results.
var ErrLocationNotFound = errors.New("location not found")

// ErrLocationAmbiguous is returned when geocoding finds multiple plausible results.
var ErrLocationAmbiguous = errors.New("location ambiguous")

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

// GeocodingResult represents a single location result from the geocoding API.
type GeocodingResult struct {
	Name        string   `json:"name"`
	Country     string   `json:"country"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Admin1      string   `json:"admin1,omitempty"`       // State/province
	Admin2      string   `json:"admin2,omitempty"`       // County/district
	Admin3      string   `json:"admin3,omitempty"`       // Municipality
	Admin4      string   `json:"admin4,omitempty"`       // Division
	Population  int      `json:"population,omitempty"`   // For result ranking
	FeatureCode string   `json:"feature_code,omitempty"` // Place type
	Timezone    string   `json:"timezone,omitempty"`     // Local timezone
	Postcodes   []string `json:"postcodes,omitempty"`    // Associated postcodes
}

// GeocodingResponse represents the geocoding API response.
type GeocodingResponse struct {
	Results []GeocodingResult `json:"results"`
}

// ResolvedLocation contains the coordinates and metadata for a resolved location.
type ResolvedLocation struct {
	Latitude  float64
	Longitude float64
	Name      string
	Country   string
	Admin1    string // State/province, if available
}

// Client wraps HTTP client for Open-Meteo API calls.
type Client struct {
	httpClient HTTPClient
	baseURL    string
	geocodeURL string
}

// NewClient creates a new Open-Meteo API client.
func NewClient(httpClient HTTPClient) *Client {
	if httpClient == nil {
		httpClient = NewRealHTTPClient()
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://api.open-meteo.com/v1/forecast",
		geocodeURL: GeocodeURL,
	}
}

// SetBaseURL sets a custom base URL for the client. Used for testing.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetGeocodeURL sets a custom geocode URL for the client. Used for testing.
func (c *Client) SetGeocodeURL(url string) {
	c.geocodeURL = url
}

// ForecastRequest carries all parameters for an Open-Meteo forecast API request.
type ForecastRequest struct {
	Latitude  float64
	Longitude float64

	// Sections to request - at least one must be non-empty
	CurrentVars []string // Variables for current weather
	HourlyVars  []string // Variables for hourly forecast
	DailyVars   []string // Variables for daily forecast

	// Range limits
	HourlyForecastHours int // Number of hours for hourly (1..48)
	DailyForecastDays   int // Number of days for daily (1..14)

	// Options
	Units string // "metric" or "imperial"
}

// FetchForecast fetches weather forecast using the provided request.
func (c *Client) FetchForecast(req ForecastRequest) (*APIResponse, error) {
	if err := validateForecastRequest(&req); err != nil {
		return nil, err
	}

	httpReq, err := c.buildForecastRequest(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	resp, err := c.httpClient.Do(httpReq)
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

// validateForecastRequest validates a forecast request.
func validateForecastRequest(req *ForecastRequest) error {
	// At least one section must be requested
	hasCurrent := len(req.CurrentVars) > 0
	hasHourly := len(req.HourlyVars) > 0
	hasDaily := len(req.DailyVars) > 0

	if !hasCurrent && !hasHourly && !hasDaily {
		return fmt.Errorf("at least one section (current, hourly, daily) must be requested")
	}

	// Validate coordinate ranges
	if req.Latitude < -90 || req.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", req.Latitude)
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", req.Longitude)
	}

	// Validate range limits based on requested sections
	if hasHourly {
		if req.HourlyForecastHours < 1 || req.HourlyForecastHours > 48 {
			return fmt.Errorf("hourly forecast hours must be between 1 and 48, got %d", req.HourlyForecastHours)
		}
	}
	if hasDaily {
		if req.DailyForecastDays < 1 || req.DailyForecastDays > 14 {
			return fmt.Errorf("daily forecast days must be between 1 and 14, got %d", req.DailyForecastDays)
		}
	}

	// Validate units
	if req.Units != "metric" && req.Units != "imperial" {
		return fmt.Errorf("units must be 'metric' or 'imperial', got %q", req.Units)
	}

	return nil
}

// buildForecastRequest constructs an HTTP request for the Open-Meteo forecast API.
// The request is pure and independent from transport for easy testing.
func (c *Client) buildForecastRequest(req *ForecastRequest) (*http.Request, error) {
	// Build base URL with coordinates
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	query := u.Query()

	query.Set("latitude", fmt.Sprintf("%.4f", req.Latitude))
	query.Set("longitude", fmt.Sprintf("%.4f", req.Longitude))

	// Add requested sections with their variables
	if len(req.CurrentVars) > 0 {
		query.Set("current", strings.Join(req.CurrentVars, ","))
	}
	if len(req.HourlyVars) > 0 {
		query.Set("hourly", strings.Join(req.HourlyVars, ","))
		query.Set("forecast_hours", fmt.Sprintf("%d", req.HourlyForecastHours))
	}
	if len(req.DailyVars) > 0 {
		query.Set("daily", strings.Join(req.DailyVars, ","))
		query.Set("forecast_days", fmt.Sprintf("%d", req.DailyForecastDays))
		// Use timezone=auto when daily is requested for proper date boundaries
		query.Set("timezone", "auto")
	}

	// Map units to Open-Meteo parameters
	// Default is metric (celsius, km/h, mm) - no parameters needed
	if req.Units == "imperial" {
		query.Set("temperature_unit", "fahrenheit")
		query.Set("wind_speed_unit", "mph")
		query.Set("precipitation_unit", "inch")
	}

	u.RawQuery = query.Encode()

	httpReq, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("User-Agent", "openmeteo-cli/"+Version)
	return httpReq, nil
}

// BuildForecastURL is a pure function that builds a forecast URL from a request.
// This is useful for testing request construction without HTTP transport.
func BuildForecastURL(baseURL string, req *ForecastRequest) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}
	query := u.Query()

	query.Set("latitude", fmt.Sprintf("%.4f", req.Latitude))
	query.Set("longitude", fmt.Sprintf("%.4f", req.Longitude))

	if len(req.CurrentVars) > 0 {
		query.Set("current", strings.Join(req.CurrentVars, ","))
	}
	if len(req.HourlyVars) > 0 {
		query.Set("hourly", strings.Join(req.HourlyVars, ","))
		query.Set("forecast_hours", fmt.Sprintf("%d", req.HourlyForecastHours))
	}
	if len(req.DailyVars) > 0 {
		query.Set("daily", strings.Join(req.DailyVars, ","))
		query.Set("forecast_days", fmt.Sprintf("%d", req.DailyForecastDays))
		query.Set("timezone", "auto")
	}

	if req.Units == "imperial" {
		query.Set("temperature_unit", "fahrenheit")
		query.Set("wind_speed_unit", "mph")
		query.Set("precipitation_unit", "inch")
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// GeocodeURL is the base URL for the Open-Meteo geocoding API.
const GeocodeURL = "https://geocoding-api.open-meteo.com/v1/search"

// FetchLocation resolves a place name to coordinates and metadata.
// Returns ErrLocationNotFound if no results are found.
// Returns ErrLocationAmbiguous if multiple plausible results are found (count > 1).
func (c *Client) FetchLocation(name string, count int) (*ResolvedLocation, error) {
	if name == "" {
		return nil, fmt.Errorf("location name cannot be empty")
	}
	if count < 1 {
		count = 1
	}
	if count > 20 {
		count = 20 // Open-Meteo API limit
	}

	req, err := c.buildGeocodeRequest(name, count)
	if err != nil {
		return nil, fmt.Errorf("failed to build geocode request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch location data: %v", ErrUpstreamAPI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: geocoding API returned status %d", ErrUpstreamAPI, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read geocoding response: %v", ErrUpstreamAPI, err)
	}

	var geoResp GeocodingResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse geocoding response: %v", ErrUpstreamAPI, err)
	}

	if len(geoResp.Results) == 0 {
		return nil, ErrLocationNotFound
	}

	if len(geoResp.Results) > 1 {
		// Return an error with details about the ambiguous results
		return nil, fmt.Errorf("%w: found %d results for %q, please be more specific", ErrLocationAmbiguous, len(geoResp.Results), name)
	}

	result := geoResp.Results[0]
	return &ResolvedLocation{
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		Name:      result.Name,
		Country:   result.Country,
		Admin1:    result.Admin1,
	}, nil
}

// FetchLocationRaw returns all geocoding results without enforcing uniqueness.
// This is useful for displaying options to the user when the location is ambiguous.
func (c *Client) FetchLocationRaw(name string, count int) ([]GeocodingResult, error) {
	if name == "" {
		return nil, fmt.Errorf("location name cannot be empty")
	}
	if count < 1 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	req, err := c.buildGeocodeRequest(name, count)
	if err != nil {
		return nil, fmt.Errorf("failed to build geocode request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch location data: %v", ErrUpstreamAPI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: geocoding API returned status %d", ErrUpstreamAPI, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read geocoding response: %v", ErrUpstreamAPI, err)
	}

	var geoResp GeocodingResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse geocoding response: %v", ErrUpstreamAPI, err)
	}

	if len(geoResp.Results) == 0 {
		return nil, ErrLocationNotFound
	}

	return geoResp.Results, nil
}

// buildGeocodeRequest constructs the HTTP request for the geocoding API.
func (c *Client) buildGeocodeRequest(name string, count int) (*http.Request, error) {
	baseURL := fmt.Sprintf("%s?name=%s&count=%d&language=en&format=json",
		c.geocodeURL,
		url.QueryEscape(name),
		count,
	)

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "openmeteo-cli/"+Version)
	return req, nil
}
