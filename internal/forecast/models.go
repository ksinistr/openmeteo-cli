// Package forecast contains the public output models for weather forecasts.
//
// This package defines the response models that are used for the forecast command.
// It also provides the Service interface for weather forecast shaping logic.
package forecast

import (
	"encoding/json"

	"github.com/toon-format/toon-go"
)

// ForecastOutput is the unified response for the forecast command.
// It supports variable-driven current, hourly, and daily sections.
type ForecastOutput struct {
	Meta    Meta             `json:"meta" toon:"meta"`
	Current *CurrentOutput   `json:"current,omitempty" toon:"current,omitempty"`
	Hourly  *HourlyOutputNew `json:"hourly,omitempty" toon:"hourly,omitempty"`
	Daily   *DailyOutputNew  `json:"daily,omitempty" toon:"daily,omitempty"`
}

// CurrentOutput represents current weather conditions with variable-driven fields.
// The "time" field is always included as the first field.
// MarshalJSON produces flat JSON objects like {"time": "...", "temperature_2m": 7.2}.
type CurrentOutput struct {
	Fields []string       `json:"-" toon:"-"` // Field names in order (for TOON output)
	Values map[string]any `json:"-" toon:"-"` // Field values keyed by name
}

// MarshalJSON implements json.Marshaler to produce flat JSON output.
func (c *CurrentOutput) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	return json.Marshal(c.Values)
}

// ToTOON converts CurrentOutput to a toon.Object with ordered fields.
func (c *CurrentOutput) ToTOON() toon.Object {
	if c == nil {
		return toon.Object{}
	}
	fields := make([]toon.Field, 0, len(c.Fields))
	for _, name := range c.Fields {
		if val, ok := c.Values[name]; ok {
			fields = append(fields, toon.Field{Key: name, Value: val})
		}
	}
	return toon.Object{Fields: fields}
}

// HourlyOutputNew represents hourly forecast data grouped by day with variable-driven fields.
// Each day group always includes "date" and "weekday" as the first two fields.
// Each hour entry always includes "time" as the first field.
// MarshalJSON serializes directly as an array to match the documented public contract.
type HourlyOutputNew struct {
	Days []HourlyDay `json:"-" toon:"days"` // Exported for TOON, custom JSON marshaling
}

// MarshalJSON implements json.Marshaler to serialize hourly data directly as an array.
// This matches the documented public contract: "hourly": [{"date": "...", "hours": [...]}]
func (h *HourlyOutputNew) MarshalJSON() ([]byte, error) {
	if h == nil {
		return []byte("null"), nil
	}
	return json.Marshal(h.Days)
}

// ToTOON converts HourlyOutputNew to a slice of toon.Object with ordered fields.
func (h *HourlyOutputNew) ToTOON() []any {
	if h == nil {
		return nil
	}
	result := make([]any, 0, len(h.Days))
	for i := range h.Days {
		result = append(result, h.Days[i].ToTOON())
	}
	return result
}

// HourlyDay contains hourly forecast data for a single day.
// MarshalJSON produces flat JSON for the day-level fields.
type HourlyDay struct {
	Fields []string       `json:"-" toon:"-"`         // Field names in order (date, weekday, ...)
	Values map[string]any `json:"-" toon:"-"`         // Field values keyed by name
	Hours  []HourlyEntry  `json:"hours" toon:"hours"` // Hourly entries
}

// MarshalJSON implements json.Marshaler to produce flat JSON for day fields.
func (h *HourlyDay) MarshalJSON() ([]byte, error) {
	if h == nil {
		return []byte("null"), nil
	}
	// Create a map that includes both day values and hours
	result := make(map[string]any)
	for k, v := range h.Values {
		result[k] = v
	}
	result["hours"] = h.Hours
	return json.Marshal(result)
}

// ToTOON converts HourlyDay to a toon.Object with ordered fields.
func (h *HourlyDay) ToTOON() toon.Object {
	if h == nil {
		return toon.Object{}
	}
	fields := make([]toon.Field, 0, len(h.Fields)+1)
	for _, name := range h.Fields {
		if val, ok := h.Values[name]; ok {
			fields = append(fields, toon.Field{Key: name, Value: val})
		}
	}
	// Add hours as array of objects
	hours := make([]any, 0, len(h.Hours))
	for i := range h.Hours {
		hours = append(hours, h.Hours[i].ToTOON())
	}
	fields = append(fields, toon.Field{Key: "hours", Value: hours})
	return toon.Object{Fields: fields}
}

// HourlyEntry represents a single hourly data point with variable-driven fields.
// The "time" field is always included as the first field.
// MarshalJSON produces flat JSON like {"time": "00:00", "temperature_2m": 6.1}.
type HourlyEntry struct {
	Fields []string       `json:"-" toon:"-"` // Field names in order (time, ...)
	Values map[string]any `json:"-" toon:"-"` // Field values keyed by name
}

// MarshalJSON implements json.Marshaler to produce flat JSON output.
func (h *HourlyEntry) MarshalJSON() ([]byte, error) {
	if h == nil {
		return []byte("null"), nil
	}
	return json.Marshal(h.Values)
}

// ToTOON converts HourlyEntry to a toon.Object with ordered fields.
func (h *HourlyEntry) ToTOON() toon.Object {
	if h == nil {
		return toon.Object{}
	}
	fields := make([]toon.Field, 0, len(h.Fields))
	for _, name := range h.Fields {
		if val, ok := h.Values[name]; ok {
			fields = append(fields, toon.Field{Key: name, Value: val})
		}
	}
	return toon.Object{Fields: fields}
}

// DailyOutputNew represents daily forecast data with variable-driven fields.
// Each row always includes "date" and "weekday" as the first two fields.
// MarshalJSON serializes directly as an array to match the documented public contract.
type DailyOutputNew struct {
	Fields []string   `json:"-" toon:"-"`    // Field names in order (date, weekday, ...)
	Rows   []DailyRow `json:"-" toon:"rows"` // Daily rows, custom JSON marshaling
}

// MarshalJSON implements json.Marshaler to serialize daily data directly as an array.
// This matches the documented public contract: "daily": [{"date": "...", ...}]
func (d *DailyOutputNew) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	return json.Marshal(d.Rows)
}

// ToTOON converts DailyOutputNew to a slice of toon.Object with ordered fields.
func (d *DailyOutputNew) ToTOON() []any {
	if d == nil {
		return nil
	}
	result := make([]any, 0, len(d.Rows))
	for i := range d.Rows {
		result = append(result, d.Rows[i].toTOONWithFields(d.Fields))
	}
	return result
}

// toTOONWithFields converts DailyRow to a toon.Object with ordered fields using provided field order.
func (d *DailyRow) toTOONWithFields(fieldOrder []string) toon.Object {
	if d == nil {
		return toon.Object{}
	}
	fields := make([]toon.Field, 0, len(fieldOrder))
	for _, name := range fieldOrder {
		if val, ok := d.Values[name]; ok {
			fields = append(fields, toon.Field{Key: name, Value: val})
		}
	}
	return toon.Object{Fields: fields}
}

// DailyRow represents a single daily data point with variable-driven fields.
// MarshalJSON produces flat JSON like {"date": "...", "weekday": "...", "temperature_2m_min": 11.4}.
type DailyRow struct {
	Fields []string       `json:"-" toon:"-"` // Field names in order (for TOON output)
	Values map[string]any `json:"-" toon:"-"` // Field values keyed by name
}

// MarshalJSON implements json.Marshaler to produce flat JSON output.
func (d *DailyRow) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	return json.Marshal(d.Values)
}

// Meta contains metadata about the forecast.
type Meta struct {
	Units     Units     `json:"units" toon:"units"`
	Timezone  string    `json:"timezone" toon:"timezone"`
	Latitude  float64   `json:"latitude" toon:"latitude"`
	Longitude float64   `json:"longitude" toon:"longitude"`
	Location  *Location `json:"location,omitempty" toon:"location,omitempty"`
}

// Location contains resolved place metadata from geocoding.
type Location struct {
	Name    string `json:"name" toon:"name"`
	Country string `json:"country" toon:"country"`
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

// Legacy types for backward compatibility during transition.

// HourlyOutput represents the response for the hourly forecast (legacy).
// It includes metadata and hourly forecast data grouped by day.
type HourlyOutput struct {
	Meta Meta                `json:"meta" toon:"meta"`
	Days map[string]DayHours `json:"days" toon:"days"`
}

// DayHours contains hourly forecast data for a single day.
type DayHours struct {
	Hours []Hour `json:"hours" toon:"hours"`
}

// DailyOutput represents the response for the daily forecast (legacy).
// It includes metadata and daily forecast data.
type DailyOutput struct {
	Meta Meta  `json:"meta" toon:"meta"`
	Days []Day `json:"days" toon:"days"`
}

// Hour represents an hourly forecast entry (legacy).
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

// Day represents a daily forecast entry (legacy).
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

// Current represents the current weather conditions (legacy, kept for compatibility).
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
