package output

import (
	"fmt"
	"strings"

	"openmeteo-cli/internal/forecast"
	"openmeteo-cli/internal/weathercode"
)

// ToonEncoder encodes forecast output to TOON format.
type ToonEncoder struct {
	mapper *weathercode.Mapper
}

// NewToonEncoder creates a new TOON encoder.
func NewToonEncoder(mapper *weathercode.Mapper) *ToonEncoder {
	return &ToonEncoder{mapper: mapper}
}

// EncodeToday encodes a TodayOutput to TOON format.
func (e *ToonEncoder) EncodeToday(output *forecast.TodayOutput) (string, error) {
	units := e.buildUnitsTOON(output.Meta.Units)
	current := e.buildCurrentTOON(output.Current)
	hours := e.buildHoursTOON(output.Hours)

	return e.buildTodayOutputTOON(output.Meta, units, current, hours), nil
}

// EncodeDay encodes a DayOutput to TOON format.
func (e *ToonEncoder) EncodeDay(output *forecast.DayOutput) (string, error) {
	units := e.buildUnitsTOON(output.Meta.Units)
	day := e.buildDayTOON(output.Day)
	hours := e.buildHoursTOON(output.Hours)

	return e.buildDayOutputTOON(output.Meta, units, day, hours), nil
}

// EncodeWeek encodes a WeekOutput to TOON format.
func (e *ToonEncoder) EncodeWeek(output *forecast.WeekOutput) (string, error) {
	units := e.buildUnitsTOON(output.Meta.Units)
	days := e.buildDaysTOON(output.Days)

	return e.buildWeekOutputTOON(output.Meta, units, days), nil
}

// buildUnitsTOON builds the units table TOON string.
func (e *ToonEncoder) buildUnitsTOON(units forecast.Units) string {
	return "# units\n  temperature " + units.Temperature +
		"\n  humidity " + units.Humidity +
		"\n  wind_speed " + units.WindSpeed +
		"\n  wind_direction " + units.WindDirection +
		"\n  precipitation " + units.Precipitation +
		"\n  precipitation_probability " + units.PrecipitationProbability +
		"\n  uv_index " + units.UVIndex
}

// buildCurrentTOON builds the current conditions TOON string.
func (e *ToonEncoder) buildCurrentTOON(current forecast.Current) string {
	return "# current\n  time " + current.Time +
		"\n  weather " + current.Weather +
		"\n  temperature " + floatToString(current.Temperature) +
		"\n  apparent_temperature " + floatToString(current.ApparentTemperature) +
		"\n  humidity " + intToString(current.Humidity) +
		"\n  precipitation " + floatToString(current.Precipitation) +
		"\n  precipitation_probability " + intToString(current.PrecipitationProbability) +
		"\n  wind_speed " + floatToString(current.WindSpeed) +
		"\n  wind_gusts " + floatToString(current.WindGusts) +
		"\n  wind_direction " + intToString(current.WindDirection) +
		"\n  uv_index " + floatToString(current.UVIndex)
}

// buildHourTOON builds a single hour TOON string.
func (e *ToonEncoder) buildHourTOON(hour forecast.Hour) string {
	return "# hour\n  time " + hour.Time +
		"\n  weather " + hour.Weather +
		"\n  temperature " + floatToString(hour.Temperature) +
		"\n  apparent_temperature " + floatToString(hour.ApparentTemperature) +
		"\n  humidity " + intToString(hour.Humidity) +
		"\n  precipitation " + floatToString(hour.Precipitation) +
		"\n  precipitation_probability " + intToString(hour.PrecipitationProbability) +
		"\n  wind_speed " + floatToString(hour.WindSpeed) +
		"\n  wind_gusts " + floatToString(hour.WindGusts) +
		"\n  wind_direction " + intToString(hour.WindDirection) +
		"\n  uv_index " + floatToString(hour.UVIndex)
}

// buildHoursTOON builds all hours TOON string.
func (e *ToonEncoder) buildHoursTOON(hours []forecast.Hour) string {
	if len(hours) == 0 {
		return ""
	}

	result := "# hours"
	for _, h := range hours {
		result += "\n" + e.buildHourTOON(h)
	}
	return result
}

// buildDayTOON builds the day TOON string.
func (e *ToonEncoder) buildDayTOON(day forecast.Day) string {
	return "# day\n  date " + day.Date +
		"\n  weather " + day.Weather +
		"\n  temp_min " + floatToString(day.TempMin) +
		"\n  temp_max " + floatToString(day.TempMax) +
		"\n  precipitation_sum " + floatToString(day.PrecipitationSum) +
		"\n  precipitation_probability_max " + intToString(day.PrecipitationProbabilityMax) +
		"\n  wind_speed_max " + floatToString(day.WindSpeedMax) +
		"\n  wind_gusts_max " + floatToString(day.WindGustsMax) +
		"\n  uv_index_max " + floatToString(day.UVIndexMax) +
		"\n  sunrise " + day.Sunrise +
		"\n  sunset " + day.Sunset
}

// buildDaysTOON builds all days TOON string.
func (e *ToonEncoder) buildDaysTOON(days []forecast.Day) string {
	if len(days) == 0 {
		return ""
	}

	result := "# days"
	for _, d := range days {
		result += "\n" + e.buildDayTOON(d)
	}
	return result
}

// buildMetaTOON builds the meta table TOON string.
func (e *ToonEncoder) buildMetaTOON(meta forecast.Meta) string {
	return "# meta\n  generated_at " + meta.GeneratedAt.Format("2006-01-02T15:04") +
		"\n  timezone " + meta.Timezone +
		"\n  latitude " + floatToString(meta.Latitude) +
		"\n  longitude " + floatToString(meta.Longitude)
}

// buildTodayOutputTOON builds the complete today output TOON string.
func (e *ToonEncoder) buildTodayOutputTOON(meta forecast.Meta, units, current string, hours string) string {
	result := e.buildMetaTOON(meta)
	result += "\n" + units
	result += "\n" + current
	if hours != "" {
		result += "\n" + hours
	}
	return result
}

// buildDayOutputTOON builds the complete day output TOON string.
func (e *ToonEncoder) buildDayOutputTOON(meta forecast.Meta, units, day string, hours string) string {
	result := e.buildMetaTOON(meta)
	result += "\n" + units
	result += "\n" + day
	if hours != "" {
		result += "\n" + hours
	}
	return result
}

// buildWeekOutputTOON builds the complete week output TOON string.
func (e *ToonEncoder) buildWeekOutputTOON(meta forecast.Meta, units string, days string) string {
	result := e.buildMetaTOON(meta)
	result += "\n" + units
	if days != "" {
		result += "\n" + days
	}
	return result
}

// floatToString converts float64 to string without scientific notation.
func floatToString(f float64) string {
	return formatFloatNoTrailing(f)
}

// formatFloatNoTrailing formats a float without trailing zeros.
func formatFloatNoTrailing(f float64) string {
	// Use fmt.Sprintf to handle the float, then clean up
	s := fmt.Sprintf("%.6g", f)

	// Handle special cases
	if s == "+Inf" {
		return "inf"
	}
	if s == "-Inf" {
		return "-inf"
	}
	if s == "NaN" {
		return "NaN"
	}

	// Remove trailing zeros after decimal point
	if idx := strings.IndexByte(s, '.'); idx >= 0 {
		// Remove trailing zeros
		for len(s) > idx+1 && s[len(s)-1] == '0' {
			s = s[:len(s)-1]
		}
		// Remove trailing decimal point if no decimals remain
		if len(s) > idx+1 && s[len(s)-1] == '.' {
			s = s[:len(s)-1]
		}
	}

	return s
}

// intToString converts int to string.
func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}
