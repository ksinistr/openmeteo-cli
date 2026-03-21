// Package weathercode provides weather code to description mapping.
//
// This package uses the Open-Meteo weather interpretation codes (WW) to map
// numeric weather codes to human-readable descriptions.
package weathercode

import (
	"fmt"
)

// Mapper translates Open-Meteo weather codes to descriptions.
type Mapper struct {
	codes map[int]string
}

// NewMapper creates a new weather code mapper with the standard codes.
func NewMapper() *Mapper {
	return &Mapper{
		codes: map[int]string{
			0:  "Clear sky",
			1:  "Mainly clear",
			2:  "Partly cloudy",
			3:  "Overcast",
			45: "Fog",
			48: "Depositing rime fog",
			51: "Light drizzle",
			53: "Moderate drizzle",
			55: "Dense drizzle",
			56: "Light freezing drizzle",
			57: "Dense freezing drizzle",
			61: "Slight rain",
			63: "Moderate rain",
			65: "Heavy rain",
			66: "Light freezing rain",
			67: "Heavy freezing rain",
			71: "Slight snow fall",
			73: "Moderate snow fall",
			75: "Heavy snow fall",
			77: "Snow grains",
			80: "Slight rain showers",
			81: "Moderate rain showers",
			82: "Violent rain showers",
			85: "Slight snow showers",
			86: "Heavy snow showers",
			95: "Thunderstorm",
			96: "Thunderstorm with slight hail",
			99: "Thunderstorm with heavy hail",
		},
	}
}

// GetDescription returns the human-readable description for a weather code.
// Returns "Unknown weather code: <code>" for unknown codes.
func (m *Mapper) GetDescription(code int) string {
	if desc, ok := m.codes[code]; ok {
		return desc
	}
	return fmt.Sprintf("Unknown weather code: %d", code)
}

// GetCode returns the weather code for a description, or -1 if not found.
func (m *Mapper) GetCode(description string) int {
	for code, desc := range m.codes {
		if desc == description {
			return code
		}
	}
	return -1
}
