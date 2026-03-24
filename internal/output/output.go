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
	switch d := data.(type) {
	case *forecast.HourlyOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil HourlyOutput")
		}
	case *forecast.DailyOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil DailyOutput")
		}
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// writeTOON encodes the data as TOON format using toon-go library.
func writeTOON(data interface{}, out io.Writer) error {
	var toonData interface{}
	var err error

	switch d := data.(type) {
	case *forecast.HourlyOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil HourlyOutput")
		}
		toonData = convertHourlyToTOON(d)
	case *forecast.DailyOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil DailyOutput")
		}
		toonData = convertDailyToTOON(d)
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

// EncodeHourlyTo encodes a HourlyOutput to JSON format.
// Deprecated: Use Writer.Write with format "json" instead.
func (e *JSONEncoder) EncodeHourlyTo(output *forecast.HourlyOutput, out io.Writer) error {
	if output == nil {
		return fmt.Errorf("cannot write nil HourlyOutput")
	}
	return writeJSON(output, out)
}

// EncodeDailyTo encodes a DailyOutput to JSON format.
// Deprecated: Use Writer.Write with format "json" instead.
func (e *JSONEncoder) EncodeDailyTo(output *forecast.DailyOutput, out io.Writer) error {
	if output == nil {
		return fmt.Errorf("cannot write nil DailyOutput")
	}
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

// EncodeHourly encodes a HourlyOutput to TOON format.
// Deprecated: Use Writer.Write with format "toon" instead.
func (e *ToonEncoder) EncodeHourly(output *forecast.HourlyOutput) (string, error) {
	if output == nil {
		return "", fmt.Errorf("cannot write nil HourlyOutput")
	}
	toonData := convertHourlyToTOON(output)
	return toon.MarshalString(toonData, toon.WithIndent(2))
}

// EncodeDaily encodes a DailyOutput to TOON format.
// Deprecated: Use Writer.Write with format "toon" instead.
func (e *ToonEncoder) EncodeDaily(output *forecast.DailyOutput) (string, error) {
	if output == nil {
		return "", fmt.Errorf("cannot write nil DailyOutput")
	}
	toonData := convertDailyToTOON(output)
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

type toonHourlyOutput struct {
	Meta  toonMeta   `toon:"meta"`
	Units toonUnits  `toon:"units"`
	Hours []toonHour `toon:"hours"`
}

type toonDailyOutput struct {
	Meta  toonMeta  `toon:"meta"`
	Units toonUnits `toon:"units"`
	Days  []toonDay `toon:"days"`
}

// convertHourlyToTOON converts a HourlyOutput to a toon-go marshallable structure.
func convertHourlyToTOON(output *forecast.HourlyOutput) toonHourlyOutput {
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

	return toonHourlyOutput{
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
		Hours: hours,
	}
}

// convertDailyToTOON converts a DailyOutput to a toon-go marshallable structure.
func convertDailyToTOON(output *forecast.DailyOutput) toonDailyOutput {
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

	return toonDailyOutput{
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
