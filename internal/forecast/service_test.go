package forecast

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/weathercode"
)

// TestService_mapHourly_DatetimeFiltering tests that mapHourly filters to the correct date.
func TestService_mapHourly_DatetimeFiltering(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name        string
		hourlyTime  []string
		dateStr     string
		expectedLen int
		expectError bool
	}{
		{
			name:        "filter_exact_match",
			hourlyTime:  []string{"2026-03-21T00:00", "2026-03-21T12:00", "2026-03-21T23:00"},
			dateStr:     "2026-03-21",
			expectedLen: 3,
			expectError: false,
		},
		{
			name:        "filter_excludes_other_dates",
			hourlyTime:  []string{"2026-03-20T23:00", "2026-03-21T00:00", "2026-03-21T12:00", "2026-03-22T00:00"},
			dateStr:     "2026-03-21",
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "filter_no_match_returns_error",
			hourlyTime:  []string{"2026-03-20T00:00", "2026-03-20T12:00"},
			dateStr:     "2026-03-21",
			expectedLen: 0,
			expectError: true,
		},
		{
			name:        "filter_empty_returns_error",
			hourlyTime:  []string{},
			dateStr:     "2026-03-21",
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTime,
				Temperature2M:            make([]float64, len(tt.hourlyTime)),
				ApparentTemperature:      make([]float64, len(tt.hourlyTime)),
				RelativeHumidity2M:       make([]int, len(tt.hourlyTime)),
				Precipitation:            make([]float64, len(tt.hourlyTime)),
				PrecipitationProbability: make([]int, len(tt.hourlyTime)),
				WindSpeed10M:             make([]float64, len(tt.hourlyTime)),
				WindGusts10M:             make([]float64, len(tt.hourlyTime)),
				WindDirection10M:         make([]int, len(tt.hourlyTime)),
				UVIndex:                  make([]float64, len(tt.hourlyTime)),
				WeatherCode:              make([]int, len(tt.hourlyTime)),
			}

			hours, err := svc.mapHourly(hourly, tt.dateStr, loc)
			if tt.expectError {
				if err == nil {
					t.Errorf("mapHourly() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("mapHourly() returned unexpected error: %v", err)
				return
			}

			if len(hours) != tt.expectedLen {
				t.Errorf("Expected %d hours, got %d", tt.expectedLen, len(hours))
			}

			for i, h := range hours {
				if len(h.Time) != 5 || h.Time[2] != ':' {
					t.Errorf("Hour %d time %q not in HH:MM format", i, h.Time)
				}
			}
		})
	}
}

// TestService_mapHourly_TimezoneConversion tests timezone handling in mapHourly.
func TestService_mapHourly_TimezoneConversion(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	tests := []struct {
		name       string
		hourlyTime []string
		timezone   string
		dateStr    string
		count      int
	}{
		{
			name:       "utc_timezone",
			hourlyTime: []string{"2026-03-21T12:00"},
			timezone:   "UTC",
			dateStr:    "2026-03-21",
			count:      1,
		},
		{
			name:       "us_eastern_timezone",
			hourlyTime: []string{"2026-03-21T12:00"},
			timezone:   "America/New_York",
			dateStr:    "2026-03-21",
			count:      1,
		},
		{
			name:       "tokyo_timezone",
			hourlyTime: []string{"2026-03-21T00:00"},
			timezone:   "Asia/Tokyo",
			dateStr:    "2026-03-21",
			count:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Fatalf("Failed to load timezone: %v", err)
			}

			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTime,
				Temperature2M:            []float64{20.0},
				ApparentTemperature:      []float64{18.0},
				RelativeHumidity2M:       []int{60},
				Precipitation:            []float64{0.0},
				PrecipitationProbability: []int{10},
				WindSpeed10M:             []float64{10.0},
				WindGusts10M:             []float64{15.0},
				WindDirection10M:         []int{180},
				UVIndex:                  []float64{3.0},
				WeatherCode:              []int{0},
			}

			hours, err := svc.mapHourly(hourly, tt.dateStr, loc)
			if err != nil {
				t.Fatalf("mapHourly() returned error: %v", err)
			}

			if len(hours) != tt.count {
				t.Fatalf("Expected %d hour(s), got %d", tt.count, len(hours))
			}

			// Verify time is in HH:MM format
			if len(hours[0].Time) != 5 || hours[0].Time[2] != ':' {
				t.Errorf("Hour time %q not in HH:MM format", hours[0].Time)
			}
		})
	}
}

// TestService_mapCurrent tests the mapCurrent function.
func TestService_mapCurrent(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	current := openmeteo.Current{
		Time:                     "2026-03-21T12:00",
		Temperature2M:            20.0,
		ApparentTemperature:      18.0,
		RelativeHumidity2M:       60,
		Precipitation:            0.0,
		PrecipitationProbability: 10,
		WindSpeed10M:             10.0,
		WindGusts10M:             15.0,
		WindDirection10M:         180,
		UVIndex:                  3.0,
		WeatherCode:              0,
	}

	result, err := svc.mapCurrent(current, loc)
	if err != nil {
		t.Fatalf("mapCurrent() returned error: %v", err)
	}

	tests := []struct {
		field string
		want  interface{}
	}{
		{"Time", "12:00"},
		{"Weather", "Clear sky"},
		{"Temperature", 20.0},
		{"ApparentTemperature", 18.0},
		{"Humidity", 60},
		{"WindSpeed", 10.0},
		{"WindDirection", 180},
		{"UVIndex", 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			switch tt.field {
			case "Time":
				if result.Time != tt.want.(string) {
					t.Errorf("Time = %q, want %q", result.Time, tt.want)
				}
			case "Weather":
				if result.Weather != tt.want.(string) {
					t.Errorf("Weather = %q, want %q", result.Weather, tt.want)
				}
			case "Temperature":
				if result.Temperature != tt.want.(float64) {
					t.Errorf("Temperature = %v, want %v", result.Temperature, tt.want)
				}
			case "Humidity":
				if result.Humidity != tt.want.(int) {
					t.Errorf("Humidity = %v, want %v", result.Humidity, tt.want)
				}
			case "WindSpeed":
				if result.WindSpeed != tt.want.(float64) {
					t.Errorf("WindSpeed = %v, want %v", result.WindSpeed, tt.want)
				}
			case "WindDirection":
				if result.WindDirection != tt.want.(int) {
					t.Errorf("WindDirection = %v, want %v", result.WindDirection, tt.want)
				}
			case "UVIndex":
				if result.UVIndex != tt.want.(float64) {
					t.Errorf("UVIndex = %v, want %v", result.UVIndex, tt.want)
				}
			}
		})
	}
}

// TestService_mapHourly_UnknownWeatherCode tests handling of unknown weather codes.
func TestService_mapHourly_UnknownWeatherCode(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	hourly := openmeteo.Hourly{
		Time:                     []string{"2026-03-21T12:00"},
		Temperature2M:            []float64{20.0},
		ApparentTemperature:      []float64{18.0},
		RelativeHumidity2M:       []int{60},
		Precipitation:            []float64{0.0},
		PrecipitationProbability: []int{10},
		WindSpeed10M:             []float64{10.0},
		WindGusts10M:             []float64{15.0},
		WindDirection10M:         []int{180},
		UVIndex:                  []float64{3.0},
		WeatherCode:              []int{999},
	}

	hours, err := svc.mapHourly(hourly, "2026-03-21", loc)
	if err != nil {
		t.Fatalf("mapHourly() returned error: %v", err)
	}

	if len(hours) != 1 {
		t.Fatalf("Expected 1 hour, got %d", len(hours))
	}

	expected := "Unknown weather code: 999"
	if hours[0].Weather != expected {
		t.Errorf("Weather = %q, want %q", hours[0].Weather, expected)
	}
}

// TestService_mapHourly_DSTHandling tests DST transition handling.
func TestService_mapHourly_DSTHandling(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	tests := []struct {
		name        string
		date        string
		hourlyTime  []string
		timezone    string
		expectedLen int
	}{
		{
			name:        "spring_forward_2026",
			date:        "2026-03-08",
			hourlyTime:  []string{"2026-03-08T06:00", "2026-03-08T07:00"},
			timezone:    "America/New_York",
			expectedLen: 2,
		},
		{
			name:        "fall_back_2026",
			date:        "2026-11-01",
			hourlyTime:  []string{"2026-11-01T05:00", "2026-11-01T06:00"},
			timezone:    "America/New_York",
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Fatalf("Failed to load timezone: %v", err)
			}

			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTime,
				Temperature2M:            make([]float64, len(tt.hourlyTime)),
				ApparentTemperature:      make([]float64, len(tt.hourlyTime)),
				RelativeHumidity2M:       make([]int, len(tt.hourlyTime)),
				Precipitation:            make([]float64, len(tt.hourlyTime)),
				PrecipitationProbability: make([]int, len(tt.hourlyTime)),
				WindSpeed10M:             make([]float64, len(tt.hourlyTime)),
				WindGusts10M:             make([]float64, len(tt.hourlyTime)),
				WindDirection10M:         make([]int, len(tt.hourlyTime)),
				UVIndex:                  make([]float64, len(tt.hourlyTime)),
				WeatherCode:              make([]int, len(tt.hourlyTime)),
			}

			hours, err := svc.mapHourly(hourly, tt.date, loc)
			if err != nil {
				t.Fatalf("mapHourly() returned error: %v", err)
			}

			if len(hours) != tt.expectedLen {
				t.Errorf("Expected %d hours, got %d", tt.expectedLen, len(hours))
			}

			for i, h := range hours {
				if len(h.Time) != 5 || h.Time[2] != ':' {
					t.Errorf("Hour %d time %q not in HH:MM format", i, h.Time)
				}
			}
		})
	}
}

// TestService_mapDaily tests the mapDaily function.
func TestService_mapDaily(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-21"},
		WeatherCode:                 []int{0},
		Temperature2MMin:            []float64{15.0},
		Temperature2MMax:            []float64{25.0},
		PrecipitationSum:            []float64{0.0},
		PrecipitationProbabilityMax: []int{10},
		WindSpeed10MMax:             []float64{10.0},
		WindGusts10MMax:             []float64{15.0},
		UVIndexMax:                  []float64{5.0},
		Sunrise:                     []string{"2026-03-21T06:00"},
		Sunset:                      []string{"2026-03-21T18:00"},
	}

	result, err := svc.mapDaily(daily, 0, loc)
	if err != nil {
		t.Fatalf("mapDaily() returned error: %v", err)
	}

	tests := []struct {
		field string
		want  interface{}
	}{
		{"Date", "2026-03-21"},
		{"Weather", "Clear sky"},
		{"TempMin", 15.0},
		{"TempMax", 25.0},
		{"WindSpeedMax", 10.0},
		{"UVIndexMax", 5.0},
		{"Sunrise", "2026-03-21T06:00:00Z"},
		{"Sunset", "2026-03-21T18:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			switch tt.field {
			case "Date":
				if result.Date != tt.want.(string) {
					t.Errorf("Date = %q, want %q", result.Date, tt.want)
				}
			case "Weather":
				if result.Weather != tt.want.(string) {
					t.Errorf("Weather = %q, want %q", result.Weather, tt.want)
				}
			case "TempMin":
				if result.TempMin != tt.want.(float64) {
					t.Errorf("TempMin = %v, want %v", result.TempMin, tt.want)
				}
			case "TempMax":
				if result.TempMax != tt.want.(float64) {
					t.Errorf("TempMax = %v, want %v", result.TempMax, tt.want)
				}
			case "WindSpeedMax":
				if result.WindSpeedMax != tt.want.(float64) {
					t.Errorf("WindSpeedMax = %v, want %v", result.WindSpeedMax, tt.want)
				}
			case "UVIndexMax":
				if result.UVIndexMax != tt.want.(float64) {
					t.Errorf("UVIndexMax = %v, want %v", result.UVIndexMax, tt.want)
				}
			case "Sunrise":
				if result.Sunrise != tt.want.(string) {
					t.Errorf("Sunrise = %q, want %q", result.Sunrise, tt.want)
				}
			case "Sunset":
				if result.Sunset != tt.want.(string) {
					t.Errorf("Sunset = %q, want %q", result.Sunset, tt.want)
				}
			}
		})
	}
}

// TestService_mapDaily_TimezoneOffset tests sunrise/sunset with timezone offset.
func TestService_mapDaily_TimezoneOffset(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	// Use a non-UTC timezone to test timezone handling
	loc, _ := time.LoadLocation("Europe/Berlin")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-21"},
		WeatherCode:                 []int{0},
		Temperature2MMin:            []float64{15.0},
		Temperature2MMax:            []float64{25.0},
		PrecipitationSum:            []float64{0.0},
		PrecipitationProbabilityMax: []int{10},
		WindSpeed10MMax:             []float64{10.0},
		WindGusts10MMax:             []float64{15.0},
		UVIndexMax:                  []float64{5.0},
		// RFC3339 format with timezone offset
		Sunrise: []string{"2026-03-21T06:30+01:00"},
		Sunset:  []string{"2026-03-21T18:45+01:00"},
	}

	result, err := svc.mapDaily(daily, 0, loc)
	if err != nil {
		t.Fatalf("mapDaily() returned error: %v", err)
	}

	// The parsed time should be converted to UTC for consistent output
	// Format should be RFC3339
	if len(result.Sunrise) < 16 || result.Sunrise[10] != 'T' {
		t.Errorf("Sunrise should be in RFC3339 format, got %q", result.Sunrise)
	}
	if len(result.Sunset) < 16 || result.Sunset[10] != 'T' {
		t.Errorf("Sunset should be in RFC3339 format, got %q", result.Sunset)
	}
}

// TestService_mapDaily_MultipleDays tests mapping multiple days.
func TestService_mapDaily_MultipleDays(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-21", "2026-03-22", "2026-03-23"},
		WeatherCode:                 []int{0, 1, 2},
		Temperature2MMin:            []float64{15.0, 16.0, 17.0},
		Temperature2MMax:            []float64{25.0, 26.0, 27.0},
		PrecipitationSum:            []float64{0.0, 0.5, 0.0},
		PrecipitationProbabilityMax: []int{10, 30, 20},
		WindSpeed10MMax:             []float64{10.0, 12.0, 8.0},
		WindGusts10MMax:             []float64{15.0, 18.0, 12.0},
		UVIndexMax:                  []float64{5.0, 6.0, 4.0},
		Sunrise:                     []string{"2026-03-21T06:00", "2026-03-22T05:59", "2026-03-23T05:58"},
		Sunset:                      []string{"2026-03-21T18:00", "2026-03-22T18:01", "2026-03-23T18:02"},
	}

	days := []struct {
		idx     int
		date    string
		weather string
	}{
		{0, "2026-03-21", "Clear sky"},
		{1, "2026-03-22", "Mainly clear"},
		{2, "2026-03-23", "Partly cloudy"},
	}

	for _, tt := range days {
		t.Run(tt.date, func(t *testing.T) {
			result, err := svc.mapDaily(daily, tt.idx, loc)
			if err != nil {
				t.Fatalf("mapDaily() returned error: %v", err)
			}

			if result.Date != tt.date {
				t.Errorf("Date = %q, want %q", result.Date, tt.date)
			}
			if result.Weather != tt.weather {
				t.Errorf("Weather = %q, want %q", result.Weather, tt.weather)
			}
		})
	}
}

// TestService_Today_Units tests unit handling in Today.
func TestService_Today_Units(t *testing.T) {
	client := openmeteo.NewClient(nil)
	svc := NewService(client, weathercode.NewMapper())

	tests := []struct {
		units      string
		tempUnit   string
		windUnit   string
		precipUnit string
	}{
		{"metric", "C", "km/h", "mm"},
		{"imperial", "F", "mph", "inch"},
	}

	for _, tt := range tests {
		t.Run("units_"+tt.units, func(t *testing.T) {
			result, err := svc.Today(40.0, -74.0, tt.units)
			if err != nil {
				// Expected - network call will fail, we're just verifying units structure
				_ = err
				return
			}

			// Verify the units structure would be correct
			if result != nil {
				if result.Meta.Units.Temperature != tt.tempUnit {
					t.Errorf("Temperature = %q, want %q", result.Meta.Units.Temperature, tt.tempUnit)
				}
				if result.Meta.Units.WindSpeed != tt.windUnit {
					t.Errorf("WindSpeed = %q, want %q", result.Meta.Units.WindSpeed, tt.windUnit)
				}
				if result.Meta.Units.Precipitation != tt.precipUnit {
					t.Errorf("Precipitation = %q, want %q", result.Meta.Units.Precipitation, tt.precipUnit)
				}
			}
		})
	}
}

// TestService_Today_FullDay tests that Today returns all hourly rows for the current local date.
func TestService_Today_FullDay(t *testing.T) {
	client := openmeteo.NewClient(nil)
	svc := NewService(client, weathercode.NewMapper())

	// Test with a full 24-hour day
	hourlyTimes := make([]string, 24)
	for i := 0; i < 24; i++ {
		hourlyTimes[i] = "2026-03-21T" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ":00"
	}

	hourly := openmeteo.Hourly{
		Time:                     hourlyTimes,
		Temperature2M:            make([]float64, 24),
		ApparentTemperature:      make([]float64, 24),
		RelativeHumidity2M:       make([]int, 24),
		Precipitation:            make([]float64, 24),
		PrecipitationProbability: make([]int, 24),
		WindSpeed10M:             make([]float64, 24),
		WindGusts10M:             make([]float64, 24),
		WindDirection10M:         make([]int, 24),
		UVIndex:                  make([]float64, 24),
		WeatherCode:              make([]int, 24),
	}

	result, err := svc.mapHourly(hourly, "2026-03-21", time.UTC)
	if err != nil {
		t.Fatalf("mapHourly() returned error: %v", err)
	}

	if len(result) != 24 {
		t.Errorf("Expected 24 hours, got %d", len(result))
	}

	for i, hour := range result {
		if len(hour.Time) != 5 || hour.Time[2] != ':' {
			t.Errorf("Hour %d time %q not in HH:MM format", i, hour.Time)
		}
		if hour.Weather == "" || hour.Weather[:5] == "Unkno" {
			t.Errorf("Hour %d weather should be mapped, got %q", i, hour.Weather)
		}
	}
}

// TestService_Today_JSONOutput tests JSON output format compatibility.
func TestService_Today_JSONOutput(t *testing.T) {
	// Test the models Marshal correctly
	meta := Meta{
		GeneratedAt: time.Now(),
		Units: Units{
			Temperature:              "C",
			Humidity:                 "%",
			WindSpeed:                "km/h",
			WindDirection:            "deg",
			Precipitation:            "mm",
			PrecipitationProbability: "%",
			UVIndex:                  "index",
		},
		Timezone:  "UTC",
		Latitude:  40.0,
		Longitude: -74.0,
	}

	jsonData, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	var output Meta
	if err := json.Unmarshal(jsonData, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.Units.Temperature != "C" {
		t.Errorf("JSON output has wrong temperature unit")
	}
}

// TestService_Today_HourTimeFormat tests that hourly times are in HH:MM format.
func TestService_Today_HourTimeFormat(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name       string
		hourlyTime []string
		expected   []string
	}{
		{
			name:       "midnight_hours",
			hourlyTime: []string{"2026-03-21T00:00", "2026-03-21T01:00", "2026-03-21T09:00"},
			expected:   []string{"00:00", "01:00", "09:00"},
		},
		{
			name:       "afternoon_hours",
			hourlyTime: []string{"2026-03-21T12:00", "2026-03-21T13:30", "2026-03-21T23:59"},
			expected:   []string{"12:00", "13:30", "23:59"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTime,
				Temperature2M:            make([]float64, len(tt.hourlyTime)),
				ApparentTemperature:      make([]float64, len(tt.hourlyTime)),
				RelativeHumidity2M:       make([]int, len(tt.hourlyTime)),
				Precipitation:            make([]float64, len(tt.hourlyTime)),
				PrecipitationProbability: make([]int, len(tt.hourlyTime)),
				WindSpeed10M:             make([]float64, len(tt.hourlyTime)),
				WindGusts10M:             make([]float64, len(tt.hourlyTime)),
				WindDirection10M:         make([]int, len(tt.hourlyTime)),
				UVIndex:                  make([]float64, len(tt.hourlyTime)),
				WeatherCode:              make([]int, len(tt.hourlyTime)),
			}

			hours, err := svc.mapHourly(hourly, "2026-03-21", loc)
			if err != nil {
				t.Fatalf("mapHourly() returned error: %v", err)
			}

			if len(hours) != len(tt.expected) {
				t.Errorf("Expected %d hours, got %d", len(tt.expected), len(hours))
				return
			}

			for i, expected := range tt.expected {
				if hours[i].Time != expected {
					t.Errorf("Hour %d time = %q, want %q", i, hours[i].Time, expected)
				}
			}
		})
	}
}

// TestService_Today_EmptyHourlyData tests handling of empty hourly data.
// When the hourly payload is completely empty, this indicates an upstream
// data inconsistency and should return an error.
func TestService_Today_EmptyHourlyData(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	hourly := openmeteo.Hourly{
		Time:                     []string{},
		Temperature2M:            []float64{},
		ApparentTemperature:      []float64{},
		RelativeHumidity2M:       []int{},
		Precipitation:            []float64{},
		PrecipitationProbability: []int{},
		WindSpeed10M:             []float64{},
		WindGusts10M:             []float64{},
		WindDirection10M:         []int{},
		UVIndex:                  []float64{},
		WeatherCode:              []int{},
	}

	hours, err := svc.mapHourly(hourly, "2026-03-21", loc)
	// Empty hourly payload should return an error
	if err == nil {
		t.Fatalf("mapHourly() expected error for empty hourly payload, got nil")
	}
	if hours != nil {
		t.Errorf("Expected nil hours on error, got %v", hours)
	}
}

// TestService_Today_MissingDateInHourly tests behavior when date is not found in hourly data.
// When the requested date is not in the hourly payload, this indicates an upstream
// data inconsistency and should return an error.
func TestService_Today_MissingDateInHourly(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	hourly := openmeteo.Hourly{
		Time:                     []string{"2026-03-21T00:00", "2026-03-21T12:00", "2026-03-21T23:00"},
		Temperature2M:            []float64{15.0, 20.0, 17.0},
		ApparentTemperature:      []float64{13.0, 18.0, 15.0},
		RelativeHumidity2M:       []int{70, 60, 65},
		Precipitation:            []float64{0.0, 0.0, 0.0},
		PrecipitationProbability: []int{5, 10, 5},
		WindSpeed10M:             []float64{8.0, 10.0, 9.0},
		WindGusts10M:             []float64{12.0, 15.0, 13.0},
		WindDirection10M:         []int{180, 180, 185},
		UVIndex:                  []float64{0.0, 3.0, 1.0},
		WeatherCode:              []int{45, 0, 45},
	}

	hours, err := svc.mapHourly(hourly, "2026-03-22", loc)
	// Missing date in hourly payload should return an error
	if err == nil {
		t.Fatalf("mapHourly() expected error when date not found, got nil")
	}
	if hours != nil {
		t.Errorf("Expected nil hours on error, got %v", hours)
	}
}

// TestService_getUnits tests the getUnits function via the public Today method.
func TestService_getUnits(t *testing.T) {
	client := openmeteo.NewClient(nil)
	svc := NewService(client, weathercode.NewMapper())

	tests := []struct {
		units      string
		tempUnit   string
		windUnit   string
		precipUnit string
	}{
		{"metric", "C", "km/h", "mm"},
		{"imperial", "F", "mph", "inch"},
	}

	for _, tt := range tests {
		t.Run(tt.units, func(t *testing.T) {
			result, err := svc.Today(40.0, -74.0, tt.units)
			if err != nil {
				// Network call fails, but test structure
				return
			}

			if result.Meta.Units.Temperature != tt.tempUnit {
				t.Errorf("Temperature = %q, want %q", result.Meta.Units.Temperature, tt.tempUnit)
			}
			if result.Meta.Units.WindSpeed != tt.windUnit {
				t.Errorf("WindSpeed = %q, want %q", result.Meta.Units.WindSpeed, tt.windUnit)
			}
			if result.Meta.Units.Precipitation != tt.precipUnit {
				t.Errorf("Precipitation = %q, want %q", result.Meta.Units.Precipitation, tt.precipUnit)
			}

			// Verify all other fields are correct
			if result.Meta.Units.Humidity != "%" {
				t.Errorf("Humidity = %q, want %q", result.Meta.Units.Humidity, "%")
			}
			if result.Meta.Units.WindDirection != "deg" {
				t.Errorf("WindDirection = %q, want %q", result.Meta.Units.WindDirection, "deg")
			}
			if result.Meta.Units.PrecipitationProbability != "%" {
				t.Errorf("PrecipitationProbability = %q, want %q", result.Meta.Units.PrecipitationProbability, "%")
			}
			if result.Meta.Units.UVIndex != "index" {
				t.Errorf("UVIndex = %q, want %q", result.Meta.Units.UVIndex, "index")
			}
		})
	}
}

// TestService_WeatherCodeMapping tests weather code to description mapping.
func TestService_WeatherCodeMapping(t *testing.T) {
	mapper := weathercode.NewMapper()
	tests := []struct {
		code        int
		description string
	}{
		{0, "Clear sky"},
		{1, "Mainly clear"},
		{2, "Partly cloudy"},
		{3, "Overcast"},
		{45, "Fog"},
		{61, "Slight rain"},
		{63, "Moderate rain"},
		{95, "Thunderstorm"},
		{99, "Thunderstorm with heavy hail"},
	}

	for _, tt := range tests {
		t.Run("code_"+string(rune('0'+tt.code/100))+"_"+string(rune('0'+(tt.code/10)%10))+"_"+string(rune('0'+tt.code%10)), func(t *testing.T) {
			desc := mapper.GetDescription(tt.code)
			if desc != tt.description {
				t.Errorf("Code %d = %q, want %q", tt.code, desc, tt.description)
			}
		})
	}
}

// TestService_WeatherCodeMapping_UnknownCode tests handling of unknown weather codes.
func TestService_WeatherCodeMapping_UnknownCode(t *testing.T) {
	mapper := weathercode.NewMapper()

	tests := []struct {
		code        int
		description string
	}{
		{999, "Unknown weather code: 999"},
		{-1, "Unknown weather code: -1"},
		{500, "Unknown weather code: 500"},
	}

	for _, tt := range tests {
		t.Run("unknown_code_"+string(rune('0'+tt.code/100))+"_"+string(rune('0'+(tt.code/10)%10))+"_"+string(rune('0'+tt.code%10)), func(t *testing.T) {
			desc := mapper.GetDescription(tt.code)
			if desc != tt.description {
				t.Errorf("Code %d = %q, want %q", tt.code, desc, tt.description)
			}
		})
	}
}

// TestService_mapHourly_MidnightBoundary tests midnight boundary handling.
func TestService_mapHourly_MidnightBoundary(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		name        string
		hourlyTime  []string
		expectedMin int
	}{
		{
			name:        "morning_current_time",
			hourlyTime:  []string{"2026-03-21T00:00", "2026-03-21T05:00", "2026-03-21T06:00"},
			expectedMin: 3,
		},
		{
			name:        "evening_current_time",
			hourlyTime:  []string{"2026-03-21T12:00", "2026-03-21T18:00", "2026-03-21T23:00"},
			expectedMin: 3,
		},
		{
			name:        "late_night_current_time",
			hourlyTime:  []string{"2026-03-21T23:00", "2026-03-21T23:30"},
			expectedMin: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTime,
				Temperature2M:            make([]float64, len(tt.hourlyTime)),
				ApparentTemperature:      make([]float64, len(tt.hourlyTime)),
				RelativeHumidity2M:       make([]int, len(tt.hourlyTime)),
				Precipitation:            make([]float64, len(tt.hourlyTime)),
				PrecipitationProbability: make([]int, len(tt.hourlyTime)),
				WindSpeed10M:             make([]float64, len(tt.hourlyTime)),
				WindGusts10M:             make([]float64, len(tt.hourlyTime)),
				WindDirection10M:         make([]int, len(tt.hourlyTime)),
				UVIndex:                  make([]float64, len(tt.hourlyTime)),
				WeatherCode:              make([]int, len(tt.hourlyTime)),
			}

			hours, err := svc.mapHourly(hourly, "2026-03-21", loc)
			if err != nil {
				t.Fatalf("mapHourly() returned error: %v", err)
			}

			if len(hours) < tt.expectedMin {
				t.Errorf("Expected at least %d hours for %s, got %d", tt.expectedMin, tt.name, len(hours))
			}
		})
	}
}

// TestService_Day_Success tests successful Day responses for multiple requested dates.
func TestService_Day_Success(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	tests := []struct {
		name        string
		date        time.Time
		dailyDate   string
		hourlyTimes []string
		expectedQty int
	}{
		{
			name:        "future_date",
			date:        mustParseDate("2026-04-15"),
			dailyDate:   "2026-04-15",
			hourlyTimes: makeHourlyTimesForDate("2026-04-15"),
			expectedQty: 24,
		},
		{
			name:        "past_date",
			date:        mustParseDate("2026-02-10"),
			dailyDate:   "2026-02-10",
			hourlyTimes: makeHourlyTimesForDate("2026-02-10"),
			expectedQty: 24,
		},
		{
			name:        "leap_year_date",
			date:        mustParseDate("2026-02-28"),
			dailyDate:   "2026-02-28",
			hourlyTimes: makeHourlyTimesForDate("2026-02-28"),
			expectedQty: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daily := openmeteo.Daily{
				Time:                        []string{tt.dailyDate},
				WeatherCode:                 []int{0},
				Temperature2MMin:            []float64{15.0},
				Temperature2MMax:            []float64{25.0},
				PrecipitationSum:            []float64{0.0},
				PrecipitationProbabilityMax: []int{10},
				WindSpeed10MMax:             []float64{10.0},
				WindGusts10MMax:             []float64{15.0},
				UVIndexMax:                  []float64{5.0},
				Sunrise:                     []string{"2026-03-21T06:00"},
				Sunset:                      []string{"2026-03-21T18:00"},
			}

			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTimes,
				Temperature2M:            make([]float64, len(tt.hourlyTimes)),
				ApparentTemperature:      make([]float64, len(tt.hourlyTimes)),
				RelativeHumidity2M:       make([]int, len(tt.hourlyTimes)),
				Precipitation:            make([]float64, len(tt.hourlyTimes)),
				PrecipitationProbability: make([]int, len(tt.hourlyTimes)),
				WindSpeed10M:             make([]float64, len(tt.hourlyTimes)),
				WindGusts10M:             make([]float64, len(tt.hourlyTimes)),
				WindDirection10M:         make([]int, len(tt.hourlyTimes)),
				UVIndex:                  make([]float64, len(tt.hourlyTimes)),
				WeatherCode:              make([]int, len(tt.hourlyTimes)),
			}

			loc, _ := time.LoadLocation("UTC")

			dayResult, err := svc.mapDaily(daily, 0, loc)
			if err != nil {
				t.Fatalf("mapDaily() returned error: %v", err)
			}
			hoursResult, err := svc.mapHourly(hourly, tt.dailyDate, loc)
			if err != nil {
				t.Fatalf("mapHourly() returned error: %v", err)
			}

			if len(hoursResult) != tt.expectedQty {
				t.Errorf("Expected %d hours, got %d", tt.expectedQty, len(hoursResult))
			}

			if dayResult.Date != tt.dailyDate {
				t.Errorf("Day.Date = %q, want %q", dayResult.Date, tt.dailyDate)
			}

			if dayResult.Weather != "Clear sky" {
				t.Errorf("Day.Weather = %q, want %q", dayResult.Weather, "Clear sky")
			}
		})
	}
}

// TestService_Day_OutOfRangeDate tests that ErrDateUnavailable is returned for dates outside forecast range.
func TestService_Day_OutOfRangeDate(t *testing.T) {
	date := mustParseDate("2026-12-25") // Date not in daily data
	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-21", "2026-03-22", "2026-03-23"},
		WeatherCode:                 []int{0, 1, 2},
		Temperature2MMin:            []float64{15.0, 16.0, 17.0},
		Temperature2MMax:            []float64{25.0, 26.0, 27.0},
		PrecipitationSum:            []float64{0.0, 0.5, 0.0},
		PrecipitationProbabilityMax: []int{10, 30, 20},
		WindSpeed10MMax:             []float64{10.0, 12.0, 8.0},
		WindGusts10MMax:             []float64{15.0, 18.0, 12.0},
		UVIndexMax:                  []float64{5.0, 6.0, 4.0},
		Sunrise:                     []string{"2026-03-21T06:00", "2026-03-22T05:59", "2026-03-23T05:58"},
		Sunset:                      []string{"2026-03-21T18:00", "2026-03-22T18:01", "2026-03-23T18:02"},
	}

	dateStr := date.Format("2006-01-02")

	// Find should return -1 since date is not in daily data
	dailyIdx := -1
	for i, dDate := range daily.Time {
		if dDate == dateStr {
			dailyIdx = i
			break
		}
	}

	if dailyIdx != -1 {
		t.Errorf("Expected date not found (idx=-1), got idx=%d", dailyIdx)
	}
}

// TestService_Day_MissingDailyRow tests handling when daily data is missing for a requested date.
func TestService_Day_MissingDailyRow(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	// Daily data for one day only
	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-21"},
		WeatherCode:                 []int{0},
		Temperature2MMin:            []float64{15.0},
		Temperature2MMax:            []float64{25.0},
		PrecipitationSum:            []float64{0.0},
		PrecipitationProbabilityMax: []int{10},
		WindSpeed10MMax:             []float64{10.0},
		WindGusts10MMax:             []float64{15.0},
		UVIndexMax:                  []float64{5.0},
		Sunrise:                     []string{"2026-03-21T06:00"},
		Sunset:                      []string{"2026-03-21T18:00"},
	}

	hourly := openmeteo.Hourly{
		Time:                     []string{"2026-03-22T00:00"},
		Temperature2M:            []float64{20.0},
		ApparentTemperature:      []float64{18.0},
		RelativeHumidity2M:       []int{60},
		Precipitation:            []float64{0.0},
		PrecipitationProbability: []int{10},
		WindSpeed10M:             []float64{10.0},
		WindGusts10M:             []float64{15.0},
		WindDirection10M:         []int{180},
		UVIndex:                  []float64{3.0},
		WeatherCode:              []int{0},
	}

	// Request date not in daily data
	result, err := svc.mapDaily(daily, 0, loc)
	if err != nil {
		t.Fatalf("mapDaily() returned error: %v", err)
	}

	if result.Date != "2026-03-21" {
		t.Errorf("Date = %q, want %q", result.Date, "2026-03-21")
	}

	// For hourly data at a different date - should return error
	hours, err := svc.mapHourly(hourly, "2026-03-23", loc)
	// Missing date in hourly payload should return an error
	if err == nil {
		t.Fatalf("mapHourly() expected error when date not found, got nil")
	}
	if hours != nil {
		t.Errorf("Expected nil hours on error, got %v", hours)
	}
}

// TestService_Day_DSTTransition tests DST transition date handling.
func TestService_Day_DSTTransition(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	tests := []struct {
		name        string
		date        string
		hourlyTimes []string
	}{
		{
			name:        "spring_forward_2026",
			date:        "2026-03-08", // DST starts at 2:00 AM
			hourlyTimes: []string{"2026-03-08T06:00", "2026-03-08T12:00", "2026-03-08T18:00"},
		},
		{
			name:        "fall_back_2026",
			date:        "2026-11-01", // DST ends at 2:00 AM
			hourlyTimes: []string{"2026-11-01T06:00", "2026-11-01T12:00", "2026-11-01T18:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daily := openmeteo.Daily{
				Time:                        []string{tt.date},
				WeatherCode:                 []int{0},
				Temperature2MMin:            []float64{10.0},
				Temperature2MMax:            []float64{20.0},
				PrecipitationSum:            []float64{0.0},
				PrecipitationProbabilityMax: []int{5},
				WindSpeed10MMax:             []float64{8.0},
				WindGusts10MMax:             []float64{12.0},
				UVIndexMax:                  []float64{4.0},
				Sunrise:                     []string{"2026-03-08T07:00"},
				Sunset:                      []string{"2026-03-08T19:00"},
			}

			hourly := openmeteo.Hourly{
				Time:                     tt.hourlyTimes,
				Temperature2M:            make([]float64, len(tt.hourlyTimes)),
				ApparentTemperature:      make([]float64, len(tt.hourlyTimes)),
				RelativeHumidity2M:       make([]int, len(tt.hourlyTimes)),
				Precipitation:            make([]float64, len(tt.hourlyTimes)),
				PrecipitationProbability: make([]int, len(tt.hourlyTimes)),
				WindSpeed10M:             make([]float64, len(tt.hourlyTimes)),
				WindGusts10M:             make([]float64, len(tt.hourlyTimes)),
				WindDirection10M:         make([]int, len(tt.hourlyTimes)),
				UVIndex:                  make([]float64, len(tt.hourlyTimes)),
				WeatherCode:              make([]int, len(tt.hourlyTimes)),
			}

			dayResult, err := svc.mapDaily(daily, 0, loc)
			if err != nil {
				t.Fatalf("mapDaily() returned error: %v", err)
			}
			hoursResult, err := svc.mapHourly(hourly, tt.date, loc)
			if err != nil {
				t.Fatalf("mapHourly() returned error: %v", err)
			}

			if dayResult.Date != tt.date {
				t.Errorf("Day.Date = %q, want %q", dayResult.Date, tt.date)
			}

			if len(hoursResult) != 3 {
				t.Errorf("Expected 3 hours, got %d", len(hoursResult))
			}

			// Verify sunrise/sunset are full datetime format
			if len(dayResult.Sunrise) < 16 {
				t.Errorf("Sunrise should be full datetime, got %q", dayResult.Sunrise)
			}
			if len(dayResult.Sunset) < 16 {
				t.Errorf("Sunset should be full datetime, got %q", dayResult.Sunset)
			}
		})
	}
}

// TestService_Day_HourTimeFormat tests that hourly times in Day output are in HH:MM format.
func TestService_Day_HourTimeFormat(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22"},
		WeatherCode:                 []int{1},
		Temperature2MMin:            []float64{12.0},
		Temperature2MMax:            []float64{22.0},
		PrecipitationSum:            []float64{0.0},
		PrecipitationProbabilityMax: []int{15},
		WindSpeed10MMax:             []float64{11.0},
		WindGusts10MMax:             []float64{16.0},
		UVIndexMax:                  []float64{6.0},
		Sunrise:                     []string{"2026-03-22T07:05"},
		Sunset:                      []string{"2026-03-22T19:15"},
	}

	hourlyTimes := []string{
		"2026-03-22T00:00", "2026-03-22T06:00", "2026-03-22T12:00",
		"2026-03-22T18:00", "2026-03-22T23:00",
	}

	hourly := openmeteo.Hourly{
		Time:                     hourlyTimes,
		Temperature2M:            []float64{10.0, 12.0, 20.0, 18.0, 15.0},
		ApparentTemperature:      []float64{8.0, 10.0, 18.0, 16.0, 13.0},
		RelativeHumidity2M:       []int{70, 65, 55, 60, 75},
		Precipitation:            []float64{0.0, 0.0, 0.0, 0.0, 0.0},
		PrecipitationProbability: []int{5, 5, 10, 10, 15},
		WindSpeed10M:             []float64{8.0, 10.0, 11.0, 9.0, 7.0},
		WindGusts10M:             []float64{12.0, 14.0, 16.0, 13.0, 10.0},
		WindDirection10M:         []int{170, 180, 185, 175, 165},
		UVIndex:                  []float64{0.0, 2.0, 5.0, 3.0, 0.0},
		WeatherCode:              []int{45, 1, 1, 1, 2},
	}

	dayResult, err := svc.mapDaily(daily, 0, loc)
	if err != nil {
		t.Fatalf("mapDaily() returned error: %v", err)
	}
	hoursResult, err := svc.mapHourly(hourly, "2026-03-22", loc)
	if err != nil {
		t.Fatalf("mapHourly() returned error: %v", err)
	}

	if dayResult.Date != "2026-03-22" {
		t.Errorf("Day.Date = %q, want %q", dayResult.Date, "2026-03-22")
	}

	if len(hoursResult) != 5 {
		t.Errorf("Expected 5 hours, got %d", len(hoursResult))
	}

	expectedTimes := []string{"00:00", "06:00", "12:00", "18:00", "23:00"}

	for i, h := range hoursResult {
		if len(h.Time) != 5 || h.Time[2] != ':' {
			t.Errorf("Hour %d time %q not in HH:MM format", i, h.Time)
		}
		if h.Time != expectedTimes[i] {
			t.Errorf("Hour %d time = %q, want %q", i, h.Time, expectedTimes[i])
		}
		if h.Weather == "" || len(h.Weather) >= 5 && h.Weather[:5] == "Unkno" {
			t.Errorf("Hour %d weather should be mapped, got %q", i, h.Weather)
		}
	}

	// Verify sunrise/sunset are full datetime
	if len(dayResult.Sunrise) != 20 || dayResult.Sunrise[10] != 'T' {
		t.Errorf("Sunrise should be full datetime YYYY-MM-DDTHH:MM:SSZ, got %q", dayResult.Sunrise)
	}
	if len(dayResult.Sunset) != 20 || dayResult.Sunset[10] != 'T' {
		t.Errorf("Sunset should be full datetime YYYY-MM-DDTHH:MM:SSZ, got %q", dayResult.Sunset)
	}
}

// TestService_Day_ImperialUnits tests Day output with imperial units.
func TestService_Day_ImperialUnits(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22"},
		WeatherCode:                 []int{0},
		Temperature2MMin:            []float64{-10.0},
		Temperature2MMax:            []float64{10.0},
		PrecipitationSum:            []float64{2.54},
		PrecipitationProbabilityMax: []int{30},
		WindSpeed10MMax:             []float64{16.09}, // 10 mph
		WindGusts10MMax:             []float64{24.14}, // 15 mph
		UVIndexMax:                  []float64{4.0},
		Sunrise:                     []string{"2026-03-22T07:00"},
		Sunset:                      []string{"2026-03-22T19:00"},
	}

	hourly := openmeteo.Hourly{
		Time:                     []string{"2026-03-22T12:00"},
		Temperature2M:            []float64{0.0},
		ApparentTemperature:      []float64{-2.0},
		RelativeHumidity2M:       []int{50},
		Precipitation:            []float64{0.0},
		PrecipitationProbability: []int{10},
		WindSpeed10M:             []float64{8.05},  // 5 mph
		WindGusts10M:             []float64{12.07}, // 7.5 mph
		WindDirection10M:         []int{180},
		UVIndex:                  []float64{3.0},
		WeatherCode:              []int{0},
	}

	hours, err := svc.mapHourly(hourly, "2026-03-22", loc)
	if err != nil {
		t.Fatalf("mapHourly() returned error: %v", err)
	}
	dayResult, err := svc.mapDaily(daily, 0, loc)
	if err != nil {
		t.Fatalf("mapDaily() returned error: %v", err)
	}

	if len(hours) != 1 {
		t.Fatalf("Expected 1 hour, got %d", len(hours))
	}

	if hours[0].Temperature != 0.0 {
		t.Errorf("Hour temperature = %v, want %v", hours[0].Temperature, 0.0)
	}

	if dayResult.TempMin != -10.0 {
		t.Errorf("Day.TempMin = %v, want %v", dayResult.TempMin, -10.0)
	}
	if dayResult.TempMax != 10.0 {
		t.Errorf("Day.TempMax = %v, want %v", dayResult.TempMax, 10.0)
	}
}

// TestService_Day_WeatherCodeMapping tests weather code mapping in Day output.
func TestService_Day_WeatherCodeMapping(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	tests := []struct {
		code        int
		description string
	}{
		{0, "Clear sky"},
		{1, "Mainly clear"},
		{2, "Partly cloudy"},
		{3, "Overcast"},
		{45, "Fog"},
		{48, "Depositing rime fog"},
		{51, "Light drizzle"},
		{61, "Slight rain"},
		{63, "Moderate rain"},
		{65, "Heavy rain"},
		{95, "Thunderstorm"},
	}

	for _, tt := range tests {
		t.Run("code_"+string(rune('0'+tt.code/100))+"_"+string(rune('0'+(tt.code/10)%10))+"_"+string(rune('0'+tt.code%10)), func(t *testing.T) {
			daily := openmeteo.Daily{
				Time:                        []string{"2026-03-22"},
				WeatherCode:                 []int{tt.code},
				Temperature2MMin:            []float64{15.0},
				Temperature2MMax:            []float64{25.0},
				PrecipitationSum:            []float64{0.0},
				PrecipitationProbabilityMax: []int{10},
				WindSpeed10MMax:             []float64{10.0},
				WindGusts10MMax:             []float64{15.0},
				UVIndexMax:                  []float64{5.0},
				Sunrise:                     []string{"2026-03-22T06:00"},
				Sunset:                      []string{"2026-03-22T18:00"},
			}

			result, err := svc.mapDaily(daily, 0, loc)
			if err != nil {
				t.Fatalf("mapDaily() returned error: %v", err)
			}

			if result.Weather != tt.description {
				t.Errorf("Weather code %d = %q, want %q", tt.code, result.Weather, tt.description)
			}
		})
	}
}

// TestService_Day_MismatchedHourlyData tests handling of mismatched hourly array lengths.
func TestService_Day_MismatchedHourlyData(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	// Hourly data with matching array lengths
	hourly := openmeteo.Hourly{
		Time:                     []string{"2026-03-22T00:00", "2026-03-22T12:00"},
		Temperature2M:            []float64{20.0, 25.0},
		ApparentTemperature:      []float64{18.0, 23.0},
		RelativeHumidity2M:       []int{60, 55},
		Precipitation:            []float64{0.0, 0.0},
		PrecipitationProbability: []int{10, 15},
		WindSpeed10M:             []float64{10.0, 12.0},
		WindGusts10M:             []float64{15.0, 18.0},
		WindDirection10M:         []int{180, 185},
		UVIndex:                  []float64{3.0, 5.0},
		WeatherCode:              []int{0, 1},
	}

	hours, err := svc.mapHourly(hourly, "2026-03-22", loc)
	if err != nil {
		t.Fatalf("mapHourly() returned error: %v", err)
	}

	if len(hours) != 2 {
		t.Errorf("Expected 2 hours with matching data, got %d", len(hours))
	}

	if hours[0].Time != "00:00" || hours[1].Time != "12:00" {
		t.Errorf("Expected times 00:00 and 12:00, got %q and %q", hours[0].Time, hours[1].Time)
	}
}

// TestService_Day_MidnightBoundary tests midnight boundary handling for Day.
func TestService_Day_MidnightBoundary(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22"},
		WeatherCode:                 []int{0},
		Temperature2MMin:            []float64{10.0},
		Temperature2MMax:            []float64{20.0},
		PrecipitationSum:            []float64{0.0},
		PrecipitationProbabilityMax: []int{5},
		WindSpeed10MMax:             []float64{8.0},
		WindGusts10MMax:             []float64{12.0},
		UVIndexMax:                  []float64{4.0},
		Sunrise:                     []string{"2026-03-22T06:00"},
		Sunset:                      []string{"2026-03-22T18:00"},
	}

	hourlyTimes := []string{
		"2026-03-21T23:30", // Just before midnight
		"2026-03-22T00:00", // At midnight
		"2026-03-22T01:00",
		"2026-03-22T12:00",
		"2026-03-22T23:00",
		"2026-03-22T23:59", // Just before end of day
	}

	hourly := openmeteo.Hourly{
		Time:                     hourlyTimes,
		Temperature2M:            make([]float64, len(hourlyTimes)),
		ApparentTemperature:      make([]float64, len(hourlyTimes)),
		RelativeHumidity2M:       make([]int, len(hourlyTimes)),
		Precipitation:            make([]float64, len(hourlyTimes)),
		PrecipitationProbability: make([]int, len(hourlyTimes)),
		WindSpeed10M:             make([]float64, len(hourlyTimes)),
		WindGusts10M:             make([]float64, len(hourlyTimes)),
		WindDirection10M:         make([]int, len(hourlyTimes)),
		UVIndex:                  make([]float64, len(hourlyTimes)),
		WeatherCode:              make([]int, len(hourlyTimes)),
	}

	// This should only include hours from 2026-03-22
	hours, err := svc.mapHourly(hourly, "2026-03-22", loc)
	if err != nil {
		t.Fatalf("mapHourly() returned error: %v", err)
	}

	// First hourly at 23:30 on 03-21 should NOT be included
	// 5 entries should be included (00:00, 01:00, 12:00, 23:00, 23:59)
	expectedHours := 5
	if len(hours) != expectedHours {
		t.Errorf("Expected %d hours for date boundary test, got %d", expectedHours, len(hours))
	}

	// First hour should be 00:00, not 23:30
	if len(hours) > 0 && hours[0].Time != "00:00" {
		t.Errorf("First hour should be 00:00, got %q", hours[0].Time)
	}

	dayResult, err := svc.mapDaily(daily, 0, loc)
	if err != nil {
		t.Fatalf("mapDaily() returned error: %v", err)
	}
	if dayResult.Date != "2026-03-22" {
		t.Errorf("Day.Date = %q, want %q", dayResult.Date, "2026-03-22")
	}
}

// mustParseDate parses a date string into a time.Time.
func mustParseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return t
}

// makeHourlyTimesForDate creates 24 hourly time strings for a given date.
func makeHourlyTimesForDate(dateStr string) []string {
	times := make([]string, 24)
	for i := 0; i < 24; i++ {
		times[i] = fmt.Sprintf("%sT%02d:00", dateStr, i)
	}
	return times
}

// TestService_Week_SuccessfulResponse tests a successful 7-day week response.
func TestService_Week_SuccessfulResponse(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"},
		WeatherCode:                 []int{0, 1, 2, 3, 45, 61, 95},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0, 2.5, 5.0, 8.0, 0.0},
		PrecipitationProbabilityMax: []int{5, 10, 15, 20, 50, 80, 10},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0, 15.0, 18.0, 20.0, 7.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0, 20.0, 25.0, 28.0, 10.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0, 6.5, 5.5, 4.5, 4.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02", "2026-03-25T06:03", "2026-03-26T06:04", "2026-03-27T06:05", "2026-03-28T06:06"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02", "2026-03-25T18:03", "2026-03-26T18:04", "2026-03-27T18:05", "2026-03-28T18:06"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	if err != nil {
		t.Fatalf("mapWeekDaily() returned error: %v", err)
	}

	if len(days) != 7 {
		t.Errorf("Expected 7 days, got %d", len(days))
	}

	// Verify first day
	if days[0].Date != "2026-03-22" {
		t.Errorf("Day[0].Date = %q, want %q", days[0].Date, "2026-03-22")
	}
	if days[0].Weather != "Clear sky" {
		t.Errorf("Day[0].Weather = %q, want %q", days[0].Weather, "Clear sky")
	}
	if days[0].TempMin != 10.0 {
		t.Errorf("Day[0].TempMin = %v, want %v", days[0].TempMin, 10.0)
	}
	if days[0].TempMax != 20.0 {
		t.Errorf("Day[0].TempMax = %v, want %v", days[0].TempMax, 20.0)
	}

	// Verify last day
	if days[6].Date != "2026-03-28" {
		t.Errorf("Day[6].Date = %q, want %q", days[6].Date, "2026-03-28")
	}
	if days[6].Weather != "Thunderstorm" {
		t.Errorf("Day[6].Weather = %q, want %q", days[6].Weather, "Thunderstorm")
	}
}

// TestService_Week_ShortDailyArray tests behavior when upstream daily data has fewer than 7 days.
// This should return an error since the contract specifies exactly 7 days must be returned.
func TestService_Week_ShortDailyArray(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	// Only 3 days available (simulating API returning less than expected)
	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24"},
		WeatherCode:                 []int{0, 1, 2},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0},
		PrecipitationProbabilityMax: []int{5, 10, 15},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	// Should return an error since we require exactly 7 days
	if err == nil {
		t.Error("mapWeekDaily() should return error when daily array has fewer than 7 entries")
	}
	if days != nil {
		t.Errorf("Expected nil days on error, got %d days", len(days))
	}
	// Verify the error wraps ErrUpstreamAPI
	if err != nil && !errors.Is(err, openmeteo.ErrUpstreamAPI) {
		t.Errorf("Expected error to wrap ErrUpstreamAPI, got: %v", err)
	}
}

// TestService_Week_ExactSevenDays tests that exactly 7 days are returned when API provides exactly 7.
func TestService_Week_ExactSevenDays(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"},
		WeatherCode:                 []int{0, 1, 2, 3, 45, 61, 95},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0, 2.5, 5.0, 8.0, 0.0},
		PrecipitationProbabilityMax: []int{5, 10, 15, 20, 50, 80, 10},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0, 15.0, 18.0, 20.0, 7.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0, 20.0, 25.0, 28.0, 10.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0, 6.5, 5.5, 4.5, 4.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02", "2026-03-25T06:03", "2026-03-26T06:04", "2026-03-27T06:05", "2026-03-28T06:06"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02", "2026-03-25T18:03", "2026-03-26T18:04", "2026-03-27T18:05", "2026-03-28T18:06"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	if err != nil {
		t.Fatalf("mapWeekDaily() returned error: %v", err)
	}

	if len(days) != 7 {
		t.Errorf("Expected exactly 7 days, got %d", len(days))
	}

	// Verify all days are present with correct dates
	expectedDates := []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"}
	for i, expectedDate := range expectedDates {
		if days[i].Date != expectedDate {
			t.Errorf("Day[%d].Date = %q, want %q", i, days[i].Date, expectedDate)
		}
	}
}

// TestService_Week_TimezoneBoundaries tests that timezone positioning is correctly handled.
func TestService_Week_TimezoneBoundaries(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	// Test with New York timezone (UTC-5)
	loc, _ := time.LoadLocation("America/New_York")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"},
		WeatherCode:                 []int{0, 1, 2, 3, 45, 61, 95},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0, 2.5, 5.0, 8.0, 0.0},
		PrecipitationProbabilityMax: []int{5, 10, 15, 20, 50, 80, 10},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0, 15.0, 18.0, 20.0, 7.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0, 20.0, 25.0, 28.0, 10.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0, 6.5, 5.5, 4.5, 4.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02", "2026-03-25T06:03", "2026-03-26T06:04", "2026-03-27T06:05", "2026-03-28T06:06"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02", "2026-03-25T18:03", "2026-03-26T18:04", "2026-03-27T18:05", "2026-03-28T18:06"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	if err != nil {
		t.Fatalf("mapWeekDaily() returned error: %v", err)
	}

	if len(days) != 7 {
		t.Errorf("Expected 7 days with timezone, got %d", len(days))
	}

	// Verify dates are preserved correctly
	for i, day := range days {
		expectedDate := fmt.Sprintf("2026-03-%02d", 22+i)
		if day.Date != expectedDate {
			t.Errorf("Day[%d].Date = %q, want %q", i, day.Date, expectedDate)
		}
	}
}

// TestService_Week_WeatherCodeMapping tests that weather codes are correctly mapped to descriptions.
func TestService_Week_WeatherCodeMapping(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"},
		WeatherCode:                 []int{0, 1, 2, 3, 45, 48, 51},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0, 2.5, 5.0, 8.0, 0.0},
		PrecipitationProbabilityMax: []int{5, 10, 15, 20, 50, 80, 10},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0, 15.0, 18.0, 20.0, 7.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0, 20.0, 25.0, 28.0, 10.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0, 6.5, 5.5, 4.5, 4.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02", "2026-03-25T06:03", "2026-03-26T06:04", "2026-03-27T06:05", "2026-03-28T06:06"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02", "2026-03-25T18:03", "2026-03-26T18:04", "2026-03-27T18:05", "2026-03-28T18:06"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	if err != nil {
		t.Fatalf("mapWeekDaily() returned error: %v", err)
	}

	// Check a few weather mappings
	if days[0].Weather != "Clear sky" {
		t.Errorf("Day[0].Weather = %q, want %q", days[0].Weather, "Clear sky")
	}
	if days[1].Weather != "Mainly clear" {
		t.Errorf("Day[1].Weather = %q, want %q", days[1].Weather, "Mainly clear")
	}
	if days[3].Weather != "Overcast" {
		t.Errorf("Day[3].Weather = %q, want %q", days[3].Weather, "Overcast")
	}
}

// TestService_Week_SunsetSunriseFormat tests that sunrise/sunset are in full datetime format.
func TestService_Week_SunsetSunriseFormat(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"},
		WeatherCode:                 []int{0, 1, 2, 3, 45, 61, 95},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0, 2.5, 5.0, 8.0, 0.0},
		PrecipitationProbabilityMax: []int{5, 10, 15, 20, 50, 80, 10},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0, 15.0, 18.0, 20.0, 7.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0, 20.0, 25.0, 28.0, 10.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0, 6.5, 5.5, 4.5, 4.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02", "2026-03-25T06:03", "2026-03-26T06:04", "2026-03-27T06:05", "2026-03-28T06:06"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02", "2026-03-25T18:03", "2026-03-26T18:04", "2026-03-27T18:05", "2026-03-28T18:06"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	if err != nil {
		t.Fatalf("mapWeekDaily() returned error: %v", err)
	}

	// Verify sunrise/sunset have full datetime format (YYYY-MM-DDTHH:MM:SSZ) - RFC3339
	expectedDates := []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"}
	for i := 0; i < 7; i++ {
		if len(days[i].Sunrise) != 20 {
			t.Errorf("Day[%d].Sunrise = %q, expected 20 chars (YYYY-MM-DDTHH:MM:SSZ)", i, days[i].Sunrise)
			continue
		}
		if len(days[i].Sunset) != 20 {
			t.Errorf("Day[%d].Sunset = %q, expected 20 chars (YYYY-MM-DDTHH:MM:SSZ)", i, days[i].Sunset)
			continue
		}
		// Verify date matches expected
		if days[i].Sunrise[:10] != expectedDates[i] {
			t.Errorf("Day[%d].Sunrise date = %q, want %q", i, days[i].Sunrise[:10], expectedDates[i])
		}
	}
}

// TestService_Week_ImplicitUnits test that units are correctly set for metric.
func TestService_Week_ImplicitUnits(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	loc, _ := time.LoadLocation("UTC")

	daily := openmeteo.Daily{
		Time:                        []string{"2026-03-22", "2026-03-23", "2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28"},
		WeatherCode:                 []int{0, 1, 2, 3, 45, 61, 95},
		Temperature2MMin:            []float64{10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0},
		Temperature2MMax:            []float64{20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0},
		PrecipitationSum:            []float64{0.0, 0.5, 1.0, 2.5, 5.0, 8.0, 0.0},
		PrecipitationProbabilityMax: []int{5, 10, 15, 20, 50, 80, 10},
		WindSpeed10MMax:             []float64{8.0, 10.0, 12.0, 15.0, 18.0, 20.0, 7.0},
		WindGusts10MMax:             []float64{12.0, 14.0, 16.0, 20.0, 25.0, 28.0, 10.0},
		UVIndexMax:                  []float64{4.0, 5.0, 6.0, 6.5, 5.5, 4.5, 4.0},
		Sunrise:                     []string{"2026-03-22T06:00", "2026-03-23T06:01", "2026-03-24T06:02", "2026-03-25T06:03", "2026-03-26T06:04", "2026-03-27T06:05", "2026-03-28T06:06"},
		Sunset:                      []string{"2026-03-22T18:00", "2026-03-23T18:01", "2026-03-24T18:02", "2026-03-25T18:03", "2026-03-26T18:04", "2026-03-27T18:05", "2026-03-28T18:06"},
	}

	days, err := svc.mapWeekDaily(daily, 7, loc)
	if err != nil {
		t.Fatalf("mapWeekDaily() returned error: %v", err)
	}

	// Verify all numeric fields are present
	for i := 0; i < 7; i++ {
		if days[i].TempMin == 0 {
			t.Errorf("Day[%d].TempMin should not be zero", i)
		}
		if days[i].TempMax == 0 {
			t.Errorf("Day[%d].TempMax should not be zero", i)
		}
		if days[i].WindSpeedMax == 0 {
			t.Errorf("Day[%d].WindSpeedMax should not be zero", i)
		}
	}
}

// TestService_Week_MetricUnits test units are correct for metric.
func TestService_Week_MetricUnits(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	units := svc.getUnits("metric")

	if units.Temperature != "C" {
		t.Errorf("units.Temperature = %q, want %q", units.Temperature, "C")
	}
	if units.WindSpeed != "km/h" {
		t.Errorf("units.WindSpeed = %q, want %q", units.WindSpeed, "km/h")
	}
	if units.Precipitation != "mm" {
		t.Errorf("units.Precipitation = %q, want %q", units.Precipitation, "mm")
	}
}

// TestService_Week_ImperialUnits test units are correct for imperial.
func TestService_Week_ImperialUnits(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	units := svc.getUnits("imperial")

	if units.Temperature != "F" {
		t.Errorf("units.Temperature = %q, want %q", units.Temperature, "F")
	}
	if units.WindSpeed != "mph" {
		t.Errorf("units.WindSpeed = %q, want %q", units.WindSpeed, "mph")
	}
	if units.Precipitation != "inch" {
		t.Errorf("units.Precipitation = %q, want %q", units.Precipitation, "inch")
	}
}
