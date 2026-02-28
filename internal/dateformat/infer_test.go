package dateformat

import (
	"testing"
)

func TestInferFormat(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "ISO date",
			values:   []string{"2024-01-15", "2024-01-16", "2024-01-17"},
			expected: "2006-01-02",
		},
		{
			name:     "US date",
			values:   []string{"01/15/2024", "01/16/2024", "01/17/2024"},
			expected: "01/02/2006",
		},
		{
			name:     "Text date",
			values:   []string{"Jan 15, 2024", "Jan 16, 2024", "Jan 17, 2024"},
			expected: "Jan 2, 2006",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferFormat(tt.values)
			if result == nil {
				t.Fatalf("expected result, got nil")
			}
			if result.Format.Layout != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.Format.Layout)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	converter := NewConverter("01/02/2006", "2006-01-02")

	tests := []struct {
		input    string
		expected string
	}{
		{"01/15/2024", "2024-01-15"},
		{"12/25/2023", "2023-12-25"},
		{"", ""},
	}

	for _, tt := range tests {
		result := converter.Convert(tt.input)
		if result != tt.expected {
			t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
