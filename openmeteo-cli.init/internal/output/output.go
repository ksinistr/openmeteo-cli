// Package output provides output formatting and writing utilities.
package output

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Writer handles writing responses to stdout/stderr.
type Writer struct {
	buf bytes.Buffer
}

// NewWriter creates a new output writer.
func NewWriter() *Writer {
	return &Writer{}
}

// Write outputs the result in the specified format.
func (w *Writer) Write(data interface{}, format string) error {
	var buf bytes.Buffer
	var err error

	switch format {
	case "json":
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		err = enc.Encode(data)
	default:
		err = errors.New("unknown format")
	}

	if err != nil {
		return err
	}

	// Print directly
	// We return the error for the app to handle output
	return nil
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
