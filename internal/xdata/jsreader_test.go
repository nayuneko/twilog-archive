package xdata

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestStripJSPrefixReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Twitter archive tweets.js style",
			input:    `window.YTD.tweets.part0 = [ {"tweet": {"id": "123"}} ];`,
			expected: `[ {"tweet": {"id": "123"}} ];`,
		},
		{
			name:     "Already valid JSON array",
			input:    `[{"id": 1}]`,
			expected: `[{"id": 1}]`,
		},
		{
			name:     "JSON object style",
			input:    `var data = {"key": "value"};`,
			expected: `{"key": "value"};`,
		},
		{
			name:     "With newlines and spaces before array",
			input:    "window.YTD.like.part0 =\n  [\n    {\"like\": {}}\n  ];",
			expected: "[\n    {\"like\": {}}\n  ];",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewStripJSPrefixReader(strings.NewReader(tt.input))
			out, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(out) != tt.expected {
				t.Errorf("got %q, want %q", string(out), tt.expected)
			}

			// JSON デコーダでパースできるか確認
			r2 := NewStripJSPrefixReader(strings.NewReader(tt.input))
			var dummy interface{}
			if err := json.NewDecoder(r2).Decode(&dummy); err != nil {
				t.Errorf("failed to decode JSON: %v", err)
			}
		})
	}
}
