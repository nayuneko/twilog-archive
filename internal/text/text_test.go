package text

import "testing"

func TestUnescapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "User requested example",
			input:    "(*&gt;△&lt;)&lt;ナーンナーンっっ",
			expected: "(*>△<)<ナーンナーンっっ",
		},
		{
			name:     "Double escaped ampersand entity",
			input:    "Blu-ray&amp;amp;DVD",
			expected: "Blu-ray&DVD",
		},
		{
			name:     "Single escaped ampersand entity",
			input:    "Blu-ray&amp;DVD",
			expected: "Blu-ray&DVD",
		},
		{
			name:     "Quotes and apostrophes",
			input:    "&quot;Hello&#39;s World&quot;",
			expected: `"Hello's World"`,
		},
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnescapeHTML(tt.input)
			if got != tt.expected {
				t.Errorf("UnescapeHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
