// Package forecast contains the public output models for weather forecasts.
//
// This package defines the response models that are used for the forecast command.
// It also provides the Service interface for weather forecast shaping logic.
package forecast

import (
	"time"
)

// HourlyOutput represents the response for the hourly forecast.
// It includes metadata and hourly forecast data grouped by day.
type HourlyOutput struct {
	Meta Meta              `json:"meta" toon:"meta"`
	Days map[string]DayHours `json:"days" toon:"days"`
}

// DayHours contains hourly forecast data for a single day.
type DayHours struct {
	Hours []Hour `json:"hours" toon:"hours"`
}

// DailyOutput represents the response for the daily forecast.
// It includes metadata and daily forecast data.
type DailyOutput struct {
	Meta Meta  `json:"meta" toon:"meta"`
	Days []Day `json:"days" toon:"days"`
}

// Meta contains metadata about the forecast.
type Meta struct {
	GeneratedAt time.Time `json:"generated_at" toon:"generated_at"`
	Units       Units     `json:"units" toon:"units"`
	Timezone    string    `json:"timezone" toon:"timezone"`
	Latitude    float64   `json:"latitude" toon:"latitude"`
	Longitude   float64   `json:"longitude" toon:"longitude"`
}

// Units contains unit information for all numeric fields.
type Units struct {
	Temperature              string `json:"temperature" toon:"temperature"`
	Humidity                 string `json:"humidity" toon:"humidity"`
	WindSpeed                string `json:"wind_speed" toon:"wind_speed"`
	WindDirection            string `json:"wind_direction" toon:"wind_direction"`
	Precipitation            string `json:"precipitation" toon:"precipitation"`
	PrecipitationProbability string `json:"precipitation_probability" toon:"precipitation_probability"`
	UVIndex                  string `json:"uv_index" toon:"uv_index"`
}

// Hour represents an hourly forecast entry.
type Hour struct {
	Time                     string  `json:"time" toon:"time"`
	Weather                  string  `json:"weather" toon:"weather"`
	Temperature              float64 `json:"temperature" toon:"temperature"`
	ApparentTemperature      float64 `json:"apparent_temperature" toon:"apparent_temperature"`
	Humidity                 int     `json:"humidity" toon:"humidity"`
	Precipitation            float64 `json:"precipitation" toon:"precipitation"`
	PrecipitationProbability int     `json:"precipitation_probability" toon:"precipitation_probability"`
	WindSpeed                float64 `json:"wind_speed" toon:"wind_speed"`
	WindGusts                float64 `json:"wind_gusts" toon:"wind_gusts"`
	WindDirection            int     `json:"wind_direction" toon:"wind_direction"`
	UVIndex                  float64 `json:"uv_index" toon:"uv_index"`
}

// Day represents a daily forecast entry.
type Day struct {
	Date                        string  `json:"date" toon:"date"`
	Weather                     string  `json:"weather" toon:"weather"`
	TempMin                     float64 `json:"temp_min" toon:"temp_min"`
	TempMax                     float64 `json:"temp_max" toon:"temp_max"`
	PrecipitationSum            float64 `json:"precipitation_sum" toon:"precipitation_sum"`
	PrecipitationProbabilityMax int     `json:"precipitation_probability_max" toon:"precipitation_probability_max"`
	WindSpeedMax                float64 `json:"wind_speed_max" toon:"wind_speed_max"`
	WindGustsMax                float64 `json:"wind_gusts_max" toon:"wind_gusts_max"`
	UVIndexMax                  float64 `json:"uv_index_max" toon:"uv_index_max"`
	Sunrise                     string  `json:"sunrise" toon:"sunrise"`
	Sunset                      string  `json:"sunset" toon:"sunset"`
}

// Current represents the current weather conditions (kept for potential future use).
type Current struct {
	Time                     string  `json:"time,omitempty" toon:"time,omitempty"`
	Weather                  string  `json:"weather,omitempty" toon:"weather,omitempty"`
	Temperature              float64 `json:"temperature,omitempty" toon:"temperature,omitempty"`
	ApparentTemperature      float64 `json:"apparent_temperature,omitempty" toon:"apparent_temperature,omitempty"`
	Humidity                 int     `json:"humidity,omitempty" toon:"humidity,omitempty"`
	Precipitation            float64 `json:"precipitation,omitempty" toon:"precipitation,omitempty"`
	PrecipitationProbability int     `json:"precipitation_probability,omitempty" toon:"precipitation_probability,omitempty"`
	WindSpeed                float64 `json:"wind_speed,omitempty" toon:"wind_speed,omitempty"`
	WindGusts                float64 `json:"wind_gusts,omitempty" toon:"wind_gusts,omitempty"`
	WindDirection            int     `json:"wind_direction,omitempty" toon:"wind_direction,omitempty"`
	UVIndex                  float64 `json:"uv_index,omitempty" toon:"uv_index,omitempty"`
}
