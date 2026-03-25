package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseForecast(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *ForecastConfig
		wantErr bool
		errMsg  string
	}{
		// Valid current-only requests
		{
			name: "current-only with coordinates",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m,weather_code"},
			want: &ForecastConfig{
				Latitude:    40.7128,
				Longitude:   -74.0060,
				CurrentVars: []string{"temperature_2m", "weather_code"},
				Units:       "metric",
				Format:      "toon",
			},
			wantErr: false,
		},
		{
			name: "current-only with default",
			args: []string{"--latitude", "52.52", "--longitude", "13.41", "--current", "default"},
			want: &ForecastConfig{
				Latitude:    52.52,
				Longitude:   13.41,
				CurrentVars: []string{"default"},
				Units:       "metric",
				Format:      "toon",
			},
			wantErr: false,
		},

		// Valid hourly-only requests
		{
			name: "hourly-only with coordinates and forecast-hours",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m,precipitation_probability", "--forecast-hours", "24"},
			want: &ForecastConfig{
				Latitude:      40.7128,
				Longitude:     -74.0060,
				HourlyVars:    []string{"temperature_2m", "precipitation_probability"},
				ForecastHours: 24,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly-only with max forecast-hours",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "default", "--forecast-hours", "48"},
			want: &ForecastConfig{
				Latitude:      40.7128,
				Longitude:     -74.0060,
				HourlyVars:    []string{"default"},
				ForecastHours: 48,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly-only with min forecast-hours",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "1"},
			want: &ForecastConfig{
				Latitude:      40.7128,
				Longitude:     -74.0060,
				HourlyVars:    []string{"temperature_2m"},
				ForecastHours: 1,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},

		// Valid daily-only requests
		{
			name: "daily-only with coordinates and forecast-days",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code,temperature_2m_max,temperature_2m_min", "--forecast-days", "7"},
			want: &ForecastConfig{
				Latitude:     40.7128,
				Longitude:    -74.0060,
				DailyVars:    []string{"weather_code", "temperature_2m_max", "temperature_2m_min"},
				ForecastDays: 7,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "daily-only with max forecast-days",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "default", "--forecast-days", "14"},
			want: &ForecastConfig{
				Latitude:     40.7128,
				Longitude:    -74.0060,
				DailyVars:    []string{"default"},
				ForecastDays: 14,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},

		// Valid combined requests
		{
			name: "current and hourly combined",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--hourly", "temperature_2m", "--forecast-hours", "24"},
			want: &ForecastConfig{
				Latitude:      40.7128,
				Longitude:     -74.0060,
				CurrentVars:   []string{"temperature_2m"},
				HourlyVars:    []string{"temperature_2m"},
				ForecastHours: 24,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},
		{
			name: "current and daily combined",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--daily", "default", "--forecast-days", "3"},
			want: &ForecastConfig{
				Latitude:     40.7128,
				Longitude:    -74.0060,
				CurrentVars:  []string{"temperature_2m"},
				DailyVars:    []string{"default"},
				ForecastDays: 3,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly and daily combined",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "12", "--daily", "weather_code", "--forecast-days", "5"},
			want: &ForecastConfig{
				Latitude:      40.7128,
				Longitude:     -74.0060,
				HourlyVars:    []string{"temperature_2m"},
				ForecastHours: 12,
				DailyVars:     []string{"weather_code"},
				ForecastDays:  5,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},
		{
			name: "current, hourly, and daily combined",
			args: []string{"--latitude", "52.52", "--longitude", "13.41", "--current", "default", "--hourly", "temperature_2m", "--forecast-hours", "24", "--daily", "default", "--forecast-days", "7"},
			want: &ForecastConfig{
				Latitude:      52.52,
				Longitude:     13.41,
				CurrentVars:   []string{"default"},
				HourlyVars:    []string{"temperature_2m"},
				ForecastHours: 24,
				DailyVars:     []string{"default"},
				ForecastDays:  7,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},

		// Valid with city
		{
			name: "current-only with city",
			args: []string{"--city", "Berlin", "--current", "temperature_2m,weather_code"},
			want: &ForecastConfig{
				City:        "Berlin",
				CurrentVars: []string{"temperature_2m", "weather_code"},
				Units:       "metric",
				Format:      "toon",
			},
			wantErr: false,
		},
		{
			name: "hourly-only with city",
			args: []string{"--city", "New York", "--hourly", "temperature_2m", "--forecast-hours", "48"},
			want: &ForecastConfig{
				City:          "New York",
				HourlyVars:    []string{"temperature_2m"},
				ForecastHours: 48,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},
		{
			name: "daily-only with city",
			args: []string{"--city", "Tokyo", "--daily", "default", "--forecast-days", "14"},
			want: &ForecastConfig{
				City:         "Tokyo",
				DailyVars:    []string{"default"},
				ForecastDays: 14,
				Units:        "metric",
				Format:       "toon",
			},
			wantErr: false,
		},
		{
			name: "combined with city",
			args: []string{"--city", "Paris", "--current", "default", "--hourly", "temperature_2m", "--forecast-hours", "24", "--daily", "weather_code", "--forecast-days", "5"},
			want: &ForecastConfig{
				City:          "Paris",
				CurrentVars:   []string{"default"},
				HourlyVars:    []string{"temperature_2m"},
				ForecastHours: 24,
				DailyVars:     []string{"weather_code"},
				ForecastDays:  5,
				Units:         "metric",
				Format:        "toon",
			},
			wantErr: false,
		},

		// Valid with units and format
		{
			name: "with imperial units",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--units", "imperial"},
			want: &ForecastConfig{
				Latitude:    40.7128,
				Longitude:   -74.0060,
				CurrentVars: []string{"temperature_2m"},
				Units:       "imperial",
				Format:      "toon",
			},
			wantErr: false,
		},
		{
			name: "with json format",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--format", "json"},
			want: &ForecastConfig{
				Latitude:    40.7128,
				Longitude:   -74.0060,
				CurrentVars: []string{"temperature_2m"},
				Units:       "metric",
				Format:      "json",
			},
			wantErr: false,
		},
		{
			name: "with both imperial units and json format",
			args: []string{"--city", "London", "--current", "default", "--units", "imperial", "--format", "json"},
			want: &ForecastConfig{
				City:        "London",
				CurrentVars: []string{"default"},
				Units:       "imperial",
				Format:      "json",
			},
			wantErr: false,
		},

		// Variable discovery
		{
			name: "variables query - all variables",
			args: []string{"variables"},
			want: &ForecastConfig{
				VariablesQuery: "",
				Units:          "metric",
				Format:         "toon",
			},
			wantErr: false,
		},
		{
			name: "variables query - current variables",
			args: []string{"variables", "current"},
			want: &ForecastConfig{
				VariablesQuery: "current",
				Units:          "metric",
				Format:         "toon",
			},
			wantErr: false,
		},
		{
			name: "variables query - hourly variables",
			args: []string{"variables", "hourly"},
			want: &ForecastConfig{
				VariablesQuery: "hourly",
				Units:          "metric",
				Format:         "toon",
			},
			wantErr: false,
		},
		{
			name: "variables query - daily variables",
			args: []string{"variables", "daily"},
			want: &ForecastConfig{
				VariablesQuery: "daily",
				Units:          "metric",
				Format:         "toon",
			},
			wantErr: false,
		},

		// CSV parsing with spaces
		{
			name: "CSV with spaces around values",
			args: []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", " temperature_2m , weather_code "},
			want: &ForecastConfig{
				Latitude:    40.7128,
				Longitude:   -74.0060,
				CurrentVars: []string{"temperature_2m", "weather_code"},
				Units:       "metric",
				Format:      "toon",
			},
			wantErr: false,
		},

		// Validation errors - no section specified
		{
			name:    "error: no section specified with coordinates",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060"},
			want:    nil,
			wantErr: true,
			errMsg:  "at least one of --current, --hourly, or --daily is required",
		},
		{
			name:    "error: no section specified with city",
			args:    []string{"--city", "Berlin"},
			want:    nil,
			wantErr: true,
			errMsg:  "at least one of --current, --hourly, or --daily is required",
		},

		// Validation errors - missing required flags
		{
			name:    "error: hourly without forecast-hours",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-hours is required when --hourly is specified",
		},
		{
			name:    "error: daily without forecast-days",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-days is required when --daily is specified",
		},
		{
			name:    "error: no city or coordinates",
			args:    []string{"--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "either --city or both --latitude and --longitude are required",
		},

		// Validation errors - city mutually exclusive with coordinates
		{
			name:    "error: city with latitude",
			args:    []string{"--city", "Berlin", "--latitude", "52.52", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--city cannot be combined with --latitude or --longitude",
		},
		{
			name:    "error: city with longitude",
			args:    []string{"--city", "Berlin", "--longitude", "13.41", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--city cannot be combined with --latitude or --longitude",
		},
		{
			name:    "error: city with both coordinates",
			args:    []string{"--city", "Berlin", "--latitude", "52.52", "--longitude", "13.41", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--city cannot be combined with --latitude or --longitude",
		},

		// Validation errors - latitude and longitude must be together
		{
			name:    "error: latitude without longitude",
			args:    []string{"--latitude", "52.52", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--latitude and --longitude must be provided together",
		},
		{
			name:    "error: longitude without latitude",
			args:    []string{"--longitude", "13.41", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--latitude and --longitude must be provided together",
		},

		// Validation errors - forecast-hours out of range
		{
			name:    "error: forecast-hours zero",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "0"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-hours must be at least 1",
		},
		{
			name:    "error: forecast-hours exceeds max",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "49"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-hours must be at most 48",
		},
		{
			name:    "error: forecast-hours invalid",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "abc"},
			want:    nil,
			wantErr: true,
			errMsg:  "invalid forecast-hours value",
		},

		// Validation errors - forecast-days out of range
		{
			name:    "error: forecast-days zero",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code", "--forecast-days", "0"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-days must be at least 1",
		},
		{
			name:    "error: forecast-days exceeds max",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code", "--forecast-days", "15"},
			want:    nil,
			wantErr: true,
			errMsg:  "--forecast-days must be at most 14",
		},
		{
			name:    "error: forecast-days invalid",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code", "--forecast-days", "xyz"},
			want:    nil,
			wantErr: true,
			errMsg:  "invalid forecast-days value",
		},

		// Validation errors - coordinate ranges
		{
			name:    "error: latitude too high",
			args:    []string{"--latitude", "91", "--longitude", "0", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name:    "error: latitude too low",
			args:    []string{"--latitude", "-91", "--longitude", "0", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name:    "error: longitude too high",
			args:    []string{"--latitude", "0", "--longitude", "181", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name:    "error: longitude too low",
			args:    []string{"--latitude", "0", "--longitude", "-181", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},

		// Validation errors - units and format
		{
			name:    "error: invalid units",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--units", "celsius"},
			want:    nil,
			wantErr: true,
			errMsg:  "units must be 'metric' or 'imperial'",
		},
		{
			name:    "error: invalid format",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--format", "xml"},
			want:    nil,
			wantErr: true,
			errMsg:  "format must be 'toon' or 'json'",
		},

		// Validation errors - duplicate variables
		{
			name:    "error: duplicate current variable",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m,temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate current variable: temperature_2m",
		},
		{
			name:    "error: duplicate hourly variable",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m,temperature_2m", "--forecast-hours", "24"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate hourly variable: temperature_2m",
		},
		{
			name:    "error: duplicate daily variable",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code,weather_code", "--forecast-days", "7"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate daily variable: weather_code",
		},
		{
			name:    "error: duplicate with spaces",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m, temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate current variable: temperature_2m",
		},

		// Validation errors - unknown variables
		{
			name:    "error: unknown current variable",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "not_a_variable"},
			want:    nil,
			wantErr: true,
			errMsg:  "unknown current variable: not_a_variable",
		},
		{
			name:    "error: unknown hourly variable",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "invalid_var", "--forecast-hours", "24"},
			want:    nil,
			wantErr: true,
			errMsg:  "unknown hourly variable: invalid_var",
		},
		{
			name:    "error: unknown daily variable",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "fake_var", "--forecast-days", "7"},
			want:    nil,
			wantErr: true,
			errMsg:  "unknown daily variable: fake_var",
		},
		{
			name:    "error: unknown variable among valid ones",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m,invalid_var"},
			want:    nil,
			wantErr: true,
			errMsg:  "unknown current variable: invalid_var",
		},

		// Validation errors - invalid city input
		{
			name:    "error: empty city",
			args:    []string{"--city", "", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "--city requires a non-empty value",
		},

		// Validation errors - duplicate flags
		{
			name:    "error: duplicate --current",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--current", "weather_code"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --current",
		},
		{
			name:    "error: duplicate --hourly",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "24", "--hourly", "weather_code"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --hourly",
		},
		{
			name:    "error: duplicate --daily",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--daily", "weather_code", "--forecast-days", "7", "--daily", "temperature_2m_max"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --daily",
		},
		{
			name:    "error: duplicate --city",
			args:    []string{"--city", "Berlin", "--city", "Paris", "--current", "temperature_2m"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --city",
		},
		{
			name:    "error: duplicate --forecast-hours",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--hourly", "temperature_2m", "--forecast-hours", "24", "--forecast-hours", "48"},
			want:    nil,
			wantErr: true,
			errMsg:  "duplicate flag: --forecast-hours",
		},

		// Validation errors - unknown flags
		{
			name:    "error: unknown flag",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "--unknown"},
			want:    nil,
			wantErr: true,
			errMsg:  "unknown flag: --unknown",
		},
		{
			name:    "error: unexpected positional argument",
			args:    []string{"--latitude", "40.7128", "--longitude", "-74.0060", "--current", "temperature_2m", "extra"},
			want:    nil,
			wantErr: true,
			errMsg:  "unexpected argument: extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseForecast(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseForecast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("ParseForecast() error message = %q, want %q", err.Error(), tt.errMsg)
				}
			}
			if err == nil && !tt.wantErr {
				if got.City != tt.want.City {
					t.Errorf("ParseForecast() City = %v, want %v", got.City, tt.want.City)
				}
				if got.Country != tt.want.Country {
					t.Errorf("ParseForecast() Country = %v, want %v", got.Country, tt.want.Country)
				}
				if got.Latitude != tt.want.Latitude {
					t.Errorf("ParseForecast() Latitude = %v, want %v", got.Latitude, tt.want.Latitude)
				}
				if got.Longitude != tt.want.Longitude {
					t.Errorf("ParseForecast() Longitude = %v, want %v", got.Longitude, tt.want.Longitude)
				}
				if got.ForecastHours != tt.want.ForecastHours {
					t.Errorf("ParseForecast() ForecastHours = %v, want %v", got.ForecastHours, tt.want.ForecastHours)
				}
				if got.ForecastDays != tt.want.ForecastDays {
					t.Errorf("ParseForecast() ForecastDays = %v, want %v", got.ForecastDays, tt.want.ForecastDays)
				}
				if got.Units != tt.want.Units {
					t.Errorf("ParseForecast() Units = %v, want %v", got.Units, tt.want.Units)
				}
				if got.Format != tt.want.Format {
					t.Errorf("ParseForecast() Format = %v, want %v", got.Format, tt.want.Format)
				}
				if got.VariablesQuery != tt.want.VariablesQuery {
					t.Errorf("ParseForecast() VariablesQuery = %v, want %v", got.VariablesQuery, tt.want.VariablesQuery)
				}
				if !reflect.DeepEqual(got.CurrentVars, tt.want.CurrentVars) {
					t.Errorf("ParseForecast() CurrentVars = %v, want %v", got.CurrentVars, tt.want.CurrentVars)
				}
				if !reflect.DeepEqual(got.HourlyVars, tt.want.HourlyVars) {
					t.Errorf("ParseForecast() HourlyVars = %v, want %v", got.HourlyVars, tt.want.HourlyVars)
				}
				if !reflect.DeepEqual(got.DailyVars, tt.want.DailyVars) {
					t.Errorf("ParseForecast() DailyVars = %v, want %v", got.DailyVars, tt.want.DailyVars)
				}
			}
		})
	}
}

func TestVariableDefinitions(t *testing.T) {
	vars := GetVariableDefinitions()

	t.Run("current defaults are defined", func(t *testing.T) {
		if len(vars.CurrentDefaults) == 0 {
			t.Error("Expected non-empty current defaults")
		}
		for _, v := range vars.CurrentDefaults {
			if _, ok := vars.CurrentVars[v]; !ok {
				t.Errorf("Current default %q not found in CurrentVars", v)
			}
		}
	})

	t.Run("hourly defaults are defined", func(t *testing.T) {
		if len(vars.HourlyDefaults) == 0 {
			t.Error("Expected non-empty hourly defaults")
		}
		for _, v := range vars.HourlyDefaults {
			if _, ok := vars.HourlyVars[v]; !ok {
				t.Errorf("Hourly default %q not found in HourlyVars", v)
			}
		}
	})

	t.Run("daily defaults are defined", func(t *testing.T) {
		if len(vars.DailyDefaults) == 0 {
			t.Error("Expected non-empty daily defaults")
		}
		for _, v := range vars.DailyDefaults {
			if _, ok := vars.DailyVars[v]; !ok {
				t.Errorf("Daily default %q not found in DailyVars", v)
			}
		}
	})

	t.Run("expected current variables exist", func(t *testing.T) {
		expected := []string{"temperature_2m", "apparent_temperature", "precipitation", "wind_speed_10m", "weather_code"}
		for _, v := range expected {
			if _, ok := vars.CurrentVars[v]; !ok {
				t.Errorf("Expected current variable %q not found", v)
			}
		}
	})

	t.Run("expected hourly variables exist", func(t *testing.T) {
		expected := []string{"temperature_2m", "precipitation_probability", "precipitation", "wind_speed_10m", "weather_code"}
		for _, v := range expected {
			if _, ok := vars.HourlyVars[v]; !ok {
				t.Errorf("Expected hourly variable %q not found", v)
			}
		}
	})

	t.Run("expected daily variables exist", func(t *testing.T) {
		expected := []string{"weather_code", "temperature_2m_min", "temperature_2m_max", "precipitation_sum", "sunrise", "sunset"}
		for _, v := range expected {
			if _, ok := vars.DailyVars[v]; !ok {
				t.Errorf("Expected daily variable %q not found", v)
			}
		}
	})
}

func TestExpandDefaultVars(t *testing.T) {
	vars := GetVariableDefinitions()

	t.Run("expand current default", func(t *testing.T) {
		input := []string{"default"}
		result := vars.ExpandDefaultVars(input, "current")
		if !reflect.DeepEqual(result, vars.CurrentDefaults) {
			t.Errorf("ExpandDefaultVars(current) = %v, want %v", result, vars.CurrentDefaults)
		}
	})

	t.Run("expand hourly default", func(t *testing.T) {
		input := []string{"default"}
		result := vars.ExpandDefaultVars(input, "hourly")
		if !reflect.DeepEqual(result, vars.HourlyDefaults) {
			t.Errorf("ExpandDefaultVars(hourly) = %v, want %v", result, vars.HourlyDefaults)
		}
	})

	t.Run("expand daily default", func(t *testing.T) {
		input := []string{"default"}
		result := vars.ExpandDefaultVars(input, "daily")
		if !reflect.DeepEqual(result, vars.DailyDefaults) {
			t.Errorf("ExpandDefaultVars(daily) = %v, want %v", result, vars.DailyDefaults)
		}
	})

	t.Run("mix default with explicit variables", func(t *testing.T) {
		input := []string{"temperature_2m", "default", "weather_code"}
		result := vars.ExpandDefaultVars(input, "current")
		if len(result) != len(vars.CurrentDefaults)+2 {
			t.Errorf("Expected %d variables, got %d", len(vars.CurrentDefaults)+2, len(result))
		}
		if result[0] != "temperature_2m" {
			t.Errorf("First variable should be temperature_2m, got %v", result[0])
		}
		if result[len(result)-1] != "weather_code" {
			t.Errorf("Last variable should be weather_code, got %v", result[len(result)-1])
		}
	})

	t.Run("no default keyword", func(t *testing.T) {
		input := []string{"temperature_2m", "weather_code"}
		result := vars.ExpandDefaultVars(input, "current")
		if !reflect.DeepEqual(result, input) {
			t.Errorf("ExpandDefaultVars() = %v, want %v", result, input)
		}
	})
}

func TestValidateVars(t *testing.T) {
	vars := GetVariableDefinitions()

	t.Run("valid current variables", func(t *testing.T) {
		input := []string{"temperature_2m", "weather_code"}
		err := vars.ValidateVars(input, "current")
		if err != nil {
			t.Errorf("ValidateVars() error = %v", err)
		}
	})

	t.Run("valid hourly variables", func(t *testing.T) {
		input := []string{"temperature_2m", "precipitation_probability"}
		err := vars.ValidateVars(input, "hourly")
		if err != nil {
			t.Errorf("ValidateVars() error = %v", err)
		}
	})

	t.Run("valid daily variables", func(t *testing.T) {
		input := []string{"weather_code", "temperature_2m_max"}
		err := vars.ValidateVars(input, "daily")
		if err != nil {
			t.Errorf("ValidateVars() error = %v", err)
		}
	})

	t.Run("default keyword is valid", func(t *testing.T) {
		input := []string{"default"}
		err := vars.ValidateVars(input, "current")
		if err != nil {
			t.Errorf("ValidateVars() with 'default' should not error, got %v", err)
		}
	})

	t.Run("invalid current variable", func(t *testing.T) {
		input := []string{"invalid_var"}
		err := vars.ValidateVars(input, "current")
		if err == nil {
			t.Error("ValidateVars() should return error for invalid variable")
		}
		if !strings.Contains(err.Error(), "unknown current variable") {
			t.Errorf("Error message should mention 'unknown current variable', got: %v", err)
		}
	})

	t.Run("invalid hourly variable", func(t *testing.T) {
		input := []string{"invalid_var"}
		err := vars.ValidateVars(input, "hourly")
		if err == nil {
			t.Error("ValidateVars() should return error for invalid variable")
		}
		if !strings.Contains(err.Error(), "unknown hourly variable") {
			t.Errorf("Error message should mention 'unknown hourly variable', got: %v", err)
		}
	})

	t.Run("invalid daily variable", func(t *testing.T) {
		input := []string{"invalid_var"}
		err := vars.ValidateVars(input, "daily")
		if err == nil {
			t.Error("ValidateVars() should return error for invalid variable")
		}
		if !strings.Contains(err.Error(), "unknown daily variable") {
			t.Errorf("Error message should mention 'unknown daily variable', got: %v", err)
		}
	})

	t.Run("invalid section", func(t *testing.T) {
		input := []string{"temperature_2m"}
		err := vars.ValidateVars(input, "invalid")
		if err == nil {
			t.Error("ValidateVars() should return error for invalid section")
		}
	})
}

func TestParseForecastWithHelpFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "forecast with -h",
			args:    []string{"-h"},
			wantErr: false,
		},
		{
			name:    "forecast with --help",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "forecast with --help and other flags",
			args:    []string{"--current", "temperature_2m", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseForecast(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseForecast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

