// Package output provides output formatting and writing utilities.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/toon-format/toon-go"

	"openmeteo-cli/internal/forecast"
)

// Writer handles writing responses to stdout/stderr.
type Writer struct {
	out io.Writer
	err io.Writer
}

// NewWriter creates a new output writer.
func NewWriter() *Writer {
	return &Writer{
		out: os.Stdout,
		err: os.Stderr,
	}
}

// SetOutput sets the output writer (default is os.Stdout).
func (w *Writer) SetOutput(out io.Writer) {
	w.out = out
}

// SetError sets the error output writer (default is os.Stderr).
func (w *Writer) SetError(err io.Writer) {
	w.err = err
}

// Write outputs the result in the specified format.
func (w *Writer) Write(data interface{}, format string) error {
	var err error

	switch format {
	case "json":
		err = writeJSON(data, w.out)
	case "toon":
		err = writeTOON(data, w.out)
	default:
		err = fmt.Errorf("unknown format: %s", format)
	}

	if err != nil {
		return &EncodingError{Err: err}
	}
	return nil
}

// WriteError writes an error to stderr.
func (w *Writer) WriteError(err error) {
	fmt.Fprintln(w.err, err)
}

// IsEncodingError checks if the error is an encoding error.
func IsEncodingError(err error) bool {
	var encodingError *EncodingError
	return errors.As(err, &encodingError)
}

// EncodingError represents an output encoding error.
type EncodingError struct {
	Err error
}

func (e *EncodingError) Error() string {
	return e.Err.Error()
}

// writeJSON encodes the data as JSON format.
func writeJSON(data interface{}, out io.Writer) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// writeTOON encodes the data as TOON format.
func writeTOON(data interface{}, out io.Writer) error {
	output, err := toon.Marshal(data)
	if err != nil {
		return err
	}
	_, err = out.Write(output)
	return err
}

// JSONEncoder encodes forecast output to JSON format.
type JSONEncoder struct{}

// NewJSONEncoder creates a new JSON encoder.
func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{}
}

// EncodeTodayTo encodes a TodayOutput to JSON format.
func (e *JSONEncoder) EncodeTodayTo(output *forecast.TodayOutput, out io.Writer) error {
	return writeJSON(output, out)
}

// EncodeDayTo encodes a DayOutput to JSON format.
func (e *JSONEncoder) EncodeDayTo(output *forecast.DayOutput, out io.Writer) error {
	return writeJSON(output, out)
}

// EncodeWeekTo encodes a WeekOutput to JSON format.
func (e *JSONEncoder) EncodeWeekTo(output *forecast.WeekOutput, out io.Writer) error {
	return writeJSON(output, out)
}

// ToonEncoder encodes forecast output to TOON format.
type ToonEncoder struct{}

// NewToonEncoder creates a new TOON encoder.
func NewToonEncoder() *ToonEncoder {
	return &ToonEncoder{}
}

// EncodeToday encodes a TodayOutput to TOON format.
func (e *ToonEncoder) EncodeToday(output *forecast.TodayOutput) (string, error) {
	units := buildUnitsTOON(output.Meta.Units)
	current := buildCurrentTOON(output.Current)
	hours := buildHoursTOON(output.Hours)

	return buildTodayOutputTOON(output.Meta, units, current, hours), nil
}

// EncodeDay encodes a DayOutput to TOON format.
func (e *ToonEncoder) EncodeDay(output *forecast.DayOutput) (string, error) {
	units := buildUnitsTOON(output.Meta.Units)
	day := buildDayTOON(output.Day)
	hours := buildHoursTOON(output.Hours)

	return buildDayOutputTOON(output.Meta, units, day, hours), nil
}

// EncodeWeek encodes a WeekOutput to TOON format.
func (e *ToonEncoder) EncodeWeek(output *forecast.WeekOutput) (string, error) {
	units := buildUnitsTOON(output.Meta.Units)
	days := buildDaysTOON(output.Days)

	return buildWeekOutputTOON(output.Meta, units, days), nil
}

// buildUnitsTOON builds the units table TOON string.
func buildUnitsTOON(units forecast.Units) string {
	return "# units\n  temperature " + units.Temperature +
		"\n  humidity " + units.Humidity +
		"\n  wind_speed " + units.WindSpeed +
		"\n  wind_direction " + units.WindDirection +
		"\n  precipitation " + units.Precipitation +
		"\n  precipitation_probability " + units.PrecipitationProbability +
		"\n  uv_index " + units.UVIndex
}

// buildCurrentTOON builds the current conditions TOON string.
func buildCurrentTOON(current forecast.Current) string {
	return "# current\n  time " + current.Time +
		"\n  weather " + current.Weather +
		"\n  temperature " + formatFloatNoTrailing(current.Temperature) +
		"\n  apparent_temperature " + formatFloatNoTrailing(current.ApparentTemperature) +
		"\n  humidity " + intToString(current.Humidity) +
		"\n  precipitation " + formatFloatNoTrailing(current.Precipitation) +
		"\n  precipitation_probability " + intToString(current.PrecipitationProbability) +
		"\n  wind_speed " + formatFloatNoTrailing(current.WindSpeed) +
		"\n  wind_gusts " + formatFloatNoTrailing(current.WindGusts) +
		"\n  wind_direction " + intToString(current.WindDirection) +
		"\n  uv_index " + formatFloatNoTrailing(current.UVIndex)
}

// buildHourTOON builds a single hour TOON string.
func buildHourTOON(hour forecast.Hour) string {
	return "# hour\n  time " + hour.Time +
		"\n  weather " + hour.Weather +
		"\n  temperature " + formatFloatNoTrailing(hour.Temperature) +
		"\n  apparent_temperature " + formatFloatNoTrailing(hour.ApparentTemperature) +
		"\n  humidity " + intToString(hour.Humidity) +
		"\n  precipitation " + formatFloatNoTrailing(hour.Precipitation) +
		"\n  precipitation_probability " + intToString(hour.PrecipitationProbability) +
		"\n  wind_speed " + formatFloatNoTrailing(hour.WindSpeed) +
		"\n  wind_gusts " + formatFloatNoTrailing(hour.WindGusts) +
		"\n  wind_direction " + intToString(hour.WindDirection) +
		"\n  uv_index " + formatFloatNoTrailing(hour.UVIndex)
}

// buildHoursTOON builds all hours TOON string.
func buildHoursTOON(hours []forecast.Hour) string {
	if len(hours) == 0 {
		return ""
	}

	result := "# hours"
	for _, h := range hours {
		result += "\n" + buildHourTOON(h)
	}
	return result
}

// buildDayTOON builds the day TOON string.
func buildDayTOON(day forecast.Day) string {
	return "# day\n  date " + day.Date +
		"\n  weather " + day.Weather +
		"\n  temp_min " + formatFloatNoTrailing(day.TempMin) +
		"\n  temp_max " + formatFloatNoTrailing(day.TempMax) +
		"\n  precipitation_sum " + formatFloatNoTrailing(day.PrecipitationSum) +
		"\n  precipitation_probability_max " + intToString(day.PrecipitationProbabilityMax) +
		"\n  wind_speed_max " + formatFloatNoTrailing(day.WindSpeedMax) +
		"\n  wind_gusts_max " + formatFloatNoTrailing(day.WindGustsMax) +
		"\n  uv_index_max " + formatFloatNoTrailing(day.UVIndexMax) +
		"\n  sunrise " + day.Sunrise +
		"\n  sunset " + day.Sunset
}

// buildDaysTOON builds all days TOON string.
func buildDaysTOON(days []forecast.Day) string {
	if len(days) == 0 {
		return ""
	}

	result := "# days"
	for _, d := range days {
		result += "\n" + buildDayTOON(d)
	}
	return result
}

// buildMetaTOON builds the meta table TOON string.
func buildMetaTOON(meta forecast.Meta) string {
	return "# meta\n  generated_at " + meta.GeneratedAt.Format("2006-01-02T15:04") +
		"\n  timezone " + meta.Timezone +
		"\n  latitude " + formatFloatNoTrailing(meta.Latitude) +
		"\n  longitude " + formatFloatNoTrailing(meta.Longitude)
}

// buildTodayOutputTOON builds the complete today output TOON string.
func buildTodayOutputTOON(meta forecast.Meta, units, current string, hours string) string {
	result := buildMetaTOON(meta)
	result += "\n" + units
	result += "\n" + current
	if hours != "" {
		result += "\n" + hours
	}
	return result
}

// buildDayOutputTOON builds the complete day output TOON string.
func buildDayOutputTOON(meta forecast.Meta, units, day string, hours string) string {
	result := buildMetaTOON(meta)
	result += "\n" + units
	result += "\n" + day
	if hours != "" {
		result += "\n" + hours
	}
	return result
}

// buildWeekOutputTOON builds the complete week output TOON string.
func buildWeekOutputTOON(meta forecast.Meta, units string, days string) string {
	result := buildMetaTOON(meta)
	result += "\n" + units
	if days != "" {
		result += "\n" + days
	}
	return result
}

// formatFloatNoTrailing formats a float without trailing zeros for TOON compatibility.
func formatFloatNoTrailing(f float64) string {
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
		for len(s) > idx+1 && s[len(s)-1] == '0' {
			s = s[:len(s)-1]
		}
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
