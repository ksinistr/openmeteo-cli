package forecast

import (
	"fmt"
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
func (s *Service) Forecast(lat, lon float64, units string, hourly bool, forecastDays int) (interface{}, error) {
	if hourly {
		return s.fetchHourlyForecast(lat, lon, units, forecastDays)
	}
	return s.fetchDailyForecast(lat, lon, units, forecastDays)
}

// fetchHourlyForecast fetches hourly forecast for the specified number of days (max 2).
func (s *Service) fetchHourlyForecast(lat, lon float64, units string, forecastDays int) (*HourlyOutput, error) {
	if forecastDays < 1 || forecastDays > 2 {
		return nil, fmt.Errorf("hourly forecast supports 1-2 days, got %d", forecastDays)
	}

	now := time.Now()
	apiResp, err := s.client.FetchForecast(lat, lon, units, "auto", forecastDays)
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
	hours, err := s.mapHourlyForDays(apiResp.Hourly, today, forecastDays, loc)
	if err != nil {
		return nil, err
	}

	return &HourlyOutput{
		Meta: Meta{
			GeneratedAt: now,
			Units:       s.getUnits(units),
			Timezone:    apiResp.Timezone,
			Latitude:    apiResp.Latitude,
			Longitude:   apiResp.Longitude,
		},
		Hours: hours,
	}, nil
}

// mapHourlyForDays maps API hourly data to output format for multiple days.
func (s *Service) mapHourlyForDays(hourly openmeteo.Hourly, startDate string, numDays int, loc *time.Location) ([]Hour, error) {
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

	hours := make([]Hour, 0, numDays*24)
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

		// Pre-allocate with estimated size for better performance - use dayHours for potential future use
		_ = make([]Hour, 0, 24)
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

		// Move to the next day
		dateStr = time.Now().AddDate(0, 0, day+1).Format("2006-01-02")
	}

	return hours, nil
}

// fetchDailyForecast fetches daily forecast for the specified number of days (max 14).
func (s *Service) fetchDailyForecast(lat, lon float64, units string, forecastDays int) (*DailyOutput, error) {
	if forecastDays < 1 || forecastDays > 14 {
		return nil, fmt.Errorf("daily forecast supports 1-14 days, got %d", forecastDays)
	}

	now := time.Now()
	apiResp, err := s.client.FetchForecast(lat, lon, units, "auto", forecastDays)
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

	return &DailyOutput{
		Meta: Meta{
			GeneratedAt: now,
			Units:       s.getUnits(units),
			Timezone:    apiResp.Timezone,
			Latitude:    apiResp.Latitude,
			Longitude:   apiResp.Longitude,
		},
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
		Sunrise:                     sunriseTime.Format(time.RFC3339),
		Sunset:                      sunsetTime.Format(time.RFC3339),
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
