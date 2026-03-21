// Package forecast contains the public output models for weather forecasts.
//
// This package defines the response models that are shared across all forecast
// commands (today, day, week). It also provides the Service interface for
// weather forecast shaping logic.
package forecast

import (
	"time"
)

// TodayOutput represents the response for the today command.
// It includes metadata, current conditions, and hourly forecast for the day.
type TodayOutput struct {
	Meta    Meta    `json:"meta"`
	Current Current `json:"current"`
	Hours   []Hour  `json:"hours"`
}

// DayOutput represents the response for the day command.
// It includes metadata, a daily summary, and hourly forecast for a specific date.
type DayOutput struct {
	Meta  Meta   `json:"meta"`
	Day   Day    `json:"day"`
	Hours []Hour `json:"hours"`
}

// WeekOutput represents the response for the week command.
// It includes metadata and 7 daily forecasts.
type WeekOutput struct {
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

// Current represents the current weather conditions.
type Current struct {
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
