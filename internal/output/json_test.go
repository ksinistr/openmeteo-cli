package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"openmeteo-cli/internal/forecast"
)

func TestJSONEncoder_EncodeHourly(t *testing.T) {
	tests := []struct {
		name     string
		input    *forecast.HourlyOutput
		checkers []func(*testing.T, *bytes.Buffer)
	}{
		{
			name: "basic hourly output",
			input: &forecast.HourlyOutput{
				Meta: forecast.Meta{
					GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
					Units: forecast.Units{
						Temperature:              "C",
						Humidity:                 "%",
						WindSpeed:                "km/h",
						WindDirection:            "deg",
						Precipitation:            "mm",
						PrecipitationProbability: "%",
						UVIndex:                  "index",
					},
					Timezone:  "Europe/Berlin",
					Latitude:  52.52,
					Longitude: 13.405,
				},
				Days: map[string]forecast.DayHours{
					"2026-03-21": {
						Hours: []forecast.Hour{
							{
								Time:                     "06:00",
								Weather:                  "Clear sky",
								Temperature:              18.2,
								ApparentTemperature:      17.5,
								Humidity:                 72,
								Precipitation:            0.0,
								PrecipitationProbability: 0,
								WindSpeed:                8.5,
								WindGusts:                12.1,
								WindDirection:            230,
								UVIndex:                  1.2,
							},
							{
								Time:                     "12:00",
								Weather:                  "Mainly clear",
								Temperature:              22.5,
								ApparentTemperature:      21.8,
								Humidity:                 65,
								Precipitation:            0.0,
								PrecipitationProbability: 10,
								WindSpeed:                12.3,
								WindGusts:                18.7,
								WindDirection:            245,
								UVIndex:                  5.2,
							},
						},
					},
				},
			},
			checkers: []func(*testing.T, *bytes.Buffer){
				checkJSONHasMeta,
				checkJSONHasDays,
				checkJSONHasNumericValues,
			},
		},
		{
			name: "empty hours array",
			input: &forecast.HourlyOutput{
				Meta: forecast.Meta{
					GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
					Units: forecast.Units{
						Temperature:              "C",
						Humidity:                 "%",
						WindSpeed:                "km/h",
						WindDirection:            "deg",
						Precipitation:            "mm",
						PrecipitationProbability: "%",
						UVIndex:                  "index",
					},
					Timezone:  "Europe/Berlin",
					Latitude:  52.52,
					Longitude: 13.405,
				},
				Days: map[string]forecast.DayHours{},
			},
			checkers: []func(*testing.T, *bytes.Buffer){
				checkJSONHasMeta,
				checkJSONHasEmptyHourlyDays,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeJSON(tt.input, &buf)
			if err != nil {
				t.Fatalf("writeJSON() error = %v", err)
			}

			for _, checker := range tt.checkers {
				checker(t, &buf)
			}
		})
	}
}

func TestJSONEncoder_EncodeDaily(t *testing.T) {
	tests := []struct {
		name     string
		input    *forecast.DailyOutput
		checkers []func(*testing.T, *bytes.Buffer)
	}{
		{
			name: "basic daily output",
			input: &forecast.DailyOutput{
				Meta: forecast.Meta{
					GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
					Units: forecast.Units{
						Temperature:              "C",
						Humidity:                 "%",
						WindSpeed:                "km/h",
						WindDirection:            "deg",
						Precipitation:            "mm",
						PrecipitationProbability: "%",
						UVIndex:                  "index",
					},
					Timezone:  "Europe/Berlin",
					Latitude:  52.52,
					Longitude: 13.405,
				},
				Days: []forecast.Day{
					{
						Date:                        "2026-03-21",
						Weather:                     "Clear sky",
						TempMin:                     15.0,
						TempMax:                     25.0,
						PrecipitationSum:            0.0,
						PrecipitationProbabilityMax: 10,
						WindSpeedMax:                12.0,
						WindGustsMax:                18.0,
						UVIndexMax:                  5.2,
						Sunrise:                     "06:00",
						Sunset:                      "18:00",
					},
					{
						Date:                        "2026-03-22",
						Weather:                     "Mainly clear",
						TempMin:                     16.0,
						TempMax:                     26.0,
						PrecipitationSum:            0.5,
						PrecipitationProbabilityMax: 15,
						WindSpeedMax:                14.0,
						WindGustsMax:                20.0,
						UVIndexMax:                  6.0,
						Sunrise:                     "05:59",
						Sunset:                      "18:01",
					},
				},
			},
			checkers: []func(*testing.T, *bytes.Buffer){
				checkJSONHasMeta,
				checkJSONHasDays,
				checkJSONHasNumericValues,
			},
		},
		{
			name: "empty days array",
			input: &forecast.DailyOutput{
				Meta: forecast.Meta{
					GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
					Units: forecast.Units{
						Temperature:              "C",
						Humidity:                 "%",
						WindSpeed:                "km/h",
						WindDirection:            "deg",
						Precipitation:            "mm",
						PrecipitationProbability: "%",
						UVIndex:                  "index",
					},
					Timezone:  "Europe/Berlin",
					Latitude:  52.52,
					Longitude: 13.405,
				},
				Days: []forecast.Day{},
			},
			checkers: []func(*testing.T, *bytes.Buffer){
				checkJSONHasMeta,
				checkJSONHasEmptyDays,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeJSON(tt.input, &buf)
			if err != nil {
				t.Fatalf("writeJSON() error = %v", err)
			}

			for _, checker := range tt.checkers {
				checker(t, &buf)
			}
		})
	}
}

// Deprecated encoder tests (Backward compatibility is maintained through Writer.Write)

func TestJSONEncoder_EncodeHourlyTo(t *testing.T) {
	output := &forecast.HourlyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature: "C",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: map[string]forecast.DayHours{
			"2026-03-21": {
				Hours: []forecast.Hour{
					{Time: "12:00", Temperature: 20.0},
				},
			},
		},
	}

	var buf bytes.Buffer
	enc := NewJSONEncoder()
	err := enc.EncodeHourlyTo(output, &buf)
	if err != nil {
		t.Fatalf("EncodeHourlyTo() error = %v", err)
	}

	if !strings.Contains(buf.String(), `"temperature": 20`) {
		t.Errorf("Expected temperature value in output, got: %s", buf.String())
	}
}

func TestJSONEncoder_EncodeDailyTo(t *testing.T) {
	output := &forecast.DailyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature: "C",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: []forecast.Day{
			{Date: "2026-03-21", TempMin: 15.0, TempMax: 25.0},
		},
	}

	var buf bytes.Buffer
	enc := NewJSONEncoder()
	err := enc.EncodeDailyTo(output, &buf)
	if err != nil {
		t.Fatalf("EncodeDailyTo() error = %v", err)
	}

	if !strings.Contains(buf.String(), `"temp_max": 25`) {
		t.Errorf("Expected temp_max value in output, got: %s", buf.String())
	}
}

func TestJSONEncoder_TypedNilHourly(t *testing.T) {
	var typedNilHourly *forecast.HourlyOutput = nil
	enc := NewJSONEncoder()
	err := enc.EncodeHourlyTo(typedNilHourly, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for typed nil HourlyOutput, got nil")
	}
}

func TestJSONEncoder_TypedNilDaily(t *testing.T) {
	var typedNilDaily *forecast.DailyOutput = nil
	enc := NewJSONEncoder()
	err := enc.EncodeDailyTo(typedNilDaily, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for typed nil DailyOutput, got nil")
	}
}

// Deprecated encoder tests (Backward compatibility is maintained through Writer.Write)

func TestToonEncoder_EncodeHourly(t *testing.T) {
	output := &forecast.HourlyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature: "C",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: map[string]forecast.DayHours{
			"2026-03-21": {
				Hours: []forecast.Hour{
					{Time: "12:00", Temperature: 20.0},
				},
			},
		},
	}

	enc := NewToonEncoder()
	result, err := enc.EncodeHourly(output)
	if err != nil {
		t.Fatalf("EncodeHourly() error = %v", err)
	}

	if !strings.Contains(result, ",20,") {
		t.Errorf("Expected temperature value in TOON output, got: %s", result)
	}
}

func TestToonEncoder_EncodeDaily(t *testing.T) {
	output := &forecast.DailyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature: "C",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: []forecast.Day{
			{Date: "2026-03-21", TempMin: 15.0, TempMax: 25.0},
		},
	}

	enc := NewToonEncoder()
	result, err := enc.EncodeDaily(output)
	if err != nil {
		t.Fatalf("EncodeDaily() error = %v", err)
	}

	if !strings.Contains(result, ",25,") {
		t.Errorf("Expected temp_max value in TOON output, got: %s", result)
	}
}

func TestToonEncoder_TypedNilHourly(t *testing.T) {
	var typedNilHourly *forecast.HourlyOutput = nil
	enc := NewToonEncoder()
	_, err := enc.EncodeHourly(typedNilHourly)
	if err == nil {
		t.Error("expected error for typed nil HourlyOutput, got nil")
	}
}

func TestToonEncoder_TypedNilDaily(t *testing.T) {
	var typedNilDaily *forecast.DailyOutput = nil
	enc := NewToonEncoder()
	_, err := enc.EncodeDaily(typedNilDaily)
	if err == nil {
		t.Error("expected error for typed nil DailyOutput, got nil")
	}
}

// TestWriter_EncodeHourly tests Writer.Write with hourly output.
func TestWriter_EncodeHourly(t *testing.T) {
	output := &forecast.HourlyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature:              "C",
				Humidity:                 "%",
				WindSpeed:                "km/h",
				WindDirection:            "deg",
				Precipitation:            "mm",
				PrecipitationProbability: "%",
				UVIndex:                  "index",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: map[string]forecast.DayHours{
			"2026-03-21": {
				Hours: []forecast.Hour{
					{
						Time:                     "12:00",
						Weather:                  "Clear sky",
						Temperature:              20.0,
						ApparentTemperature:      18.0,
						Humidity:                 65,
						Precipitation:            0.0,
						PrecipitationProbability: 10,
						WindSpeed:                12.0,
						WindGusts:                18.0,
						WindDirection:            245,
						UVIndex:                  5.2,
					},
				},
			},
		},
	}

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(output, "json")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Parse the JSON to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify required fields exist
	if _, ok := result["meta"]; !ok {
		t.Error("Expected 'meta' field in output")
	}
	if _, ok := result["days"]; !ok {
		t.Error("Expected 'days' field in output")
	}

	// Verify days data structure
	days, ok := result["days"].(map[string]interface{})
	if !ok {
		t.Fatal("Days should be a map")
	}

	dayData, ok := days["2026-03-21"].(map[string]interface{})
	if !ok {
		t.Fatal("Day data should be a map")
	}

	hours, ok := dayData["hours"].([]interface{})
	if !ok || len(hours) != 1 {
		t.Errorf("Expected 1 hour, got %d", len(hours))
	}

	hour, ok := hours[0].(map[string]interface{})
	if !ok {
		t.Fatal("Hour should be a map")
	}
	if hour["temperature"] != 20.0 {
		t.Errorf("Expected temperature 20.0, got %v", hour["temperature"])
	}
}

// TestWriter_EncodeDaily tests Writer.Write with daily output.
func TestWriter_EncodeDaily(t *testing.T) {
	output := &forecast.DailyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature:              "C",
				Humidity:                 "%",
				WindSpeed:                "km/h",
				WindDirection:            "deg",
				Precipitation:            "mm",
				PrecipitationProbability: "%",
				UVIndex:                  "index",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: []forecast.Day{
			{
				Date:                        "2026-03-21",
				Weather:                     "Clear sky",
				TempMin:                     15.0,
				TempMax:                     25.0,
				PrecipitationSum:            0.0,
				PrecipitationProbabilityMax: 10,
				WindSpeedMax:                12.0,
				WindGustsMax:                18.0,
				UVIndexMax:                  5.2,
				Sunrise:                     "2026-03-21T06:00:00Z",
				Sunset:                      "2026-03-21T18:00:00Z",
			},
		},
	}

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(output, "json")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Parse the JSON to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify required fields exist
	if _, ok := result["meta"]; !ok {
		t.Error("Expected 'meta' field in output")
	}
	if _, ok := result["days"]; !ok {
		t.Error("Expected 'days' field in output")
	}

	// Verify days data
	days, ok := result["days"].([]interface{})
	if !ok || len(days) != 1 {
		t.Errorf("Expected 1 day, got %d", len(days))
	}

	day, ok := days[0].(map[string]interface{})
	if !ok {
		t.Fatal("Day should be a map")
	}
	if day["temp_max"] != 25.0 {
		t.Errorf("Expected temp_max 25.0, got %v", day["temp_max"])
	}
}

// TestWriter_EncodeHourly_TOON tests Writer.Write with hourly output in TOON format.
func TestWriter_EncodeHourly_TOON(t *testing.T) {
	output := &forecast.HourlyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature: "C",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: map[string]forecast.DayHours{
			"2026-03-21": {
				Hours: []forecast.Hour{
					{Time: "12:00", Temperature: 20.0},
				},
			},
		},
	}

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(output, "toon")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if !strings.Contains(buf.String(), ",20,") {
		t.Errorf("Expected temperature value in TOON output, got: %s", buf.String())
	}
}

// TestWriter_EncodeDaily_TOON tests Writer.Write with daily output in TOON format.
func TestWriter_EncodeDaily_TOON(t *testing.T) {
	output := &forecast.DailyOutput{
		Meta: forecast.Meta{
			GeneratedAt: mustParseTime("2026-03-21T12:00:00Z"),
			Units: forecast.Units{
				Temperature: "C",
			},
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: []forecast.Day{
			{Date: "2026-03-21", TempMax: 25.0},
		},
	}

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(output, "toon")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if !strings.Contains(buf.String(), ",25,") {
		t.Errorf("Expected temp_max value in TOON output, got: %s", buf.String())
	}
}

// TestWriter_EncodeInvalidFormat tests Writer.Write with invalid format.
func TestWriter_EncodeInvalidFormat(t *testing.T) {
	output := &forecast.HourlyOutput{
		Meta: forecast.Meta{
			Timezone:  "UTC",
			Latitude:  40.0,
			Longitude: -74.0,
		},
		Days: map[string]forecast.DayHours{},
	}

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(output, "invalid")
	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected 'unknown format' error, got: %v", err)
	}
}

// TestWriter_EncodeNilInput tests Writer.Write with nil input.
func TestWriter_EncodeNilInput(t *testing.T) {
	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(nil, "json")
	if err == nil {
		t.Error("expected error for nil input, got nil")
	}
	if !IsEncodingError(err) {
		t.Errorf("expected EncodingError for nil input, got: %T", err)
	}
}

// TestWriter_TypedNilHourlyInput tests Writer.Write with typed nil input.
func TestWriter_TypedNilHourlyInput(t *testing.T) {
	var typedNil *forecast.HourlyOutput = nil

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(typedNil, "json")
	if err == nil {
		t.Error("expected error for typed nil HourlyOutput, got nil")
	}
}

// TestWriter_TypedNilDailyInput tests Writer.Write with typed nil input.
func TestWriter_TypedNilDailyInput(t *testing.T) {
	var typedNil *forecast.DailyOutput = nil

	w := NewWriter()
	var buf bytes.Buffer
	w.SetOutput(&buf)

	err := w.Write(typedNil, "json")
	if err == nil {
		t.Error("expected error for typed nil DailyOutput, got nil")
	}
}

// Helpers

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func checkJSONHasMeta(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if _, ok := result["meta"]; !ok {
		t.Error("Expected 'meta' field in output")
	}
}

func checkJSONHasHours(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if _, ok := result["hours"]; !ok {
		t.Error("Expected 'hours' field in output")
	}
}

func checkJSONHasDays(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if _, ok := result["days"]; !ok {
		t.Error("Expected 'days' field in output")
	}
}

func checkJSONHasEmptyHours(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if hours, ok := result["hours"].([]interface{}); !ok || len(hours) != 0 {
		t.Error("Expected empty 'hours' array in output")
	}
}

func checkJSONHasEmptyDays(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if days, ok := result["days"].([]interface{}); !ok || len(days) != 0 {
		t.Error("Expected empty 'days' array in output")
	}
}

func checkJSONHasEmptyHourlyDays(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if days, ok := result["days"].(map[string]interface{}); !ok || len(days) != 0 {
		t.Error("Expected empty 'days' map in output")
	}
}

func checkJSONHasNumericValues(t *testing.T, buf *bytes.Buffer) {
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'meta' to be a map")
	}
	if _, ok := meta["latitude"]; !ok {
		t.Error("Expected 'latitude' in meta")
	}
	if _, ok := meta["longitude"]; !ok {
		t.Error("Expected 'longitude' in meta")
	}
}
