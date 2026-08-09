package repository

import (
	"testing"
)

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		excludeInput string
		searchType   string
		want         string
	}{
		{
			name:         "Single word",
			input:        "golang",
			excludeInput: "",
			searchType:   "and",
			want:         `"golang"`,
		},
		{
			name:         "Space separated AND",
			input:        "golang react",
			excludeInput: "",
			searchType:   "and",
			want:         `("golang" AND "react")`,
		},
		{
			name:         "Hashtag in excludeInput",
			input:        "golang",
			excludeInput: "#告知",
			searchType:   "and",
			want:         `"golang" NOT "#告知"`,
		},
		{
			name:         "Hyphen hashtag in input",
			input:        "golang -#告知",
			excludeInput: "",
			searchType:   "and",
			want:         `"golang" NOT "#告知"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFTSQuery(tt.input, tt.excludeInput, tt.searchType)
			if got != tt.want {
				t.Errorf("BuildFTSQuery(%q, %q, %q) = %q; want %q", tt.input, tt.excludeInput, tt.searchType, got, tt.want)
			}
		})
	}
}

func TestBuildHashtagExcludeSQL(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		excludeInput string
		wantCond     string
		wantParamCount int
	}{
		{
			name:         "Exclude hashtag #告知",
			input:        "golang",
			excludeInput: "#告知",
			wantCond:     "t.id NOT IN (SELECT tweet_id FROM hashtags WHERE tag IN (?, ?, ?))",
			wantParamCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCond, gotParams := BuildHashtagExcludeSQL(tt.input, tt.excludeInput)
			if gotCond != tt.wantCond {
				t.Errorf("BuildHashtagExcludeSQL() cond = %q; want %q", gotCond, tt.wantCond)
			}
			if len(gotParams) != tt.wantParamCount {
				t.Errorf("BuildHashtagExcludeSQL() params len = %d; want %d", len(gotParams), tt.wantParamCount)
			}
		})
	}
}
