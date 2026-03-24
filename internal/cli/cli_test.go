package cli

import (
	"testing"
)

func TestParseHourly(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "hourly command with 1 day",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "1"},
			want: &Config{
				Mode:         "hourly",
				ForecastDays: 1,
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly command with 2 days",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "2"},
			want: &Config{
				Mode:         "hourly",
				ForecastDays: 2,
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly command with imperial units",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "1", "--units", "imperial"},
			want: &Config{
				Mode:         "hourly",
				ForecastDays: 1,
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Units:        "imperial",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly command with json format",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "1", "--format", "json"},
			want: &Config{
				Mode:         "hourly",
				ForecastDays: 1,
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Units:        "metric",
				Format:       "json",
			},
			wantErr: false,
		},
		{
			name:    "missing latitude",
			args:    []string{"--longitude", "-74.0060", "--forecast-days", "1"},
			want:    nil,
			wantErr: true,
			errMsg:  "--latitude and --longitude are required",
		},
		{
			name:    "missing longitude",
			args:    []string{"--latitude", "40.7128", "--forecast-days", "1"},
			want:    nil,
			wantErr: true,
			errMsg:  "--latitude and --longitude are required",
		},
		{
			name:    "invalid latitude",
			args:    []string{"--latitude", "100", "--longitude", "-74.0060", "--forecast-days", "1"},
			want:    nil,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name:    "invalid longitude",
			args:    []string{"--latitude", "40.7128", "--longitude", "-200", "--forecast-days", "1"},
			want:    nil,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name:    "missing forecast-days",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-days is required",
		},
		{
			name:    "hourly exceeds max days",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "3"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-days must be between 1 and 2 for hourly forecast",
		},
		{
			name:    "invalid units",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "1", "--units", "invalid"},
			want:    nil,
			wantErr: true,
			errMsg:  "units must be 'metric' or 'imperial'",
		},
		{
			name:    "invalid format",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "1", "--format", "invalid"},
			want:    nil,
			wantErr: true,
			errMsg:  "format must be 'toon' or 'json'",
		},
		{
			name:    "duplicate --latitude flag",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "1", "--latitude", "41.0"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --latitude",
		},
		{
			name:    "duplicate --longitude flag",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "1", "--longitude", "-75.0"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --longitude",
		},
		{
			name:    "duplicate --units flag",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "1", "--units", "metric", "--units", "imperial"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --units",
		},
		{
			name:    "duplicate --format flag",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "1", "--format", "toon", "--format", "json"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --format",
		},
		{
			name:    "duplicate --forecast-days flag",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "1", "--forecast-days", "2"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --forecast-days",
		},
		{
			name:    "invalid forecast-days value",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "abc"},
			want:    nil,
			wantErr: true,
			errMsg:  "invalid forecast-days value",
		},
		{
			name:    "forecast-days cannot be zero",
			args:    []string{"--latitude", "40.0", "--longitude", "-74.0", "--forecast-days", "0"},
			want:    nil,
			wantErr: true,
			errMsg:  "forecast-days must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args, "hourly")
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Parse() error message = %q, want %q", err.Error(), tt.errMsg)
				}
			}
			if err == nil && !tt.wantErr {
				if got.Mode != tt.want.Mode {
					t.Errorf("Parse() Mode = %v, want %v", got.Mode, tt.want.Mode)
				}
				if got.ForecastDays != tt.want.ForecastDays {
					t.Errorf("Parse() ForecastDays = %v, want %v", got.ForecastDays, tt.want.ForecastDays)
				}
				if got.Latitude != tt.want.Latitude {
					t.Errorf("Parse() Latitude = %v, want %v", got.Latitude, tt.want.Latitude)
				}
				if got.Longitude != tt.want.Longitude {
					t.Errorf("Parse() Longitude = %v, want %v", got.Longitude, tt.want.Longitude)
				}
				if got.Units != tt.want.Units {
					t.Errorf("Parse() Units = %v, want %v", got.Units, tt.want.Units)
				}
				if got.Format != tt.want.Format {
					t.Errorf("Parse() Format = %v, want %v", got.Format, tt.want.Format)
				}
			}
		})
	}
}

func TestParseDaily(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "daily command with 7 days",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "7"},
			want: &Config{
				Mode:         "daily",
				ForecastDays: 7,
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "daily command with 14 days",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "14"},
			want: &Config{
				Mode:         "daily",
				ForecastDays: 14,
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name:    "daily exceeds max days",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--forecast-days", "15"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-days must be between 1 and 14 for daily forecast",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args, "daily")
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Parse() error message = %q, want %q", err.Error(), tt.errMsg)
				}
			}
			if err == nil && !tt.wantErr {
				if got.Mode != tt.want.Mode {
					t.Errorf("Parse() Mode = %v, want %v", got.Mode, tt.want.Mode)
				}
				if got.ForecastDays != tt.want.ForecastDays {
					t.Errorf("Parse() ForecastDays = %v, want %v", got.ForecastDays, tt.want.ForecastDays)
				}
				if got.Latitude != tt.want.Latitude {
					t.Errorf("Parse() Latitude = %v, want %v", got.Latitude, tt.want.Latitude)
				}
				if got.Longitude != tt.want.Longitude {
					t.Errorf("Parse() Longitude = %v, want %v", got.Longitude, tt.want.Longitude)
				}
				if got.Units != tt.want.Units {
					t.Errorf("Parse() Units = %v, want %v", got.Units, tt.want.Units)
				}
				if got.Format != tt.want.Format {
					t.Errorf("Parse() Format = %v, want %v", got.Format, tt.want.Format)
				}
			}
		})
	}
}

func TestParseWithHelpFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errType string
	}{
		{
			name:    "with -h flag alone",
			args:    []string{"-h"},
			wantErr: false,
		},
		{
			name:    "with --help flag alone",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "with -h flag after valid args",
			args:    []string{"--latitude", "40", "--longitude", "-74", "-h"},
			wantErr: false,
		},
		{
			name:    "with --help flag after valid args",
			args:    []string{"--latitude", "40", "--longitude", "-74", "--units", "imperial", "--help"},
			wantErr: false,
		},
		{
			name:    "with --forecast-days and --help",
			args:    []string{"--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args, "hourly")
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
		args      []string
		want      *Config
		wantErr   bool
		errString string
	}{
		{
			name:      "forecast-days value help token fails",
			args:      []string{"--latitude", "40", "--longitude", "-74", "--forecast-days", "--help"},
			want:      nil,
			wantErr:   true,
			errString: "invalid forecast-days value",
		},
		{
			name:      "units value help token stays a value",
			args:      []string{"--latitude", "40", "--longitude", "-74", "--forecast-days", "1", "--units", "--help"},
			wantErr:   true,
			errString: "units must be 'metric' or 'imperial'",
		},
		{
			name:      "latitude value help token stays a value",
			args:      []string{"--latitude", "-h", "--longitude", "-74", "--forecast-days", "1"},
			wantErr:   true,
			errString: "invalid latitude value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args, "hourly")
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
			if got.Mode != tt.want.Mode {
				t.Fatalf("Parse() Mode = %v, want %v", got.Mode, tt.want.Mode)
			}
			if got.ForecastDays != tt.want.ForecastDays {
				t.Fatalf("Parse() ForecastDays = %v, want %v", got.ForecastDays, tt.want.ForecastDays)
			}
			if got.Latitude != tt.want.Latitude {
				t.Fatalf("Parse() Latitude = %v, want %v", got.Latitude, tt.want.Latitude)
			}
			if got.Longitude != tt.want.Longitude {
				t.Fatalf("Parse() Longitude = %v, want %v", got.Longitude, tt.want.Longitude)
			}
			if got.Units != tt.want.Units {
				t.Fatalf("Parse() Units = %v, want %v", got.Units, tt.want.Units)
			}
			if got.Format != tt.want.Format {
				t.Fatalf("Parse() Format = %v, want %v", got.Format, tt.want.Format)
			}
		})
	}
}
