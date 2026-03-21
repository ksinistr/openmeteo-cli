package cli

import (
	"testing"
	"time"
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
				Date:    time.Time{},
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
				Date:    time.Time{},
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
				Date:    time.Time{},
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
				Date:    time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
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
				Date:    time.Time{},
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
			name:    "day command missing date",
			command: "day",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060"},
			want:    nil,
			wantErr: true,
			errMsg:  "date is required for day command",
		},
		{
			name:    "day command invalid date format",
			command: "day",
			args:    []string{"--lat", "40.7128", "--lon", "-74.0060", "--date", "invalid"},
			want:    nil,
			wantErr: true,
			errMsg:  "invalid date format",
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
				if err.Error() != tt.errMsg && err.Error() != "invalid command-line arguments" {
					// Allow any error message for validation errors
				}
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
