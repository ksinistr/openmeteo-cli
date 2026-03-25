package forecast

import (
	"fmt"
	"sort"
	"time"

	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/weathercode"
)

// ErrDateUnavailable is returned when the requested date is outside the forecast range.
var ErrDateUnavailable = fmt.Errorf("date not available in forecast")

// Service provides weather forecast services.
type Service struct {
	client        *openmeteo.Client
	weatherMapper *weathercode.Mapper
}

// NewService creates a new forecast service.
func NewService(client *openmeteo.Client, mapper *weathercode.Mapper) *Service {
	return &Service{
		client:        client,
		weatherMapper: mapper,
	}
}

// Forecast returns weather forecast based on the mode.
// location is optional - if provided via geocoding, it will be included in the response metadata.
func (s *Service) Forecast(lat, lon float64, units string, hourly bool, forecastDays int, location *openmeteo.ResolvedLocation) (interface{}, error) {
	if hourly {
		return s.fetchHourlyForecast(lat, lon, units, forecastDays, location)
	}
	return s.fetchDailyForecast(lat, lon, units, forecastDays, location)
}

// fetchHourlyForecast fetches hourly forecast for the specified number of days (max 2).
func (s *Service) fetchHourlyForecast(lat, lon float64, units string, forecastDays int, location *openmeteo.ResolvedLocation) (*HourlyOutput, error) {
	if forecastDays < 1 || forecastDays > 2 {
		return nil, fmt.Errorf("hourly forecast supports 1-2 days, got %d", forecastDays)
	}

	now := time.Now()

	// Legacy variable set for hourly compatibility
	req := openmeteo.ForecastRequest{
		Latitude:            lat,
		Longitude:           lon,
		CurrentVars:         []string{"temperature_2m", "apparent_temperature", "relative_humidity_2m", "precipitation", "precipitation_probability", "wind_speed_10m", "wind_gusts_10m", "wind_direction_10m", "uv_index", "weather_code"},
		HourlyVars:          []string{"temperature_2m", "apparent_temperature", "relative_humidity_2m", "precipitation", "precipitation_probability", "wind_speed_10m", "wind_gusts_10m", "wind_direction_10m", "uv_index", "weather_code"},
		HourlyForecastHours: forecastDays * 24, // Convert days to hours for hourly forecast
		DailyForecastDays:   forecastDays,
		Units:               units,
	}

	apiResp, err := s.client.FetchForecast(req)
	if err != nil {
		return nil, err
	}

	// Get timezone from API response
	loc, err := time.LoadLocation(apiResp.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid timezone %q from API: %v", openmeteo.ErrUpstreamAPI, apiResp.Timezone, err)
	}

	// Map hourly conditions for the requested days
	today := now.In(loc).Format("2006-01-02")
	daysMap, err := s.mapHourlyForDays(apiResp.Hourly, today, forecastDays, loc)
	if err != nil {
		return nil, err
	}

	meta := Meta{
		Units:     s.getUnits(units),
		Timezone:  apiResp.Timezone,
		Latitude:  apiResp.Latitude,
		Longitude: apiResp.Longitude,
	}

	// Include location metadata if provided via geocoding
	if location != nil {
		meta.Location = &Location{
			Name:    location.Name,
			Country: location.Country,
		}
	}

	return &HourlyOutput{
		Meta: meta,
		Days: daysMap,
	}, nil
}

// mapHourlyForDays maps API hourly data to output format for multiple days.
func (s *Service) mapHourlyForDays(hourly openmeteo.Hourly, startDate string, numDays int, loc *time.Location) (map[string]DayHours, error) {
	// Check that arrays are not nil before validating lengths
	if hourly.Time == nil {
		return nil, fmt.Errorf("%w: hourly Time array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WeatherCode == nil {
		return nil, fmt.Errorf("%w: hourly WeatherCode array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.Temperature2M == nil {
		return nil, fmt.Errorf("%w: hourly Temperature2M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.ApparentTemperature == nil {
		return nil, fmt.Errorf("%w: hourly ApparentTemperature array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.RelativeHumidity2M == nil {
		return nil, fmt.Errorf("%w: hourly RelativeHumidity2M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.Precipitation == nil {
		return nil, fmt.Errorf("%w: hourly Precipitation array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.PrecipitationProbability == nil {
		return nil, fmt.Errorf("%w: hourly PrecipitationProbability array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WindSpeed10M == nil {
		return nil, fmt.Errorf("%w: hourly WindSpeed10M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WindGusts10M == nil {
		return nil, fmt.Errorf("%w: hourly WindGusts10M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WindDirection10M == nil {
		return nil, fmt.Errorf("%w: hourly WindDirection10M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.UVIndex == nil {
		return nil, fmt.Errorf("%w: hourly UVIndex array is nil", openmeteo.ErrUpstreamAPI)
	}

	// Validate all hourly arrays have consistent lengths to prevent index out of range errors
	expectedLen := len(hourly.Time)
	if len(hourly.WeatherCode) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WeatherCode array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WeatherCode), expectedLen)
	}
	if len(hourly.Temperature2M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly Temperature2M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.Temperature2M), expectedLen)
	}
	if len(hourly.ApparentTemperature) != expectedLen {
		return nil, fmt.Errorf("%w: hourly ApparentTemperature array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.ApparentTemperature), expectedLen)
	}
	if len(hourly.RelativeHumidity2M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly RelativeHumidity2M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.RelativeHumidity2M), expectedLen)
	}
	if len(hourly.Precipitation) != expectedLen {
		return nil, fmt.Errorf("%w: hourly Precipitation array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.Precipitation), expectedLen)
	}
	if len(hourly.PrecipitationProbability) != expectedLen {
		return nil, fmt.Errorf("%w: hourly PrecipitationProbability array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.PrecipitationProbability), expectedLen)
	}
	if len(hourly.WindSpeed10M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WindSpeed10M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WindSpeed10M), expectedLen)
	}
	if len(hourly.WindGusts10M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WindGusts10M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WindGusts10M), expectedLen)
	}
	if len(hourly.WindDirection10M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WindDirection10M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WindDirection10M), expectedLen)
	}
	if len(hourly.UVIndex) != expectedLen {
		return nil, fmt.Errorf("%w: hourly UVIndex array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.UVIndex), expectedLen)
	}

	daysMap := make(map[string]DayHours, numDays)
	dateStr := startDate

	for day := 0; day < numDays; day++ {
		// Find the first index for the given date
		dateIdx := -1
		for i, t := range hourly.Time {
			if len(t) >= 10 && t[:10] == dateStr {
				dateIdx = i
				break
			}
		}

		if dateIdx == -1 {
			return nil, fmt.Errorf("no hourly data found for date %q in forecast payload: %w", dateStr, openmeteo.ErrUpstreamAPI)
		}

		// Pre-allocate with estimated size for better performance
		dayHours := make([]Hour, 0, 24)
		for i := dateIdx; i < len(hourly.Time); i++ {
			// Skip entries that don't match the date (continue, not break, for sparse data)
			if len(hourly.Time[i]) < 10 || hourly.Time[i][:10] != dateStr {
				continue
			}

			// Parse time and extract HH:MM
			// Open-Meteo API returns time without timezone offset, use local timezone from API response
			t, err := time.ParseInLocation("2006-01-02T15:04", hourly.Time[i], loc)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to parse hourly time %q at index %d: %v", openmeteo.ErrUpstreamAPI, hourly.Time[i], i, err)
			}
			timeStr := t.Format("15:04")

			dayHours = append(dayHours, Hour{
				Time:                     timeStr,
				Weather:                  s.weatherMapper.GetDescription(hourly.WeatherCode[i]),
				Temperature:              hourly.Temperature2M[i],
				ApparentTemperature:      hourly.ApparentTemperature[i],
				Humidity:                 hourly.RelativeHumidity2M[i],
				Precipitation:            hourly.Precipitation[i],
				PrecipitationProbability: hourly.PrecipitationProbability[i],
				WindSpeed:                hourly.WindSpeed10M[i],
				WindGusts:                hourly.WindGusts10M[i],
				WindDirection:            hourly.WindDirection10M[i],
				UVIndex:                  hourly.UVIndex[i],
			})
		}

		daysMap[dateStr] = DayHours{Hours: dayHours}

		// Move to the next day
		dateStr = time.Now().AddDate(0, 0, day+1).Format("2006-01-02")
	}

	return daysMap, nil
}

// fetchDailyForecast fetches daily forecast for the specified number of days (max 14).
func (s *Service) fetchDailyForecast(lat, lon float64, units string, forecastDays int, location *openmeteo.ResolvedLocation) (*DailyOutput, error) {
	if forecastDays < 1 || forecastDays > 14 {
		return nil, fmt.Errorf("daily forecast supports 1-14 days, got %d", forecastDays)
	}

	// Legacy variable set for daily compatibility
	req := openmeteo.ForecastRequest{
		Latitude:          lat,
		Longitude:         lon,
		DailyVars:         []string{"weather_code", "temperature_2m_min", "temperature_2m_max", "precipitation_sum", "precipitation_probability_max", "wind_speed_10m_max", "wind_gusts_10m_max", "uv_index_max", "sunrise", "sunset"},
		DailyForecastDays: forecastDays,
		Units:             units,
	}

	apiResp, err := s.client.FetchForecast(req)
	if err != nil {
		return nil, err
	}

	// Get timezone from API response
	loc, err := time.LoadLocation(apiResp.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid timezone %q from API: %v", openmeteo.ErrUpstreamAPI, apiResp.Timezone, err)
	}

	// Map daily conditions
	days, err := s.mapWeekDaily(apiResp.Daily, forecastDays, loc)
	if err != nil {
		return nil, err
	}

	meta := Meta{
		Units:     s.getUnits(units),
		Timezone:  apiResp.Timezone,
		Latitude:  apiResp.Latitude,
		Longitude: apiResp.Longitude,
	}

	// Include location metadata if provided via geocoding
	if location != nil {
		meta.Location = &Location{
			Name:    location.Name,
			Country: location.Country,
		}
	}

	return &DailyOutput{
		Meta: meta,
		Days: days,
	}, nil
}

// mapCurrent maps API current conditions to output format.
func (s *Service) mapCurrent(current openmeteo.Current, loc *time.Location) (Current, error) {
	// Parse time with timezone offset support
	// Open-Meteo API returns times like: "2026-03-21T12:00"
	// The API response timezone is already used to determine loc, so use ParseInLocation
	t, err := time.ParseInLocation("2006-01-02T15:04", current.Time, loc)
	if err != nil {
		return Current{}, fmt.Errorf("%w: failed to parse current time %q: %v", openmeteo.ErrUpstreamAPI, current.Time, err)
	}

	return Current{
		Time:                     t.Format("15:04"),
		Weather:                  s.weatherMapper.GetDescription(current.WeatherCode),
		Temperature:              current.Temperature2M,
		ApparentTemperature:      current.ApparentTemperature,
		Humidity:                 current.RelativeHumidity2M,
		Precipitation:            current.Precipitation,
		PrecipitationProbability: current.PrecipitationProbability,
		WindSpeed:                current.WindSpeed10M,
		WindGusts:                current.WindGusts10M,
		WindDirection:            current.WindDirection10M,
		UVIndex:                  current.UVIndex,
	}, nil
}

// mapHourly maps API hourly data to output format, filtered by date.
func (s *Service) mapHourly(hourly openmeteo.Hourly, dateStr string, loc *time.Location) ([]Hour, error) {
	// Check that arrays are not nil before validating lengths
	if hourly.Time == nil {
		return nil, fmt.Errorf("%w: hourly Time array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WeatherCode == nil {
		return nil, fmt.Errorf("%w: hourly WeatherCode array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.Temperature2M == nil {
		return nil, fmt.Errorf("%w: hourly Temperature2M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.ApparentTemperature == nil {
		return nil, fmt.Errorf("%w: hourly ApparentTemperature array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.RelativeHumidity2M == nil {
		return nil, fmt.Errorf("%w: hourly RelativeHumidity2M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.Precipitation == nil {
		return nil, fmt.Errorf("%w: hourly Precipitation array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.PrecipitationProbability == nil {
		return nil, fmt.Errorf("%w: hourly PrecipitationProbability array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WindSpeed10M == nil {
		return nil, fmt.Errorf("%w: hourly WindSpeed10M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WindGusts10M == nil {
		return nil, fmt.Errorf("%w: hourly WindGusts10M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.WindDirection10M == nil {
		return nil, fmt.Errorf("%w: hourly WindDirection10M array is nil", openmeteo.ErrUpstreamAPI)
	}
	if hourly.UVIndex == nil {
		return nil, fmt.Errorf("%w: hourly UVIndex array is nil", openmeteo.ErrUpstreamAPI)
	}

	// Validate all hourly arrays have consistent lengths to prevent index out of range errors
	expectedLen := len(hourly.Time)
	if len(hourly.WeatherCode) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WeatherCode array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WeatherCode), expectedLen)
	}
	if len(hourly.Temperature2M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly Temperature2M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.Temperature2M), expectedLen)
	}
	if len(hourly.ApparentTemperature) != expectedLen {
		return nil, fmt.Errorf("%w: hourly ApparentTemperature array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.ApparentTemperature), expectedLen)
	}
	if len(hourly.RelativeHumidity2M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly RelativeHumidity2M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.RelativeHumidity2M), expectedLen)
	}
	if len(hourly.Precipitation) != expectedLen {
		return nil, fmt.Errorf("%w: hourly Precipitation array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.Precipitation), expectedLen)
	}
	if len(hourly.PrecipitationProbability) != expectedLen {
		return nil, fmt.Errorf("%w: hourly PrecipitationProbability array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.PrecipitationProbability), expectedLen)
	}
	if len(hourly.WindSpeed10M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WindSpeed10M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WindSpeed10M), expectedLen)
	}
	if len(hourly.WindGusts10M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WindGusts10M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WindGusts10M), expectedLen)
	}
	if len(hourly.WindDirection10M) != expectedLen {
		return nil, fmt.Errorf("%w: hourly WindDirection10M array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.WindDirection10M), expectedLen)
	}
	if len(hourly.UVIndex) != expectedLen {
		return nil, fmt.Errorf("%w: hourly UVIndex array length %d does not match Time array length %d", openmeteo.ErrUpstreamAPI, len(hourly.UVIndex), expectedLen)
	}

	// Find the first index for the given date
	dateIdx := -1
	for i, t := range hourly.Time {
		if len(t) >= 10 && t[:10] == dateStr {
			dateIdx = i
			break
		}
	}

	if dateIdx == -1 {
		return nil, fmt.Errorf("no hourly data found for date %q in forecast payload: %w", dateStr, openmeteo.ErrUpstreamAPI)
	}

	// Pre-allocate with estimated size for better performance
	hours := make([]Hour, 0, 24)
	for i := dateIdx; i < len(hourly.Time); i++ {
		// Skip entries that don't match the date (continue, not break, for sparse data)
		if len(hourly.Time[i]) < 10 || hourly.Time[i][:10] != dateStr {
			continue
		}

		// Parse time and extract HH:MM
		// Open-Meteo API returns time without timezone offset, use local timezone from API response
		t, err := time.ParseInLocation("2006-01-02T15:04", hourly.Time[i], loc)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse hourly time %q at index %d: %v", openmeteo.ErrUpstreamAPI, hourly.Time[i], i, err)
		}
		timeStr := t.Format("15:04")

		hours = append(hours, Hour{
			Time:                     timeStr,
			Weather:                  s.weatherMapper.GetDescription(hourly.WeatherCode[i]),
			Temperature:              hourly.Temperature2M[i],
			ApparentTemperature:      hourly.ApparentTemperature[i],
			Humidity:                 hourly.RelativeHumidity2M[i],
			Precipitation:            hourly.Precipitation[i],
			PrecipitationProbability: hourly.PrecipitationProbability[i],
			WindSpeed:                hourly.WindSpeed10M[i],
			WindGusts:                hourly.WindGusts10M[i],
			WindDirection:            hourly.WindDirection10M[i],
			UVIndex:                  hourly.UVIndex[i],
		})
	}

	return hours, nil
}

// parseTime attempts to parse a time string using multiple formats.
// Open-Meteo API returns times like: "2026-03-21T06:30+01:00" or "2026-03-21T06:00"
// The loc parameter is used for times without timezone offset to interpret them as local time.
func parseTime(s string, loc *time.Location) (time.Time, error) {
	// Try RFC3339 with minutes timezone (with timezone offset, no seconds)
	t, err := time.Parse("2006-01-02T15:04Z07:00", s)
	if err == nil {
		return t, nil
	}
	// Try RFC3339 format (with timezone offset and seconds)
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	// Fall back to basic format without timezone offset - interpret as local time
	return time.ParseInLocation("2006-01-02T15:04", s, loc)
}

// mapDaily maps API daily data to output format.
func (s *Service) mapDaily(daily openmeteo.Daily, idx int, loc *time.Location) (Day, error) {
	// Validate index against all daily arrays to prevent out of bounds errors
	if idx < 0 {
		return Day{}, fmt.Errorf("%w: daily index %d is negative", openmeteo.ErrUpstreamAPI, idx)
	}
	if idx >= len(daily.Time) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for Time (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Time))
	}
	if idx >= len(daily.Sunrise) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for Sunrise (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Sunrise))
	}
	if idx >= len(daily.Sunset) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for Sunset (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Sunset))
	}
	if idx >= len(daily.WeatherCode) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for WeatherCode (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.WeatherCode))
	}
	if idx >= len(daily.Temperature2MMin) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for Temperature2MMin (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Temperature2MMin))
	}
	if idx >= len(daily.Temperature2MMax) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for Temperature2MMax (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Temperature2MMax))
	}
	if idx >= len(daily.PrecipitationSum) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for PrecipitationSum (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.PrecipitationSum))
	}
	if idx >= len(daily.PrecipitationProbabilityMax) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for PrecipitationProbabilityMax (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.PrecipitationProbabilityMax))
	}
	if idx >= len(daily.WindSpeed10MMax) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for WindSpeed10MMax (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.WindSpeed10MMax))
	}
	if idx >= len(daily.WindGusts10MMax) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for WindGusts10MMax (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.WindGusts10MMax))
	}
	if idx >= len(daily.UVIndexMax) {
		return Day{}, fmt.Errorf("%w: daily index %d out of bounds for UVIndexMax (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.UVIndexMax))
	}

	// Parse sunrise and sunset times
	// Open-Meteo API returns times like: "2026-03-21T06:30+01:00" or "2026-03-21T06:00"
	sunriseTime, err := parseTime(daily.Sunrise[idx], loc)
	if err != nil {
		return Day{}, fmt.Errorf("%w: failed to parse sunrise at index %d: %v", openmeteo.ErrUpstreamAPI, idx, err)
	}
	sunsetTime, err := parseTime(daily.Sunset[idx], loc)
	if err != nil {
		return Day{}, fmt.Errorf("%w: failed to parse sunset at index %d: %v", openmeteo.ErrUpstreamAPI, idx, err)
	}

	return Day{
		Date:                        daily.Time[idx],
		Weather:                     s.weatherMapper.GetDescription(daily.WeatherCode[idx]),
		TempMin:                     daily.Temperature2MMin[idx],
		TempMax:                     daily.Temperature2MMax[idx],
		PrecipitationSum:            daily.PrecipitationSum[idx],
		PrecipitationProbabilityMax: daily.PrecipitationProbabilityMax[idx],
		WindSpeedMax:                daily.WindSpeed10MMax[idx],
		WindGustsMax:                daily.WindGusts10MMax[idx],
		UVIndexMax:                  daily.UVIndexMax[idx],
		Sunrise:                     sunriseTime.Format("15:04"),
		Sunset:                      sunsetTime.Format("15:04"),
	}, nil
}

// getUnits returns unit info based on the requested units.
func (s *Service) getUnits(units string) Units {
	temperatureUnit := "C"
	windUnit := "km/h"
	precipUnit := "mm"

	if units == "imperial" {
		temperatureUnit = "F"
		windUnit = "mph"
		precipUnit = "inch"
	}

	return Units{
		Temperature:              temperatureUnit,
		Humidity:                 "%",
		WindSpeed:                windUnit,
		WindDirection:            "deg",
		Precipitation:            precipUnit,
		PrecipitationProbability: "%",
		UVIndex:                  "index",
	}
}

// mapWeekDaily maps API daily data to output format.

// mapWeekDaily maps API daily data to output format.
// maxDays limits the number of days returned (typically 7 or 14).
func (s *Service) mapWeekDaily(daily openmeteo.Daily, maxDays int, loc *time.Location) ([]Day, error) {
	if len(daily.Time) < maxDays {
		return nil, fmt.Errorf("upstream returned %d daily entries, expected %d: %w", len(daily.Time), maxDays, openmeteo.ErrUpstreamAPI)
	}

	days := make([]Day, 0, maxDays)
	for i := 0; i < maxDays && i < len(daily.Time); i++ {
		day, err := s.mapDaily(daily, i, loc)
		if err != nil {
			return nil, err
		}
		days = append(days, day)
	}
	return days, nil
}

// =============================================================================
// Variable-driven forecast shaping methods
// =============================================================================

// ForecastRequest contains parameters for a variable-driven forecast request.
type ForecastRequest struct {
	Latitude            float64
	Longitude           float64
	CurrentVars         []string
	HourlyVars          []string
	DailyVars           []string
	HourlyForecastHours int
	DailyForecastDays   int
	Units               string
	Location            *openmeteo.ResolvedLocation
}

// ForecastVariable returns a variable-driven forecast response.
func (s *Service) ForecastVariable(req ForecastRequest) (*ForecastOutput, error) {
	// Build the Open-Meteo API request
	omReq := openmeteo.ForecastRequest{
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		CurrentVars:         req.CurrentVars,
		HourlyVars:          req.HourlyVars,
		DailyVars:           req.DailyVars,
		HourlyForecastHours: req.HourlyForecastHours,
		DailyForecastDays:   req.DailyForecastDays,
		Units:               req.Units,
	}

	apiResp, err := s.client.FetchForecast(omReq)
	if err != nil {
		return nil, err
	}

	// Get timezone from API response
	loc, err := time.LoadLocation(apiResp.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid timezone %q from API: %v", openmeteo.ErrUpstreamAPI, apiResp.Timezone, err)
	}

	// Build output
	output := &ForecastOutput{
		Meta: Meta{
			Units:     s.getUnits(req.Units),
			Timezone:  apiResp.Timezone,
			Latitude:  apiResp.Latitude,
			Longitude: apiResp.Longitude,
		},
	}

	// Include location metadata if provided via geocoding
	if req.Location != nil {
		output.Meta.Location = &Location{
			Name:    req.Location.Name,
			Country: req.Location.Country,
		}
	}

	// Shape current section if requested
	if len(req.CurrentVars) > 0 && len(apiResp.Current.Time) > 0 {
		current, err := s.shapeCurrent(apiResp.Current, req.CurrentVars, loc)
		if err != nil {
			return nil, err
		}
		output.Current = current
	}

	// Shape hourly section if requested
	if len(req.HourlyVars) > 0 && len(apiResp.Hourly.Time) > 0 {
		hourly, err := s.shapeHourly(apiResp.Hourly, req.HourlyVars, loc)
		if err != nil {
			return nil, err
		}
		output.Hourly = hourly
	}

	// Shape daily section if requested
	if len(req.DailyVars) > 0 && len(apiResp.Daily.Time) > 0 {
		daily, err := s.shapeDaily(apiResp.Daily, req.DailyVars, loc)
		if err != nil {
			return nil, err
		}
		output.Daily = daily
	}

	return output, nil
}

// shapeCurrent creates a current output section from the API response.
func (s *Service) shapeCurrent(current openmeteo.Current, vars []string, loc *time.Location) (*CurrentOutput, error) {
	// Parse time
	t, err := time.ParseInLocation("2006-01-02T15:04", current.Time, loc)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse current time %q: %v", openmeteo.ErrUpstreamAPI, current.Time, err)
	}
	timeStr := t.Format("2006-01-02T15:04")

	// Build field order: time is always first
	fields := make([]string, 0, len(vars)+1)
	fields = append(fields, "time")

	// Build values map - using flat structure with direct values
	values := make(map[string]any)
	values["time"] = timeStr

	// Add requested variables
	for _, v := range vars {
		// Skip time as we already added it
		if v == "time" {
			continue
		}
		val, outputName, err := s.getCurrentValue(current, v)
		if err != nil {
			return nil, err
		}
		values[outputName] = val
		fields = append(fields, outputName)
	}

	return &CurrentOutput{
		Fields: fields,
		Values: values,
	}, nil
}

// getCurrentValue extracts a value from the Current API response for a given variable.
// Returns the value, the output field name, and an error. For weather_code, the output
// field name is "weather" while other variables use their request variable name.
func (s *Service) getCurrentValue(current openmeteo.Current, variable string) (any, string, error) {
	// Handle weather_code specially - render as text and use "weather" as output field name
	if variable == "weather_code" {
		return s.weatherMapper.GetDescription(current.WeatherCode), "weather", nil
	}

	switch variable {
	case "temperature_2m":
		return current.Temperature2M, variable, nil
	case "apparent_temperature":
		return current.ApparentTemperature, variable, nil
	case "relative_humidity_2m":
		return current.RelativeHumidity2M, variable, nil
	case "precipitation":
		return current.Precipitation, variable, nil
	case "precipitation_probability":
		return current.PrecipitationProbability, variable, nil
	case "wind_speed_10m":
		return current.WindSpeed10M, variable, nil
	case "wind_gusts_10m":
		return current.WindGusts10M, variable, nil
	case "wind_direction_10m":
		return current.WindDirection10M, variable, nil
	case "uv_index":
		return current.UVIndex, variable, nil
	default:
		return nil, variable, fmt.Errorf("unsupported current variable: %s", variable)
	}
}

// shapeHourly creates an hourly output section from the API response.
func (s *Service) shapeHourly(hourly openmeteo.Hourly, vars []string, loc *time.Location) (*HourlyOutputNew, error) {
	if len(hourly.Time) == 0 {
		return nil, fmt.Errorf("%w: hourly time array is empty", openmeteo.ErrUpstreamAPI)
	}

	// Group by date
	daysMap := make(map[string]*HourlyDay)

	for i, timeStr := range hourly.Time {
		if len(timeStr) < 10 {
			return nil, fmt.Errorf("%w: invalid time format at index %d: %q", openmeteo.ErrUpstreamAPI, i, timeStr)
		}

		dateStr := timeStr[:10]
		day, ok := daysMap[dateStr]
		if !ok {
			// Parse date for weekday
			t, err := time.ParseInLocation("2006-01-02T15:04", timeStr, loc)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to parse hourly time %q at index %d: %v", openmeteo.ErrUpstreamAPI, timeStr, i, err)
			}
			weekday := t.Format("Mon")

			// Build field order for day: date, weekday, then requested vars (using output names)
			fields := make([]string, 0, len(vars)+2)
			fields = append(fields, "date", "weekday")
			for _, v := range vars {
				if v != "date" && v != "weekday" {
					outputName := v
					if v == "weather_code" {
						outputName = "weather"
					}
					fields = append(fields, outputName)
				}
			}

			day = &HourlyDay{
				Fields: fields,
				Values: map[string]any{
					"date":    dateStr,
					"weekday": weekday,
				},
				Hours: make([]HourlyEntry, 0),
			}
			daysMap[dateStr] = day
		}

		// Parse time for hour entry
		t, err := time.ParseInLocation("2006-01-02T15:04", timeStr, loc)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse hourly time %q at index %d: %v", openmeteo.ErrUpstreamAPI, timeStr, i, err)
		}
		hourTime := t.Format("15:04")

		// Build hour field order: time, then requested vars (using output names)
		hourFields := make([]string, 0, len(vars)+1)
		hourFields = append(hourFields, "time")
		for _, v := range vars {
			if v != "time" {
				outputName := v
				if v == "weather_code" {
					outputName = "weather"
				}
				hourFields = append(hourFields, outputName)
			}
		}

		hourValues := make(map[string]any)
		hourValues["time"] = hourTime

		// Add requested variables
		for _, v := range vars {
			if v == "time" {
				continue
			}
			val, outputName, err := s.getHourlyValue(hourly, v, i)
			if err != nil {
				return nil, err
			}
			hourValues[outputName] = val
		}

		day.Hours = append(day.Hours, HourlyEntry{
			Fields: hourFields,
			Values: hourValues,
		})
	}

	// Convert map to slice in date order
	days := make([]HourlyDay, 0, len(daysMap))
	for _, dateStr := range sortedDateKeys(daysMap) {
		days = append(days, *daysMap[dateStr])
	}

	return &HourlyOutputNew{Days: days}, nil
}

// getHourlyValue extracts a value from the Hourly API response for a given variable at index.
// Returns the value, the output field name, and an error. For weather_code, the output
// field name is "weather" while other variables use their request variable name.
func (s *Service) getHourlyValue(hourly openmeteo.Hourly, variable string, idx int) (any, string, error) {
	// Handle weather_code specially - render as text and use "weather" as output field name
	if variable == "weather_code" {
		if idx >= len(hourly.WeatherCode) {
			return nil, "weather", fmt.Errorf("%w: weather_code index %d out of bounds (length %d)", openmeteo.ErrUpstreamAPI, idx, len(hourly.WeatherCode))
		}
		return s.weatherMapper.GetDescription(hourly.WeatherCode[idx]), "weather", nil
	}

	// Helper to check bounds
	checkBounds := func(arr interface{}, name string) error {
		var length int
		switch a := arr.(type) {
		case []string:
			length = len(a)
		case []float64:
			length = len(a)
		case []int:
			length = len(a)
		}
		if idx >= length {
			return fmt.Errorf("%w: %s index %d out of bounds (length %d)", openmeteo.ErrUpstreamAPI, name, idx, length)
		}
		return nil
	}

	switch variable {
	case "temperature_2m":
		if err := checkBounds(hourly.Temperature2M, variable); err != nil {
			return nil, variable, err
		}
		return hourly.Temperature2M[idx], variable, nil
	case "apparent_temperature":
		if err := checkBounds(hourly.ApparentTemperature, variable); err != nil {
			return nil, variable, err
		}
		return hourly.ApparentTemperature[idx], variable, nil
	case "relative_humidity_2m":
		if err := checkBounds(hourly.RelativeHumidity2M, variable); err != nil {
			return nil, variable, err
		}
		return hourly.RelativeHumidity2M[idx], variable, nil
	case "precipitation":
		if err := checkBounds(hourly.Precipitation, variable); err != nil {
			return nil, variable, err
		}
		return hourly.Precipitation[idx], variable, nil
	case "precipitation_probability":
		if err := checkBounds(hourly.PrecipitationProbability, variable); err != nil {
			return nil, variable, err
		}
		return hourly.PrecipitationProbability[idx], variable, nil
	case "wind_speed_10m":
		if err := checkBounds(hourly.WindSpeed10M, variable); err != nil {
			return nil, variable, err
		}
		return hourly.WindSpeed10M[idx], variable, nil
	case "wind_gusts_10m":
		if err := checkBounds(hourly.WindGusts10M, variable); err != nil {
			return nil, variable, err
		}
		return hourly.WindGusts10M[idx], variable, nil
	case "wind_direction_10m":
		if err := checkBounds(hourly.WindDirection10M, variable); err != nil {
			return nil, variable, err
		}
		return hourly.WindDirection10M[idx], variable, nil
	case "uv_index":
		if err := checkBounds(hourly.UVIndex, variable); err != nil {
			return nil, variable, err
		}
		return hourly.UVIndex[idx], variable, nil
	default:
		return nil, variable, fmt.Errorf("unsupported hourly variable: %s", variable)
	}
}

// shapeDaily creates a daily output section from the API response.
func (s *Service) shapeDaily(daily openmeteo.Daily, vars []string, loc *time.Location) (*DailyOutputNew, error) {
	if len(daily.Time) == 0 {
		return nil, fmt.Errorf("%w: daily time array is empty", openmeteo.ErrUpstreamAPI)
	}

	// Build field order: date, weekday, then requested vars (using output names)
	fields := make([]string, 0, len(vars)+2)
	fields = append(fields, "date", "weekday")
	for _, v := range vars {
		if v != "date" && v != "weekday" {
			outputName := v
			if v == "weather_code" {
				outputName = "weather"
			}
			fields = append(fields, outputName)
		}
	}

	// Build rows
	rows := make([]DailyRow, 0, len(daily.Time))
	for i, dateStr := range daily.Time {
		// Parse date for weekday
		t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse daily date %q at index %d: %v", openmeteo.ErrUpstreamAPI, dateStr, i, err)
		}
		weekday := t.Format("Mon")

		values := make(map[string]any)
		values["date"] = dateStr
		values["weekday"] = weekday

		// Add requested variables
		for _, v := range vars {
			if v == "date" || v == "weekday" {
				continue
			}
			val, outputName, err := s.getDailyValue(daily, v, i, loc)
			if err != nil {
				return nil, err
			}
			values[outputName] = val
		}

		rows = append(rows, DailyRow{Values: values})
	}

	return &DailyOutputNew{
		Fields: fields,
		Rows:   rows,
	}, nil
}

// getDailyValue extracts a value from the Daily API response for a given variable at index.
// Returns the value, the output field name, and an error. For weather_code, the output
// field name is "weather" while other variables use their request variable name.
func (s *Service) getDailyValue(daily openmeteo.Daily, variable string, idx int, loc *time.Location) (any, string, error) {
	// Handle weather_code specially - render as text and use "weather" as output field name
	if variable == "weather_code" {
		if idx >= len(daily.WeatherCode) {
			return nil, "weather", fmt.Errorf("%w: weather_code index %d out of bounds (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.WeatherCode))
		}
		return s.weatherMapper.GetDescription(daily.WeatherCode[idx]), "weather", nil
	}

	// Helper to check bounds
	checkBounds := func(arr interface{}, name string) error {
		var length int
		switch a := arr.(type) {
		case []string:
			length = len(a)
		case []float64:
			length = len(a)
		case []int:
			length = len(a)
		}
		if idx >= length {
			return fmt.Errorf("%w: %s index %d out of bounds (length %d)", openmeteo.ErrUpstreamAPI, name, idx, length)
		}
		return nil
	}

	switch variable {
	case "temperature_2m_max":
		if err := checkBounds(daily.Temperature2MMax, variable); err != nil {
			return nil, variable, err
		}
		return daily.Temperature2MMax[idx], variable, nil
	case "temperature_2m_min":
		if err := checkBounds(daily.Temperature2MMin, variable); err != nil {
			return nil, variable, err
		}
		return daily.Temperature2MMin[idx], variable, nil
	case "apparent_temperature_max":
		// Note: Open-Meteo API doesn't return this in the current response structure
		return nil, variable, fmt.Errorf("unsupported daily variable: %s", variable)
	case "apparent_temperature_min":
		return nil, variable, fmt.Errorf("unsupported daily variable: %s", variable)
	case "precipitation_sum":
		if err := checkBounds(daily.PrecipitationSum, variable); err != nil {
			return nil, variable, err
		}
		return daily.PrecipitationSum[idx], variable, nil
	case "precipitation_probability_max":
		if err := checkBounds(daily.PrecipitationProbabilityMax, variable); err != nil {
			return nil, variable, err
		}
		return daily.PrecipitationProbabilityMax[idx], variable, nil
	case "precipitation_hours":
		return nil, variable, fmt.Errorf("unsupported daily variable: %s", variable)
	case "wind_speed_10m_max":
		if err := checkBounds(daily.WindSpeed10MMax, variable); err != nil {
			return nil, variable, err
		}
		return daily.WindSpeed10MMax[idx], variable, nil
	case "wind_gusts_10m_max":
		if err := checkBounds(daily.WindGusts10MMax, variable); err != nil {
			return nil, variable, err
		}
		return daily.WindGusts10MMax[idx], variable, nil
	case "uv_index_max":
		if err := checkBounds(daily.UVIndexMax, variable); err != nil {
			return nil, variable, err
		}
		return daily.UVIndexMax[idx], variable, nil
	case "sunrise":
		if idx >= len(daily.Sunrise) {
			return nil, variable, fmt.Errorf("%w: sunrise index %d out of bounds (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Sunrise))
		}
		// Parse and format sunrise time
		sunriseTime, err := parseTime(daily.Sunrise[idx], loc)
		if err != nil {
			return nil, variable, err
		}
		return sunriseTime.Format("15:04"), variable, nil
	case "sunset":
		if idx >= len(daily.Sunset) {
			return nil, variable, fmt.Errorf("%w: sunset index %d out of bounds (length %d)", openmeteo.ErrUpstreamAPI, idx, len(daily.Sunset))
		}
		// Parse and format sunset time
		sunsetTime, err := parseTime(daily.Sunset[idx], loc)
		if err != nil {
			return nil, variable, err
		}
		return sunsetTime.Format("15:04"), variable, nil
	case "daylight_duration":
		return nil, variable, fmt.Errorf("unsupported daily variable: %s", variable)
	case "sunshine_duration":
		return nil, variable, fmt.Errorf("unsupported daily variable: %s", variable)
	default:
		return nil, variable, fmt.Errorf("unsupported daily variable: %s", variable)
	}
}

// sortedDateKeys returns the keys of a date map in sorted order.
func sortedDateKeys(m map[string]*HourlyDay) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
