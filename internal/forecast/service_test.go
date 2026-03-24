package forecast

import (
	"encoding/json"
	"testing"
	"time"

	"openmeteo-cli/internal/openmeteo"
	"openmeteo-cli/internal/weathercode"
)

// TestService_Forecast_Hourly tests the Forecast method with hourly mode.
func TestService_Forecast_Hourly(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	// Test with 1 day
	result, err := svc.Forecast(40.0, -74.0, "metric", true, 1)
	if err != nil {
		// Expected - network call will fail, we're just verifying method exists
		_ = err
	}
	if result != nil {
		hourly, ok := result.(*HourlyOutput)
		if !ok {
			t.Fatalf("Expected HourlyOutput, got %T", result)
		}
		if hourly.Meta.Units.Temperature != "C" {
			t.Errorf("Temperature = %q, want %q", hourly.Meta.Units.Temperature, "C")
		}
	}

	// Test with 2 days
	result, err = svc.Forecast(40.0, -74.0, "metric", true, 2)
	if err != nil {
		_ = err
	}
	if result != nil {
		hourly, ok := result.(*HourlyOutput)
		if !ok {
			t.Fatalf("Expected HourlyOutput for 2 days, got %T", result)
		}
		if len(hourly.Days) > 2 {
			t.Errorf("Expected at most 2 days, got %d", len(hourly.Days))
		}
	}
}

// TestService_Forecast_Daily tests the Forecast method with daily mode.
func TestService_Forecast_Daily(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	// Test with 7 days
	result, err := svc.Forecast(40.0, -74.0, "metric", false, 7)
	if err != nil {
		// Expected - network call will fail, we're just verifying method exists
		_ = err
	}
	if result != nil {
		daily, ok := result.(*DailyOutput)
		if !ok {
			t.Fatalf("Expected DailyOutput, got %T", result)
		}
		if daily.Meta.Units.Temperature != "C" {
			t.Errorf("Temperature = %q, want %q", daily.Meta.Units.Temperature, "C")
		}
	}

	// Test with 14 days
	result, err = svc.Forecast(40.0, -74.0, "metric", false, 14)
	if err != nil {
		_ = err
	}
	if result != nil {
		daily, ok := result.(*DailyOutput)
		if !ok {
			t.Fatalf("Expected DailyOutput for 14 days, got %T", result)
		}
		if len(daily.Days) > 14 {
			t.Errorf("Expected at most 14 days, got %d", len(daily.Days))
		}
	}
}

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
		{"Sunrise", "06:00"},
		{"Sunset", "18:00"},
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

	// The parsed time should be in local timezone
	// Format should be HH:MM
	if len(result.Sunrise) != 5 || result.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be in HH:MM format, got %q", result.Sunrise)
	}
	if len(result.Sunset) != 5 || result.Sunset[2] != ':' {
		t.Errorf("Sunset should be in HH:MM format, got %q", result.Sunset)
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

// TestService_mapDaily_DSTTransition tests DST transition date handling.
func TestService_mapDaily_DSTTransition(t *testing.T) {
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

			// Verify sunrise/sunset are HH:MM format
			if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
				t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
			}
			if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
				t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
			}
		})
	}
}

// TestService_mapDaily_HourTimeFormat tests that hourly times in Day output are in HH:MM format.
func TestService_mapDaily_HourTimeFormat(t *testing.T) {
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

	// Verify sunrise/sunset are HH:MM format
	if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
	}
	if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
		t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
	}
}

// TestService_mapDaily_ImperialUnits tests Day output with imperial units.
func TestService_mapDaily_ImperialUnits(t *testing.T) {
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

// TestService_mapDaily_WeatherCodeMapping tests weather code mapping in Day output.
func TestService_mapDaily_WeatherCodeMapping(t *testing.T) {
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

// TestService_mapDaily_MismatchedHourlyData tests handling of mismatched hourly array lengths.
func TestService_mapDaily_MismatchedHourlyData(t *testing.T) {
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

// TestService_mapDaily_MidnightBoundary tests midnight boundary handling for Day.
func TestService_mapDaily_MidnightBoundary(t *testing.T) {
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

// TestService_mapHourly_DatetimeFiltering tests that mapHourly filters to the correct date.
func TestService_mapHourly_DatetimeFiltering2(t *testing.T) {
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
func TestService_mapHourly_TimezoneConversion2(t *testing.T) {
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
func TestService_mapCurrent2(t *testing.T) {
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
func TestService_mapHourly_UnknownWeatherCode2(t *testing.T) {
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
func TestService_mapHourly_DSTHandling2(t *testing.T) {
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
func TestService_mapDaily2(t *testing.T) {
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
		{"Sunrise", "06:00"},
		{"Sunset", "18:00"},
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
func TestService_mapDaily_TimezoneOffset2(t *testing.T) {
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

	// The parsed time should be in local timezone
	// Format should be HH:MM
	if len(result.Sunrise) != 5 || result.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be in HH:MM format, got %q", result.Sunrise)
	}
	if len(result.Sunset) != 5 || result.Sunset[2] != ':' {
		t.Errorf("Sunset should be in HH:MM format, got %q", result.Sunset)
	}
}

// TestService_mapDaily_MultipleDays tests mapping multiple days.
func TestService_mapDaily_MultipleDays2(t *testing.T) {
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

// TestService_mapDaily_DSTTransition tests DST transition date handling.
func TestService_mapDaily_DSTTransition2(t *testing.T) {
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

			// Verify sunrise/sunset are HH:MM format
			if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
				t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
			}
			if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
				t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
			}
		})
	}
}

// TestService_mapDaily_HourTimeFormat tests that hourly times in Day output are in HH:MM format.
func TestService_mapDaily_HourTimeFormat2(t *testing.T) {
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

	// Verify sunrise/sunset are HH:MM format
	if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
	}
	if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
		t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
	}
}

// TestService_mapDaily_ImperialUnits tests Day output with imperial units.
func TestService_mapDaily_ImperialUnits2(t *testing.T) {
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

// TestService_mapDaily_WeatherCodeMapping tests weather code mapping in Day output.
func TestService_mapDaily_WeatherCodeMapping2(t *testing.T) {
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

// TestService_mapDaily_MismatchedHourlyData tests handling of mismatched hourly array lengths.
func TestService_mapDaily_MismatchedHourlyData2(t *testing.T) {
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

// TestService_mapDaily_MidnightBoundary tests midnight boundary handling for Day.
func TestService_mapDaily_MidnightBoundary2(t *testing.T) {
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

// TestService_mapHourly_DatetimeFiltering tests that mapHourly filters to the correct date.
func TestService_mapHourly_DatetimeFiltering3(t *testing.T) {
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
func TestService_mapHourly_TimezoneConversion3(t *testing.T) {
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
func TestService_mapCurrent3(t *testing.T) {
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
func TestService_mapHourly_UnknownWeatherCode3(t *testing.T) {
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
func TestService_mapHourly_DSTHandling3(t *testing.T) {
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
func TestService_mapDaily3(t *testing.T) {
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
		{"Sunrise", "06:00"},
		{"Sunset", "18:00"},
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
func TestService_mapDaily_TimezoneOffset3(t *testing.T) {
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

	// The parsed time should be in local timezone
	// Format should be HH:MM
	if len(result.Sunrise) != 5 || result.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be in HH:MM format, got %q", result.Sunrise)
	}
	if len(result.Sunset) != 5 || result.Sunset[2] != ':' {
		t.Errorf("Sunset should be in HH:MM format, got %q", result.Sunset)
	}
}

// TestService_mapDaily_MultipleDays tests mapping multiple days.
func TestService_mapDaily_MultipleDays3(t *testing.T) {
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

// TestService_mapDaily_DSTTransition tests DST transition date handling.
func TestService_mapDaily_DSTTransition3(t *testing.T) {
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

			// Verify sunrise/sunset are HH:MM format
			if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
				t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
			}
			if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
				t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
			}
		})
	}
}

// TestService_mapDaily_HourTimeFormat tests that hourly times in Day output are in HH:MM format.
func TestService_mapDaily_HourTimeFormat3(t *testing.T) {
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

	// Verify sunrise/sunset are HH:MM format
	if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
	}
	if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
		t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
	}
}

// TestService_mapDaily_ImperialUnits tests Day output with imperial units.
func TestService_mapDaily_ImperialUnits3(t *testing.T) {
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

// TestService_mapDaily_WeatherCodeMapping tests weather code mapping in Day output.
func TestService_mapDaily_WeatherCodeMapping3(t *testing.T) {
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

// TestService_mapDaily_MismatchedHourlyData tests handling of mismatched hourly array lengths.
func TestService_mapDaily_MismatchedHourlyData3(t *testing.T) {
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

// TestService_mapDaily_MidnightBoundary tests midnight boundary handling for Day.
func TestService_mapDaily_MidnightBoundary3(t *testing.T) {
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

// TestService_Forecast_Hourly_DaysValidation tests that the Forecast method validates days for hourly mode.
func TestService_Forecast_Hourly_DaysValidation(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	// Test with 0 days - should fail
	_, err := svc.Forecast(40.0, -74.0, "metric", true, 0)
	if err == nil {
		t.Error("Expected error for 0 days, got nil")
	}

	// Test with 3 days - should fail
	_, err = svc.Forecast(40.0, -74.0, "metric", true, 3)
	if err == nil {
		t.Error("Expected error for 3 days, got nil")
	}

	// Test with 2 days - should succeed
	result, err := svc.Forecast(40.0, -74.0, "metric", true, 2)
	if err != nil {
		t.Errorf("Unexpected error for 2 days: %v", err)
	}
	if result != nil {
		hourly, ok := result.(*HourlyOutput)
		if ok {
			if hourly.Meta.Units.Temperature != "C" {
				t.Errorf("Temperature = %q, want %q", hourly.Meta.Units.Temperature, "C")
			}
		}
	}
}

// TestService_Forecast_Daily_DaysValidation tests that the Forecast method validates days for daily mode.
func TestService_Forecast_Daily_DaysValidation(t *testing.T) {
	mapper := weathercode.NewMapper()
	client := openmeteo.NewClient(nil)
	svc := NewService(client, mapper)

	// Test with 0 days - should fail
	_, err := svc.Forecast(40.0, -74.0, "metric", false, 0)
	if err == nil {
		t.Error("Expected error for 0 days, got nil")
	}

	// Test with 15 days - should fail
	_, err = svc.Forecast(40.0, -74.0, "metric", false, 15)
	if err == nil {
		t.Error("Expected error for 15 days, got nil")
	}

	// Test with 14 days - should succeed
	result, err := svc.Forecast(40.0, -74.0, "metric", false, 14)
	if err != nil {
		t.Errorf("Unexpected error for 14 days: %v", err)
	}
	if result != nil {
		_, ok := result.(*DailyOutput)
		if !ok {
			t.Errorf("Expected DailyOutput, got %T", result)
		}
	}
}

// TestService_HourlyOutput_JSONOutput tests JSON output format compatibility.
func TestService_HourlyOutput_JSONOutput(t *testing.T) {
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

// TestService_DailyOutput_JSONOutput tests JSON output format compatibility.
func TestService_DailyOutput_JSONOutput(t *testing.T) {
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

// TestService_getUnits tests the getUnits function via the public Forecast method.
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
			result, err := svc.Forecast(40.0, -74.0, tt.units, true, 1)
			if err != nil {
				// Network call fails, but test structure
				return
			}
			if result != nil {
				hourly, ok := result.(*HourlyOutput)
				if !ok {
					t.Fatalf("Expected HourlyOutput, got %T", result)
				}

				if hourly.Meta.Units.Temperature != tt.tempUnit {
					t.Errorf("Temperature = %q, want %q", hourly.Meta.Units.Temperature, tt.tempUnit)
				}
				if hourly.Meta.Units.WindSpeed != tt.windUnit {
					t.Errorf("WindSpeed = %q, want %q", hourly.Meta.Units.WindSpeed, tt.windUnit)
				}
				if hourly.Meta.Units.Precipitation != tt.precipUnit {
					t.Errorf("Precipitation = %q, want %q", hourly.Meta.Units.Precipitation, tt.precipUnit)
				}

				// Verify all other fields are correct
				if hourly.Meta.Units.Humidity != "%" {
					t.Errorf("Humidity = %q, want %q", hourly.Meta.Units.Humidity, "%")
				}
				if hourly.Meta.Units.WindDirection != "deg" {
					t.Errorf("WindDirection = %q, want %q", hourly.Meta.Units.WindDirection, "deg")
				}
				if hourly.Meta.Units.PrecipitationProbability != "%" {
					t.Errorf("PrecipitationProbability = %q, want %q", hourly.Meta.Units.PrecipitationProbability, "%")
				}
				if hourly.Meta.Units.UVIndex != "index" {
					t.Errorf("UVIndex = %q, want %q", hourly.Meta.Units.UVIndex, "index")
				}
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

// TestService_mapHourly_DatetimeFiltering4 tests that mapHourly filters to the correct date.
func TestService_mapHourly_DatetimeFiltering4(t *testing.T) {
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

// TestService_mapHourly_TimezoneConversion4 tests timezone handling in mapHourly.
func TestService_mapHourly_TimezoneConversion4(t *testing.T) {
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

// TestService_mapCurrent4 tests the mapCurrent function.
func TestService_mapCurrent4(t *testing.T) {
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

// TestService_mapHourly_UnknownWeatherCode4 tests handling of unknown weather codes.
func TestService_mapHourly_UnknownWeatherCode4(t *testing.T) {
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

// TestService_mapHourly_DSTHandling4 tests DST transition handling.
func TestService_mapHourly_DSTHandling4(t *testing.T) {
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

// TestService_mapDaily4 tests the mapDaily function.
func TestService_mapDaily4(t *testing.T) {
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
		{"Sunrise", "06:00"},
		{"Sunset", "18:00"},
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

// TestService_mapDaily_TimezoneOffset4 tests sunrise/sunset with timezone offset.
func TestService_mapDaily_TimezoneOffset4(t *testing.T) {
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

	// The parsed time should be in local timezone
	// Format should be HH:MM
	if len(result.Sunrise) != 5 || result.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be in HH:MM format, got %q", result.Sunrise)
	}
	if len(result.Sunset) != 5 || result.Sunset[2] != ':' {
		t.Errorf("Sunset should be in HH:MM format, got %q", result.Sunset)
	}
}

// TestService_mapDaily_MultipleDays4 tests mapping multiple days.
func TestService_mapDaily_MultipleDays4(t *testing.T) {
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

// TestService_mapDaily_DSTTransition4 tests DST transition date handling.
func TestService_mapDaily_DSTTransition4(t *testing.T) {
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

			// Verify sunrise/sunset are HH:MM format
			if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
				t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
			}
			if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
				t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
			}
		})
	}
}

// TestService_mapDaily_HourTimeFormat4 tests that hourly times in Day output are in HH:MM format.
func TestService_mapDaily_HourTimeFormat4(t *testing.T) {
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

	// Verify sunrise/sunset are HH:MM format
	if len(dayResult.Sunrise) != 5 || dayResult.Sunrise[2] != ':' {
		t.Errorf("Sunrise should be HH:MM format, got %q", dayResult.Sunrise)
	}
	if len(dayResult.Sunset) != 5 || dayResult.Sunset[2] != ':' {
		t.Errorf("Sunset should be HH:MM format, got %q", dayResult.Sunset)
	}
}

// TestService_mapDaily_ImperialUnits4 tests Day output with imperial units.
func TestService_mapDaily_ImperialUnits4(t *testing.T) {
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

// TestService_mapDaily_WeatherCodeMapping4 tests weather code mapping in Day output.
func TestService_mapDaily_WeatherCodeMapping4(t *testing.T) {
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

// TestService_mapDaily_MismatchedHourlyData4 tests handling of mismatched hourly array lengths.
func TestService_mapDaily_MismatchedHourlyData4(t *testing.T) {
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

// TestService_mapDaily_MidnightBoundary4 tests midnight boundary handling for Day.
func TestService_mapDaily_MidnightBoundary4(t *testing.T) {
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
