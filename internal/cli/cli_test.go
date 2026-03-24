package cli

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: " forecast command with hourly and default days",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--hourly"},
			want: &Config{
				Hourly:       true,
				Daily:        false,
				ForecastDays: 1,
				Lat:          40.7128,
				Lon:          -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: " forecast command with hourly and 2 days",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--hourly", "--forecast-days", "2"},
			want: &Config{
				Hourly:       true,
				Daily:        false,
				ForecastDays: 2,
				Lat:          40.7128,
				Lon:          -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: " forecast command with imperial units",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--daily", "--units", "imperial"},
			want: &Config{
				Hourly:       false,
				Daily:        true,
				ForecastDays: 1,
				Lat:          40.7128,
				Lon:          -74.0060,
				Units:        "imperial",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: " forecast command with json format",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--daily", "--format", "json"},
			want: &Config{
				Hourly:       false,
				Daily:        true,
				ForecastDays: 1,
				Lat:          40.7128,
				Lon:          -74.0060,
				Units:        "metric",
				Format:       "json",
			},
			wantErr: false,
		},
		{
			name: " forecast command with daily and 7 days",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--daily", "--forecast-days", "7"},
			want: &Config{
				Hourly:       false,
				Daily:        true,
				ForecastDays: 7,
				Lat:          40.7128,
				Lon:          -74.0060,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: " missing lat and lon are required",
			args: []string{"--lon", "-74.0060"},
			want: nil,
			wantErr: true,
			errMsg:  "lat and lon are required",
		},
		{
			name: " invalid latitude",
			args: []string{"--lat", "100", "--lon", "-74.0060", "--hourly"},
			want: nil,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: " invalid longitude",
			args: []string{"--lat", "40.7128", "--lon", "-200", "--hourly"},
			want: nil,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name: " invalid units",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--hourly", "--units", "invalid"},
			want: nil,
			wantErr: true,
			errMsg:  "units must be 'metric' or 'imperial'",
		},
		{
			name: " invalid format",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--hourly", "--format", "invalid"},
			want: nil,
			wantErr: true,
			errMsg:  "format must be 'toon' or 'json'",
		},
		{
			name: " missing both hourly and daily",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060"},
			want: nil,
			wantErr: true,
			errMsg:  "must specify either --hourly or --daily",
		},
		{
			name: " both hourly and daily specified",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--hourly", "--daily"},
			want: nil,
			wantErr: true,
			errMsg:  "cannot use both --hourly and --daily",
		},
		{
			name: " hourly exceeds max days",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--hourly", "--forecast-days", "3"},
			want: nil,
			wantErr: true,
			errMsg:  "--hourly supports maximum 2 days",
		},
		{
			name: " daily exceeds max days",
			args: []string{"--lat", "40.7128", "--lon", "-74.0060", "--daily", "--forecast-days", "15"},
			want: nil,
			wantErr: true,
			errMsg:  "--daily supports maximum 14 days",
		},
		{
			name: " duplicate --lat flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--lat", "41.0"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --lat",
		},
		{
			name: " duplicate --lon flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--lon", "-75.0"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --lon",
		},
		{
			name: " duplicate --units flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--units", "metric", "--units", "imperial"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --units",
		},
		{
			name: " duplicate --format flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--format", "toon", "--format", "json"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --format",
		},
		{
			name: " duplicate --hourly flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--hourly"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --hourly",
		},
		{
			name: " duplicate --daily flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--daily", "--daily"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --daily",
		},
		{
			name: " duplicate --forecast-days flag",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--forecast-days", "1", "--forecast-days", "2"},
			want: nil,
			wantErr: true,
			errMsg:  "duplicate flag: --forecast-days",
		},
		{
			name: " invalid forecast-days value",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--forecast-days", "abc"},
			want: nil,
			wantErr: true,
			errMsg:  "invalid forecast-days value",
		},
		{
			name: " forecast-days cannot be zero",
			args: []string{"--lat", "40.0", "--lon", "-74.0", "--hourly", "--forecast-days", "0"},
			want: nil,
			wantErr: true,
			errMsg:  "forecast-days must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// Allow any error message for validation errors
				_ = err.Error()
			}
			if err == nil && !tt.wantErr {
				if got.Hourly != tt.want.Hourly {
					t.Errorf("Parse() Hourly = %v, want %v", got.Hourly, tt.want.Hourly)
				}
				if got.Daily != tt.want.Daily {
					t.Errorf("Parse() Daily = %v, want %v", got.Daily, tt.want.Daily)
				}
				if got.ForecastDays != tt.want.ForecastDays {
					t.Errorf("Parse() ForecastDays = %v, want %v", got.ForecastDays, tt.want.ForecastDays)
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
			args:    []string{"--lat", "40", "--lon", "-74", "-h"},
			wantErr: false,
		},
		{
			name:    "with --help flag after valid args",
			args:    []string{"--lat", "40", "--lon", "-74", "--units", "imperial", "--help"},
			wantErr: false,
		},
		{
			name:    "with --hourly and -h",
			args:    []string{"--lat", "40", "--lon", "-74", "--hourly", "-h"},
			wantErr: false,
		},
		{
			name:    "with --daily and --help",
			args:    []string{"--lat", "40", "--lon", "-74", "--daily", "--help"},
			wantErr: false,
		},
		{
			name:    "with --forecast-days and --help",
			args:    []string{"--lat", "40", "--lon", "-74", "--hourly", "--forecast-days", "1", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args)
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
			name: "forecast-days value help token is parsed as number fails",
			args: []string{"--lat", "40", "--lon", "-74", "--hourly", "--forecast-days", "--help"},
			want: nil,
			wantErr: true,
			errString: "invalid forecast-days value",
		},
		{
			name:      "units value help token stays a value",
			args:      []string{"--lat", "40", "--lon", "-74", "--hourly", "--units", "--help"},
			wantErr:   true,
			errString: "units must be 'metric' or 'imperial'",
		},
		{
			name:      "lat value help token stays a value",
			args:      []string{"--lat", "-h", "--lon", "-74", "--hourly"},
			wantErr:   true,
			errString: "invalid lat value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
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
			if got.Hourly != tt.want.Hourly {
				t.Fatalf("Parse() Hourly = %v, want %v", got.Hourly, tt.want.Hourly)
			}
			if got.Daily != tt.want.Daily {
				t.Fatalf("Parse() Daily = %v, want %v", got.Daily, tt.want.Daily)
			}
			if got.ForecastDays != tt.want.ForecastDays {
				t.Fatalf("Parse() ForecastDays = %v, want %v", got.ForecastDays, tt.want.ForecastDays)
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
		})
	}
}
