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

// Today returns today's forecast.
func (s *Service) Today(lat, lon float64, units string) (*TodayOutput, error) {
	apiResp, err := s.client.FetchForecast(lat, lon, units, "auto", 1)
	if err != nil {
		return nil, err
	}

	// Get timezone from API response
	loc, err := time.LoadLocation(apiResp.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Map current conditions
	current := s.mapCurrent(apiResp.Current, loc)

	// Map hourly conditions for today
	hours := s.mapHourly(apiResp.Hourly, apiResp.Current.Time[:10], loc)

	return &TodayOutput{
		Meta: Meta{
			GeneratedAt: time.Now(),
			Units:       s.getUnits(units),
			Timezone:    apiResp.Timezone,
			Latitude:    apiResp.Latitude,
			Longitude:   apiResp.Longitude,
		},
		Current: current,
		Hours:   hours,
	}, nil
}

// Day returns forecast for a specific date.
func (s *Service) Day(lat, lon float64, date time.Time, units string) (*DayOutput, error) {
	// Use date as YYYY-MM-DD string
	dateStr := date.Format("2006-01-02")

	apiResp, err := s.client.FetchForecast(lat, lon, units, "auto", 16)
	if err != nil {
		return nil, err
	}

	// Get timezone from API response
	loc, err := time.LoadLocation(apiResp.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Find the daily entry for the requested date
	dailyIdx := -1
	for i, dDate := range apiResp.Daily.Time {
		if dDate == dateStr {
			dailyIdx = i
			break
		}
	}

	if dailyIdx == -1 {
		return nil, ErrDateUnavailable
	}

	// Map daily conditions
	day := s.mapDaily(apiResp.Daily, dailyIdx, loc)

	// Map hourly conditions for the requested date
	hours := s.mapHourly(apiResp.Hourly, dateStr, loc)

	return &DayOutput{
		Meta: Meta{
			GeneratedAt: time.Now(),
			Units:       s.getUnits(units),
			Timezone:    apiResp.Timezone,
			Latitude:    apiResp.Latitude,
			Longitude:   apiResp.Longitude,
		},
		Day:   day,
		Hours: hours,
	}, nil
}

// Week returns a 7-day forecast starting from today.
func (s *Service) Week(lat, lon float64, units string) (*WeekOutput, error) {
	apiResp, err := s.client.FetchForecast(lat, lon, units, "auto", 7)
	if err != nil {
		return nil, err
	}

	// Get timezone from API response
	loc, err := time.LoadLocation(apiResp.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Map daily conditions for 7 days
	days := make([]Day, 0, 7)
	for i := 0; i < 7 && i < len(apiResp.Daily.Time); i++ {
		days = append(days, s.mapDaily(apiResp.Daily, i, loc))
	}

	return &WeekOutput{
		Meta: Meta{
			GeneratedAt: time.Now(),
			Units:       s.getUnits(units),
			Timezone:    apiResp.Timezone,
			Latitude:    apiResp.Latitude,
			Longitude:   apiResp.Longitude,
		},
		Days: days,
	}, nil
}

// mapCurrent maps API current conditions to output format.
func (s *Service) mapCurrent(current openmeteo.Current, loc *time.Location) Current {
	// Parse time and convert to local time
	t, _ := time.Parse("2006-01-02T15:04", current.Time)
	localTime := t.In(loc).Format("15:04")

	return Current{
		Time:                     localTime,
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
	}
}

// mapHourly maps API hourly data to output format, filtered by date.
func (s *Service) mapHourly(hourly openmeteo.Hourly, dateStr string, loc *time.Location) []Hour {
	// Find the first index for the given date
	dateIdx := -1
	for i, t := range hourly.Time {
		if t[:10] == dateStr {
			dateIdx = i
			break
		}
	}

	if dateIdx == -1 {
		return nil
	}

	// Extract all entries for this date
	var hours []Hour
	for i := dateIdx; i < len(hourly.Time); i++ {
		// Stop when we move to the next date
		if hourly.Time[i][:10] != dateStr {
			break
		}

		// Parse time and extract HH:MM
		t, _ := time.Parse("2006-01-02T15:04", hourly.Time[i])
		timeStr := t.In(loc).Format("15:04")

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

	return hours
}

// mapDaily maps API daily data to output format.
func (s *Service) mapDaily(daily openmeteo.Daily, idx int, loc *time.Location) Day {
	sunriseTime, _ := time.Parse("2006-01-02T15:04", daily.Sunrise[idx])
	sunsetTime, _ := time.Parse("2006-01-02T15:04", daily.Sunset[idx])

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
		Sunrise:                     sunriseTime.In(loc).Format("2006-01-02T15:04"),
		Sunset:                      sunsetTime.In(loc).Format("2006-01-02T15:04"),
	}
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
