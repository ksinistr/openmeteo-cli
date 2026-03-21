// Package output provides output formatting and writing utilities.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/toon-format/toon-go"
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
