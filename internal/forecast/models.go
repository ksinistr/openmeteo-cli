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
	Meta Meta              `json:"meta"`
	Days map[string]DayHours `json:"days"`
}

// DayHours contains hourly forecast data for a single day.
type DayHours struct {
	Hours []Hour `json:"hours"`
}

// DailyOutput represents the response for the daily forecast.
// It includes metadata and daily forecast data.
type DailyOutput struct {
	Meta Meta  `json:"meta"`
	Days []Day `json:"days"`
}

// Meta contains metadata about the forecast.
type Meta struct {
	GeneratedAt time.Time `json:"generated_at"`
	Units       Units     `json:"units"`
	Timezone    string    `json:"timezone"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
}

// Units contains unit information for all numeric fields.
type Units struct {
	Temperature              string `json:"temperature"`
	Humidity                 string `json:"humidity"`
	WindSpeed                string `json:"wind_speed"`
	WindDirection            string `json:"wind_direction"`
	Precipitation            string `json:"precipitation"`
	PrecipitationProbability string `json:"precipitation_probability"`
	UVIndex                  string `json:"uv_index"`
}

// Hour represents an hourly forecast entry.
type Hour struct {
	Time                     string  `json:"time"`
	Weather                  string  `json:"weather"`
	Temperature              float64 `json:"temperature"`
	ApparentTemperature      float64 `json:"apparent_temperature"`
	Humidity                 int     `json:"humidity"`
	Precipitation            float64 `json:"precipitation"`
	PrecipitationProbability int     `json:"precipitation_probability"`
	WindSpeed                float64 `json:"wind_speed"`
	WindGusts                float64 `json:"wind_gusts"`
	WindDirection            int     `json:"wind_direction"`
	UVIndex                  float64 `json:"uv_index"`
}

// Day represents a daily forecast entry.
type Day struct {
	Date                        string  `json:"date"`
	Weather                     string  `json:"weather"`
	TempMin                     float64 `json:"temp_min"`
	TempMax                     float64 `json:"temp_max"`
	PrecipitationSum            float64 `json:"precipitation_sum"`
	PrecipitationProbabilityMax int     `json:"precipitation_probability_max"`
	WindSpeedMax                float64 `json:"wind_speed_max"`
	WindGustsMax                float64 `json:"wind_gusts_max"`
	UVIndexMax                  float64 `json:"uv_index_max"`
	Sunrise                     string  `json:"sunrise"`
	Sunset                      string  `json:"sunset"`
}

// Current represents the current weather conditions (kept for potential future use).
type Current struct {
	Time                     string  `json:"time,omitempty"`
	Weather                  string  `json:"weather,omitempty"`
	Temperature              float64 `json:"temperature,omitempty"`
	ApparentTemperature      float64 `json:"apparent_temperature,omitempty"`
	Humidity                 int     `json:"humidity,omitempty"`
	Precipitation            float64 `json:"precipitation,omitempty"`
	PrecipitationProbability int     `json:"precipitation_probability,omitempty"`
	WindSpeed                float64 `json:"wind_speed,omitempty"`
	WindGusts                float64 `json:"wind_gusts,omitempty"`
	WindDirection            int     `json:"wind_direction,omitempty"`
	UVIndex                  float64 `json:"uv_index,omitempty"`
}
