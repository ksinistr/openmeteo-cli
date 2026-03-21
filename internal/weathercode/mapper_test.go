package weathercode

import "testing"

func TestMapper_GetDescription(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "Clear sky"},
		{1, "Mainly clear"},
		{2, "Partly cloudy"},
		{3, "Overcast"},
		{45, "Fog"},
		{48, "Depositing rime fog"},
		{51, "Light drizzle"},
		{61, "Slight rain"},
		{63, "Moderate rain"},
		{65, "Heavy rain"},
		{71, "Slight snow fall"},
		{73, "Moderate snow fall"},
		{75, "Heavy snow fall"},
		{80, "Slight rain showers"},
		{95, "Thunderstorm"},
		{96, "Thunderstorm with slight hail"},
		{99, "Thunderstorm with heavy hail"},
	}

	mapper := NewMapper()

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := mapper.GetDescription(tt.code)
			if result != tt.expected {
				t.Errorf("GetDescription(%d) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestMapper_GetDescription_UnknownCode(t *testing.T) {
	mapper := NewMapper()
	result := mapper.GetDescription(999)
	expected := "Unknown weather code: 999"
	if result != expected {
		t.Errorf("GetDescription(999) = %q, want %q", result, expected)
	}
}

func TestMapper_GetCode(t *testing.T) {
	tests := []struct {
		desc     string
		expected int
	}{
		{"Clear sky", 0},
		{"Mainly clear", 1},
		{"Partly cloudy", 2},
		{"Overcast", 3},
		{"Thunderstorm", 95},
	}

	mapper := NewMapper()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := mapper.GetCode(tt.desc)
			if result != tt.expected {
				t.Errorf("GetCode(%q) = %d, want %d", tt.desc, result, tt.expected)
			}
		})
	}
}

func TestMapper_GetCode_UnknownDescription(t *testing.T) {
	mapper := NewMapper()
	result := mapper.GetCode("Unknown weather")
	if result != -1 {
		t.Errorf("GetCode(\"Unknown weather\") = %d, want -1", result)
	}
}
