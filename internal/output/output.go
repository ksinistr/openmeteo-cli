// Package output provides output formatting and writing utilities.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

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
	if data == nil {
		return &EncodingError{Err: fmt.Errorf("cannot write nil data")}
	}

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
func (w *Writer) WriteError(err error) error {
	_, err = fmt.Fprintln(w.err, err)
	return err
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

// writeTOON encodes the data as TOON format using toon-go library.
func writeTOON(data interface{}, out io.Writer) error {
	var toonData interface{}
	var err error

	switch d := data.(type) {
	case *forecast.TodayOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil TodayOutput")
		}
		toonData = convertTodayToTOON(d)
	case *forecast.DayOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil DayOutput")
		}
		toonData = convertDayToTOON(d)
	case *forecast.WeekOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil WeekOutput")
		}
		toonData = convertWeekToTOON(d)
	default:
		return fmt.Errorf("unsupported type for TOON encoding: %T", data)
	}

	// Use toon-go to marshal the data
	output, err := toon.MarshalString(toonData, toon.WithIndent(2))
	if err != nil {
		return err
	}

	_, err = out.Write([]byte(output))
	return err
}

// JSONEncoder encodes forecast output to JSON format.
// Deprecated: Use Writer.Write with format "json" instead.
type JSONEncoder struct{}

// NewJSONEncoder creates a new JSON encoder.
// Deprecated: Use Writer.Write with format "json" instead.
func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{}
}

// EncodeTodayTo encodes a TodayOutput to JSON format.
// Deprecated: Use Writer.Write with format "json" instead.
func (e *JSONEncoder) EncodeTodayTo(output *forecast.TodayOutput, out io.Writer) error {
	return writeJSON(output, out)
}

// EncodeDayTo encodes a DayOutput to JSON format.
// Deprecated: Use Writer.Write with format "json" instead.
func (e *JSONEncoder) EncodeDayTo(output *forecast.DayOutput, out io.Writer) error {
	return writeJSON(output, out)
}

// EncodeWeekTo encodes a WeekOutput to JSON format.
// Deprecated: Use Writer.Write with format "json" instead.
func (e *JSONEncoder) EncodeWeekTo(output *forecast.WeekOutput, out io.Writer) error {
	return writeJSON(output, out)
}

// ToonEncoder encodes forecast output to TOON format.
// Deprecated: Use Writer.Write with format "toon" instead.
type ToonEncoder struct{}

// NewToonEncoder creates a new TOON encoder.
// Deprecated: Use Writer.Write with format "toon" instead.
func NewToonEncoder() *ToonEncoder {
	return &ToonEncoder{}
}

// EncodeToday encodes a TodayOutput to TOON format.
// Deprecated: Use Writer.Write with format "toon" instead.
func (e *ToonEncoder) EncodeToday(output *forecast.TodayOutput) (string, error) {
	toonData := convertTodayToTOON(output)
	return toon.MarshalString(toonData, toon.WithIndent(2))
}

// EncodeDay encodes a DayOutput to TOON format.
// Deprecated: Use Writer.Write with format "toon" instead.
func (e *ToonEncoder) EncodeDay(output *forecast.DayOutput) (string, error) {
	toonData := convertDayToTOON(output)
	return toon.MarshalString(toonData, toon.WithIndent(2))
}

// EncodeWeek encodes a WeekOutput to TOON format.
// Deprecated: Use Writer.Write with format "toon" instead.
func (e *ToonEncoder) EncodeWeek(output *forecast.WeekOutput) (string, error) {
	toonData := convertWeekToTOON(output)
	return toon.MarshalString(toonData, toon.WithIndent(2))
}

// TOON structures for toon-go marshaling
// These structures are designed to produce TOON output compatible with the manual implementation

type toonMeta struct {
	GeneratedAt string  `toon:"generated_at"`
	Timezone    string  `toon:"timezone"`
	Latitude    float64 `toon:"latitude"`
	Longitude   float64 `toon:"longitude"`
}

type toonUnits struct {
	Temperature              string `toon:"temperature"`
	Humidity                 string `toon:"humidity"`
	WindSpeed                string `toon:"wind_speed"`
	WindDirection            string `toon:"wind_direction"`
	Precipitation            string `toon:"precipitation"`
	PrecipitationProbability string `toon:"precipitation_probability"`
	UVIndex                  string `toon:"uv_index"`
}

type toonCurrent struct {
	Time                     string  `toon:"time"`
	Weather                  string  `toon:"weather"`
	Temperature              float64 `toon:"temperature"`
	ApparentTemperature      float64 `toon:"apparent_temperature"`
	Humidity                 int     `toon:"humidity"`
	Precipitation            float64 `toon:"precipitation"`
	PrecipitationProbability int     `toon:"precipitation_probability"`
	WindSpeed                float64 `toon:"wind_speed"`
	WindGusts                float64 `toon:"wind_gusts"`
	WindDirection            int     `toon:"wind_direction"`
	UVIndex                  float64 `toon:"uv_index"`
}

type toonHour struct {
	Time                     string  `toon:"time"`
	Weather                  string  `toon:"weather"`
	Temperature              float64 `toon:"temperature"`
	ApparentTemperature      float64 `toon:"apparent_temperature"`
	Humidity                 int     `toon:"humidity"`
	Precipitation            float64 `toon:"precipitation"`
	PrecipitationProbability int     `toon:"precipitation_probability"`
	WindSpeed                float64 `toon:"wind_speed"`
	WindGusts                float64 `toon:"wind_gusts"`
	WindDirection            int     `toon:"wind_direction"`
	UVIndex                  float64 `toon:"uv_index"`
}

type toonDay struct {
	Date                        string  `toon:"date"`
	Weather                     string  `toon:"weather"`
	TempMin                     float64 `toon:"temp_min"`
	TempMax                     float64 `toon:"temp_max"`
	PrecipitationSum            float64 `toon:"precipitation_sum"`
	PrecipitationProbabilityMax int     `toon:"precipitation_probability_max"`
	WindSpeedMax                float64 `toon:"wind_speed_max"`
	WindGustsMax                float64 `toon:"wind_gusts_max"`
	UVIndexMax                  float64 `toon:"uv_index_max"`
	Sunrise                     string  `toon:"sunrise"`
	Sunset                      string  `toon:"sunset"`
}

type toonTodayOutput struct {
	Meta    toonMeta    `toon:"meta"`
	Units   toonUnits   `toon:"units"`
	Current toonCurrent `toon:"current"`
	Hours   []toonHour  `toon:"hours"`
}

type toonDayOutput struct {
	Meta  toonMeta   `toon:"meta"`
	Units toonUnits  `toon:"units"`
	Day   toonDay    `toon:"day"`
	Hours []toonHour `toon:"hours"`
}

type toonWeekOutput struct {
	Meta  toonMeta  `toon:"meta"`
	Units toonUnits `toon:"units"`
	Days  []toonDay `toon:"days"`
}

// convertTodayToTOON converts a TodayOutput to a toon-go marshallable structure.
func convertTodayToTOON(output *forecast.TodayOutput) toonTodayOutput {
	hours := make([]toonHour, len(output.Hours))
	for i, h := range output.Hours {
		hours[i] = toonHour{
			Time:                     h.Time,
			Weather:                  h.Weather,
			Temperature:              h.Temperature,
			ApparentTemperature:      h.ApparentTemperature,
			Humidity:                 h.Humidity,
			Precipitation:            h.Precipitation,
			PrecipitationProbability: h.PrecipitationProbability,
			WindSpeed:                h.WindSpeed,
			WindGusts:                h.WindGusts,
			WindDirection:            h.WindDirection,
			UVIndex:                  h.UVIndex,
		}
	}

	return toonTodayOutput{
		Meta: toonMeta{
			GeneratedAt: output.Meta.GeneratedAt.Format(time.RFC3339),
			Timezone:    output.Meta.Timezone,
			Latitude:    output.Meta.Latitude,
			Longitude:   output.Meta.Longitude,
		},
		Units: toonUnits{
			Temperature:              output.Meta.Units.Temperature,
			Humidity:                 output.Meta.Units.Humidity,
			WindSpeed:                output.Meta.Units.WindSpeed,
			WindDirection:            output.Meta.Units.WindDirection,
			Precipitation:            output.Meta.Units.Precipitation,
			PrecipitationProbability: output.Meta.Units.PrecipitationProbability,
			UVIndex:                  output.Meta.Units.UVIndex,
		},
		Current: toonCurrent{
			Time:                     output.Current.Time,
			Weather:                  output.Current.Weather,
			Temperature:              output.Current.Temperature,
			ApparentTemperature:      output.Current.ApparentTemperature,
			Humidity:                 output.Current.Humidity,
			Precipitation:            output.Current.Precipitation,
			PrecipitationProbability: output.Current.PrecipitationProbability,
			WindSpeed:                output.Current.WindSpeed,
			WindGusts:                output.Current.WindGusts,
			WindDirection:            output.Current.WindDirection,
			UVIndex:                  output.Current.UVIndex,
		},
		Hours: hours,
	}
}

// convertDayToTOON converts a DayOutput to a toon-go marshallable structure.
func convertDayToTOON(output *forecast.DayOutput) toonDayOutput {
	hours := make([]toonHour, len(output.Hours))
	for i, h := range output.Hours {
		hours[i] = toonHour{
			Time:                     h.Time,
			Weather:                  h.Weather,
			Temperature:              h.Temperature,
			ApparentTemperature:      h.ApparentTemperature,
			Humidity:                 h.Humidity,
			Precipitation:            h.Precipitation,
			PrecipitationProbability: h.PrecipitationProbability,
			WindSpeed:                h.WindSpeed,
			WindGusts:                h.WindGusts,
			WindDirection:            h.WindDirection,
			UVIndex:                  h.UVIndex,
		}
	}

	return toonDayOutput{
		Meta: toonMeta{
			GeneratedAt: output.Meta.GeneratedAt.Format(time.RFC3339),
			Timezone:    output.Meta.Timezone,
			Latitude:    output.Meta.Latitude,
			Longitude:   output.Meta.Longitude,
		},
		Units: toonUnits{
			Temperature:              output.Meta.Units.Temperature,
			Humidity:                 output.Meta.Units.Humidity,
			WindSpeed:                output.Meta.Units.WindSpeed,
			WindDirection:            output.Meta.Units.WindDirection,
			Precipitation:            output.Meta.Units.Precipitation,
			PrecipitationProbability: output.Meta.Units.PrecipitationProbability,
			UVIndex:                  output.Meta.Units.UVIndex,
		},
		Day: toonDay{
			Date:                        output.Day.Date,
			Weather:                     output.Day.Weather,
			TempMin:                     output.Day.TempMin,
			TempMax:                     output.Day.TempMax,
			PrecipitationSum:            output.Day.PrecipitationSum,
			PrecipitationProbabilityMax: output.Day.PrecipitationProbabilityMax,
			WindSpeedMax:                output.Day.WindSpeedMax,
			WindGustsMax:                output.Day.WindGustsMax,
			UVIndexMax:                  output.Day.UVIndexMax,
			Sunrise:                     output.Day.Sunrise,
			Sunset:                      output.Day.Sunset,
		},
		Hours: hours,
	}
}

// convertWeekToTOON converts a WeekOutput to a toon-go marshallable structure.
func convertWeekToTOON(output *forecast.WeekOutput) toonWeekOutput {
	days := make([]toonDay, len(output.Days))
	for i, d := range output.Days {
		days[i] = toonDay{
			Date:                        d.Date,
			Weather:                     d.Weather,
			TempMin:                     d.TempMin,
			TempMax:                     d.TempMax,
			PrecipitationSum:            d.PrecipitationSum,
			PrecipitationProbabilityMax: d.PrecipitationProbabilityMax,
			WindSpeedMax:                d.WindSpeedMax,
			WindGustsMax:                d.WindGustsMax,
			UVIndexMax:                  d.UVIndexMax,
			Sunrise:                     d.Sunrise,
			Sunset:                      d.Sunset,
		}
	}

	return toonWeekOutput{
		Meta: toonMeta{
			GeneratedAt: output.Meta.GeneratedAt.Format(time.RFC3339),
			Timezone:    output.Meta.Timezone,
			Latitude:    output.Meta.Latitude,
			Longitude:   output.Meta.Longitude,
		},
		Units: toonUnits{
			Temperature:              output.Meta.Units.Temperature,
			Humidity:                 output.Meta.Units.Humidity,
			WindSpeed:                output.Meta.Units.WindSpeed,
			WindDirection:            output.Meta.Units.WindDirection,
			Precipitation:            output.Meta.Units.Precipitation,
			PrecipitationProbability: output.Meta.Units.PrecipitationProbability,
			UVIndex:                  output.Meta.Units.UVIndex,
		},
		Days: days,
	}
}
