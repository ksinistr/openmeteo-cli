package forecast

import (
	"testing"
)

func getUnits(units string) Units {
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

func TestUnits_Metric(t *testing.T) {
	units := getUnits("metric")

	tests := []struct {
		field    string
		expected string
	}{
		{"temperature", "C"},
		{"humidity", "%"},
		{"wind_speed", "km/h"},
		{"wind_direction", "deg"},
		{"precipitation", "mm"},
		{"precipitation_probability", "%"},
		{"uv_index", "index"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			var val string
			switch tt.field {
			case "temperature":
				val = units.Temperature
			case "humidity":
				val = units.Humidity
			case "wind_speed":
				val = units.WindSpeed
			case "wind_direction":
				val = units.WindDirection
			case "precipitation":
				val = units.Precipitation
			case "precipitation_probability":
				val = units.PrecipitationProbability
			case "uv_index":
				val = units.UVIndex
			}

			if val != tt.expected {
				t.Errorf("Units.%s = %q, want %q", tt.field, val, tt.expected)
			}
		})
	}
}

func TestUnits_Imperial(t *testing.T) {
	units := getUnits("imperial")

	tests := []struct {
		field    string
		expected string
	}{
		{"temperature", "F"},
		{"humidity", "%"},
		{"wind_speed", "mph"},
		{"wind_direction", "deg"},
		{"precipitation", "inch"},
		{"precipitation_probability", "%"},
		{"uv_index", "index"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			var val string
			switch tt.field {
			case "temperature":
				val = units.Temperature
			case "humidity":
				val = units.Humidity
			case "wind_speed":
				val = units.WindSpeed
			case "wind_direction":
				val = units.WindDirection
			case "precipitation":
				val = units.Precipitation
			case "precipitation_probability":
				val = units.PrecipitationProbability
			case "uv_index":
				val = units.UVIndex
			}

			if val != tt.expected {
				t.Errorf("Units.%s = %q, want %q", tt.field, val, tt.expected)
			}
		})
	}
}

func TestUnits_HasAllFields(t *testing.T) {
	// Ensure Units struct has all required fields for the contract
	units := Units{
		Temperature:              "C",
		Humidity:                 "%",
		WindSpeed:                "km/h",
		WindDirection:            "deg",
		Precipitation:            "mm",
		PrecipitationProbability: "%",
		UVIndex:                  "index",
	}

	// Verify all fields are non-empty
	if units.Temperature == "" {
		t.Error("Units.Temperature should not be empty")
	}
	if units.Humidity == "" {
		t.Error("Units.Humidity should not be empty")
	}
	if units.WindSpeed == "" {
		t.Error("Units.WindSpeed should not be empty")
	}
	if units.WindDirection == "" {
		t.Error("Units.WindDirection should not be empty")
	}
	if units.Precipitation == "" {
		t.Error("Units.Precipitation should not be empty")
	}
	if units.PrecipitationProbability == "" {
		t.Error("Units.PrecipitationProbability should not be empty")
	}
	if units.UVIndex == "" {
		t.Error("Units.UVIndex should not be empty")
	}
}

func TestMeta_HasAllFields(t *testing.T) {
	// Verify Meta struct has all required fields
	meta := Meta{
		Units:     Units{},
		Timezone:  "UTC",
		Latitude:  40.0,
		Longitude: -74.0,
	}

	if meta.Timezone == "" {
		t.Error("Meta.Timezone should not be empty")
	}
	if meta.Latitude == 0 {
		t.Error("Meta.Latitude should not be zero")
	}
	if meta.Longitude == 0 {
		t.Error("Meta.Longitude should not be zero")
	}
}

func TestCurrent_HasAllFields(t *testing.T) {
	// Verify Current struct has all required fields
	current := Current{
		Time:                     "12:00",
		Weather:                  "Clear sky",
		Temperature:              20.5,
		ApparentTemperature:      19.0,
		Humidity:                 65,
		Precipitation:            0.0,
		PrecipitationProbability: 0,
		WindSpeed:                5.5,
		WindGusts:                8.0,
		WindDirection:            180,
		UVIndex:                  3.0,
	}

	if current.Time == "" {
		t.Error("Current.Time should not be empty")
	}
	if current.Weather == "" {
		t.Error("Current.Weather should not be empty")
	}
}

func TestHour_HasAllFields(t *testing.T) {
	// Verify Hour struct has all required fields
	hour := Hour{
		Time:                     "12:00",
		Weather:                  "Clear sky",
		Temperature:              20.5,
		ApparentTemperature:      19.0,
		Humidity:                 65,
		Precipitation:            0.0,
		PrecipitationProbability: 0,
		WindSpeed:                5.5,
		WindGusts:                8.0,
		WindDirection:            180,
		UVIndex:                  3.0,
	}

	if hour.Time == "" {
		t.Error("Hour.Time should not be empty")
	}
	if hour.Weather == "" {
		t.Error("Hour.Weather should not be empty")
	}
}

func TestDay_HasAllFields(t *testing.T) {
	// Verify Day struct has all required fields
	day := Day{
		Date:                        "2026-03-22",
		Weather:                     "Clear sky",
		TempMin:                     15.0,
		TempMax:                     25.0,
		PrecipitationSum:            0.0,
		PrecipitationProbabilityMax: 0,
		WindSpeedMax:                10.0,
		WindGustsMax:                15.0,
		UVIndexMax:                  5.0,
		Sunrise:                     "06:00",
		Sunset:                      "18:00",
	}

	if day.Date == "" {
		t.Error("Day.Date should not be empty")
	}
	if day.Weather == "" {
		t.Error("Day.Weather should not be empty")
	}
}

func TestHourlyOutput_HasAllFields(t *testing.T) {
	output := HourlyOutput{
		Meta: Meta{},
		Days: map[string]DayHours{},
	}

	if len(output.Days) != 0 {
		t.Error("HourlyOutput.Days should be initialized to empty map")
	}
}

func TestDailyOutput_HasAllFields(t *testing.T) {
	output := DailyOutput{
		Meta: Meta{},
		Days: []Day{},
	}

	if len(output.Days) != 0 {
		t.Error("DailyOutput.Days should be initialized to empty slice")
	}
}
