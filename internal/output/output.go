// Package output provides output formatting and writing utilities.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

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

// SetOutput sets the output writer (used by tests).
func (w *Writer) SetOutput(out io.Writer) {
	w.out = out
}

// SetError sets the error output writer (used by tests).
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
	case *forecast.ForecastOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil ForecastOutput")
		}
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// writeTOON encodes the data as TOON format using toon-go library.
func writeTOON(data interface{}, out io.Writer) error {
	switch d := data.(type) {
	case *forecast.HourlyOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil HourlyOutput")
		}
	case *forecast.DailyOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil DailyOutput")
		}
	case *forecast.ForecastOutput:
		if d == nil {
			return fmt.Errorf("cannot write nil ForecastOutput")
		}
		// Convert custom types to toon.Object for proper field ordering
		data = convertForecastOutputForTOON(d)
	}
	// Use toon-go to marshal the data directly
	// time.Time is automatically formatted as RFC3339 by toon-go
	output, err := toon.MarshalString(data, toon.WithIndent(2))
	if err != nil {
		return err
	}

	_, err = out.Write([]byte(output))
	return err
}

// convertForecastOutputForTOON converts ForecastOutput to use toon.Object for proper field ordering.
func convertForecastOutputForTOON(fo *forecast.ForecastOutput) toon.Object {
	fields := []toon.Field{
		{Key: "meta", Value: fo.Meta},
	}

	if fo.Current != nil {
		fields = append(fields, toon.Field{Key: "current", Value: fo.Current.ToTOON()})
	}
	if fo.Hourly != nil {
		fields = append(fields, toon.Field{Key: "hourly", Value: fo.Hourly.ToTOON()})
	}
	if fo.Daily != nil {
		fields = append(fields, toon.Field{Key: "daily", Value: fo.Daily.ToTOON()})
	}

	return toon.Object{Fields: fields}
}
