package ai

import (
	"testing"
)

// The JSON cleaners used by the Gemini and Kimi generators must extract the
// JSON object from model output wrapped in markdown or extra text.
func TestCleanJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Pure JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with markdown",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with text around",
			input:    "Here is the result: {\"key\": \"value\"} hope you like it",
			expected: `{"key": "value"}`,
		},
		{
			name:     "Nested JSON",
			input:    "Extra text {\"outer\": {\"inner\": 1}} more text",
			expected: `{"outer": {"inner": 1}}`,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := cleanJSON(tt.input); result != tt.expected {
				t.Errorf("cleanJSON() = %v, want %v", result, tt.expected)
			}
			if result := cleanKimiJSON(tt.input); result != tt.expected {
				t.Errorf("cleanKimiJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}
