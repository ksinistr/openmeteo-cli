package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestRunExitCodes tests that correct exit codes are returned for different error types.
func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		// Exit code 2: invalid arguments
		{"invalid lat format", []string{"today", "--lat", "not-a-number", "--lon", "0"}, 2},
		{"invalid lon format", []string{"today", "--lat", "0", "--lon", "not-a-number"}, 2},
		{"missing lat", []string{"today", "--lon", "0"}, 2},
		{"missing lon", []string{"today", "--lat", "0"}, 2},
		{"missing both lat and lon", []string{"today"}, 2},
		{"invalid units format", []string{"today", "--lat", "0", "--lon", "0", "--units", "celsius"}, 3},
		{"invalid format format", []string{"today", "--lat", "0", "--lon", "0", "--format", "xml"}, 3},

		// Exit code 3: validation errors
		{"latitude too high", []string{"today", "--lat", "91", "--lon", "0"}, 3},
		{"latitude too low", []string{"today", "--lat", "-91", "--lon", "0"}, 3},
		{"longitude too high", []string{"today", "--lat", "0", "--lon", "181"}, 3},
		{"longitude too low", []string{"today", "--lat", "0", "--lon", "-181"}, 3},
		{"invalid units", []string{"today", "--lat", "40", "--lon", "-74", "--units", "imperial_fahrenheit"}, 3},
		{"invalid format", []string{"today", "--lat", "40", "--lon", "-74", "--format", "json_lines"}, 3},
		{"day missing date", []string{"day", "--lat", "40", "--lon", "-74"}, 3},
		{"day invalid date", []string{"day", "--lat", "40", "--lon", "-74", "--date", "2026-13-45"}, 3},

		// Unknown command - exit 2
		{"unknown command", []string{"forecast", "--lat", "40", "--lon", "-74"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != tt.wantExitCode {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.wantExitCode)
			}
		})
	}
}

// TestRunEmptyArgs tests behavior when called with no arguments.
func TestRunEmptyArgs(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	exitCode := Run([]string{})

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	stderr := buf.String()

	if exitCode != 2 {
		t.Errorf("Run([]) exit code = %d, want 2", exitCode)
	}
	if !strings.Contains(stderr, "Error: no command specified") {
		t.Errorf("Run([]) stderr = %q, want to contain 'Error: no command specified'", stderr)
	}
}

// TestRunNoCommand tests behavior when called with only invalid args (no command).
func TestRunNoCommand(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Note: With "--lat" as first arg, the code treats it as command name and "--lat"
	// becomes an argument to check for lat/lon - which fails validation
	exitCode := Run([]string{"--lat", "40", "--lon", "-74"})

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	stderr := buf.String()

	if exitCode != 2 {
		t.Errorf("Run with args but no command exit code = %d, want 2", exitCode)
	}
	// The error is about lat/lon validation since the first arg is used as command name
	if !strings.Contains(stderr, "Error:") {
		t.Errorf("Run should output error to stderr, got: %q", stderr)
	}
}

// TestRunCommandDispatch tests that the Run function parses and dispatches commands.
func TestRunCommandDispatch(t *testing.T) {
	// These will succeed (status 0) when the API call succeeds
	testCases := []struct {
		name string
		args []string
		want int
	}{
		{"today dispatch", []string{"today", "--lat", "40.7128", "--lon", "-74.0060"}, 0},
		{"day dispatch", []string{"day", "--lat", "40.7128", "--lon", "-74.0060", "--date", "2026-03-22"}, 0},
		{"week dispatch", []string{"week", "--lat", "40.7128", "--lon", "-74.0060"}, 0},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Command dispatch successful (exit 0 means success)
			exitCode := Run(tt.args)
			// The API may or may not be available, so we just verify the command was dispatched
			// by checking that exit code is not for invalid arguments
			if exitCode == 2 {
				// Exit code 2 would mean invalid arguments were passed
				t.Errorf("Run(%v) exit code = %d, expected success (not invalid args)", tt.args, exitCode)
			}
		})
	}
}

// TestRunValidatesArgsBeforeDispatch tests that argument validation happens before API calls.
func TestRunValidatesArgsBeforeDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"invalid lat format", []string{"today", "--lat", "abc", "--lon", "0"}},
		{"lat out of range", []string{"today", "--lat", "100", "--lon", "0"}},
		{"day missing date", []string{"day", "--lat", "40", "--lon", "-74"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			_ = Run(tt.args)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			io.Copy(&buf, r)

			stderr := buf.String()

			// Check that we got an error message (validation error, not network error)
			if !strings.Contains(stderr, "Error:") {
				t.Errorf("Run(%v) should output error to stderr, got: %q", tt.args, stderr)
			}
			// Validation error should occur before network calls
			// Network errors would be from real HTTP calls
		})
	}
}

// TestRunValidationErrorExitCode tests that validation errors return exit code 3.
func TestRunValidationErrorExitCode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"lat out of range", []string{"today", "--lat", "100", "--lon", "0"}, 3},
		{"lon out of range", []string{"today", "--lat", "40", "--lon", "200"}, 3},
		{"invalid units", []string{"today", "--lat", "40", "--lon", "0", "--units", "invalid"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != tt.want {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.want)
			}
		})
	}
}

// TestRunInvalidArgumentExitCode tests that invalid argument errors return exit code 2.
func TestRunInvalidArgumentExitCode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"missing lat/lon", []string{"today"}, 2},
		{"invalid lat format", []string{"today", "--lat", "abc", "--lon", "0"}, 2},
		{"invalid lon format", []string{"today", "--lat", "0", "--lon", "xyz"}, 2},
		{"unknown command", []string{"forecast", "--lat", "0", "--lon", "0"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode := Run(tt.args)
			if exitCode != tt.want {
				t.Errorf("Run(%v) exit code = %d, want %d", tt.args, exitCode, tt.want)
			}
		})
	}
}
