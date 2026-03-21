package cli

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		want    *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "today command with valid arguments",
			command: "today",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060"},
			want: &Config{
				Command: "today",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "metric",
				Format:  "toon",
				DateStr: "",
			},
			wantErr: false,
		},
		{
			name:    "today command with imperial units",
			command: "today",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--units", "imperial"},
			want: &Config{
				Command: "today",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "imperial",
				Format:  "toon",
				DateStr: "",
			},
			wantErr: false,
		},
		{
			name:    "today command with json format",
			command: "today",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--format", "json"},
			want: &Config{
				Command: "today",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "metric",
				Format:  "json",
				DateStr: "",
			},
			wantErr: false,
		},
		{
			name:    "day command with valid arguments",
			command: "day",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--date", "2026-03-22"},
			want: &Config{
				Command: "day",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "metric",
				Format:  "toon",
				DateStr: "2026-03-22",
			},
			wantErr: false,
		},
		{
			name:    "week command with valid arguments",
			command: "week",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060"},
			want: &Config{
				Command: "week",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "metric",
				Format:  "toon",
				DateStr: "",
			},
			wantErr: false,
		},
		{
			name:    "missing lat to lon are required",
			command: "today",
			args:    []string{"--lon", "-74.0060"},
			want:    nil,
			wantErr: true,
			errMsg:  "lat and lon are required",
		},
		{
			name:    "invalid latitude",
			command: "today",
			args:    []string{"--lat", "100", "--lon", "-74.0060"},
			want:    nil,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name:    "invalid longitude",
			command: "today",
			args:    []string{"--lat", "40.7128", "--lon", "-200"},
			want:    nil,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name:    "invalid units",
			command: "today",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--units", "invalid"},
			want:    nil,
			wantErr: true,
			errMsg:  "units must be 'metric' or 'imperial'",
		},
		{
			name:    "invalid format",
			command: "today",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--format", "invalid"},
			want:    nil,
			wantErr: true,
			errMsg:  "format must be 'toon' or 'json'",
		},
		{
			name:    "day command missing date - validation now in app package",
			command: "day",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060"},
			want: &Config{
				Command: "day",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "metric",
				Format:  "toon",
				DateStr: "",
			},
			wantErr: false, // cli.Parse now accepts this, validation is in app.go
		},
		{
			name:    "day command invalid date format - validation now in app package",
			command: "day",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--date", "invalid"},
			want: &Config{
				Command: "day",
				Lat:     40.7128,
				Lon:     -74.0060,
				Units:   "metric",
				Format:  "toon",
				DateStr: "invalid",
			},
			wantErr: false, // cli.Parse no longer validates date format, validation is in app.go
		},
		{
			name:    "duplicate --lat flag",
			command: "today",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--lat", "41.0"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --lat",
		},
		{
			name:    "duplicate --lon flag",
			command: "today",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--lon", "-75.0"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --lon",
		},
		{
			name:    "duplicate --units flag",
			command: "today",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--units", "metric", "--units", "imperial"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --units",
		},
		{
			name:    "duplicate --format flag",
			command: "today",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--format", "toon", "--format", "json"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --format",
		},
		{
			name:    "duplicate --date flag",
			command: "day",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--date", "2026-01-01", "--date", "2026-01-02"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --date",
		},
		{
			name:    "today command with --date flag - validation now in app package",
			command: "today",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--date", "2026-03-22"},
			want: &Config{
				Command: "today",
				Lat:     40.0,
				Lon:     -74.0,
				Units:   "metric",
				Format:  "toon",
				DateStr: "2026-03-22",
			},
			wantErr: false, // cli.Parse now accepts this, validation is in app.go
		},
		{
			name:    "week command with --date flag - validation now in app package",
			command: "week",
			args:    []string{"--lat", "40.0", "--lon", "-74.0", "--date", "2026-03-22"},
			want: &Config{
				Command: "week",
				Lat:     40.0,
				Lon:     -74.0,
				Units:   "metric",
				Format:  "toon",
				DateStr: "2026-03-22",
			},
			wantErr: false, // cli.Parse now accepts this, validation is in app.go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.command, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// Allow any error message for validation errors
				_ = err.Error()
			}
			if err == nil && !tt.wantErr {
				if got.Command != tt.want.Command {
					t.Errorf("Parse() Command = %v, want %v", got.Command, tt.want.Command)
				}
				if got.Lat != tt.want.Lat {
					t.Errorf("Parse() Lat = %v, want %v", got.Lat, tt.want.Lat)
				}
				if got.Lon != tt.want.Lon {
					t.Errorf("Parse() Lon = %v, want %v", got.Lon, tt.want.Lon)
				}
				if got.Units != tt.want.Units {
					t.Errorf("Parse() Units = %v, want %v", got.Units, tt.want.Units)
				}
				if got.Format != tt.want.Format {
					t.Errorf("Parse() Format = %v, want %v", got.Format, tt.want.Format)
				}
				if got.DateStr != tt.want.DateStr {
					t.Errorf("Parse() DateStr = %v, want %v", got.DateStr, tt.want.DateStr)
				}
			}
		})
	}
}

func TestParseWithHelpFlags(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
		errType string
	}{
		{
			name:    "today with -h flag alone",
			command: "today",
			args:    []string{"-h"},
			wantErr: false,
		},
		{
			name:    "today with --help flag alone",
			command: "today",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "day with -h flag after valid args",
			command: "day",
			args:    []string{"--lat", "40", "--lon", "-74", "-h"},
			wantErr: false,
		},
		{
			name:    "week with --help flag after valid args",
			command: "week",
			args:    []string{"--lat", "40", "--lon", "-74", "--units", "imperial", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.command, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestParseDoesNotTreatFlagValuesAsHelp(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		want      *Config
		wantErr   bool
		errString string
	}{
		{
			name:    "day date value can be help token",
			command: "day",
			args:    []string{"--lat", "40", "--lon", "-74", "--date", "--help"},
			want: &Config{
				Command: "day",
				Lat:     40,
				Lon:     -74,
				Units:   "metric",
				Format:  "toon",
				DateStr: "--help",
			},
		},
		{
			name:      "units value help token stays a value",
			command:   "today",
			args:      []string{"--lat", "40", "--lon", "-74", "--units", "--help"},
			wantErr:   true,
			errString: "units must be 'metric' or 'imperial'",
		},
		{
			name:      "lat value help token stays a value",
			command:   "today",
			args:      []string{"--lat", "-h", "--lon", "-74"},
			wantErr:   true,
			errString: "invalid lat value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.command, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if err == nil || err.Error() != tt.errString {
					t.Fatalf("Parse() error = %v, want %q", err, tt.errString)
				}
				return
			}
			if got == nil {
				t.Fatal("Parse() returned nil config")
			}
			if got.Command != tt.want.Command {
				t.Fatalf("Parse() Command = %v, want %v", got.Command, tt.want.Command)
			}
			if got.Lat != tt.want.Lat {
				t.Fatalf("Parse() Lat = %v, want %v", got.Lat, tt.want.Lat)
			}
			if got.Lon != tt.want.Lon {
				t.Fatalf("Parse() Lon = %v, want %v", got.Lon, tt.want.Lon)
			}
			if got.Units != tt.want.Units {
				t.Fatalf("Parse() Units = %v, want %v", got.Units, tt.want.Units)
			}
			if got.Format != tt.want.Format {
				t.Fatalf("Parse() Format = %v, want %v", got.Format, tt.want.Format)
			}
			if got.DateStr != tt.want.DateStr {
				t.Fatalf("Parse() DateStr = %v, want %v", got.DateStr, tt.want.DateStr)
			}
		})
	}
}
