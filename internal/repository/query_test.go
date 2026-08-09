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
			name:         "Space separated OR with excludeInput",
			input:        "golang react",
			excludeInput: "python",
			searchType:   "or",
			want:         `("golang" OR "react") NOT "python"`,
		},
		{
			name:         "Hyphen exclusion inside input with OR searchType",
			input:        "golang react -python",
			excludeInput: "",
			searchType:   "or",
			want:         `("golang" OR "react") NOT "python"`,
		},
		{
			name:         "Exclude input multiple words",
			input:        "cat dog",
			excludeInput: "bird fish",
			searchType:   "or",
			want:         `("cat" OR "dog") NOT "bird" NOT "fish"`,
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

func TestBuildLikeQuery(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		excludeInput string
		searchType   string
		wantCond     string
		wantParams   []interface{}
	}{
		{
			name:         "Space separated OR with excludeInput",
			input:        "golang react",
			excludeInput: "python",
			searchType:   "or",
			wantCond:     "((t.full_text LIKE ? OR t.full_text LIKE ?) AND t.full_text NOT LIKE ?)",
			wantParams:   []interface{}{"%golang%", "%react%", "%python%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCond, gotParams := BuildLikeQuery(tt.input, tt.excludeInput, tt.searchType)
			if gotCond != tt.wantCond {
				t.Errorf("BuildLikeQuery() cond = %q; want %q", gotCond, tt.wantCond)
			}
			if len(gotParams) != len(tt.wantParams) {
				t.Fatalf("BuildLikeQuery() params len = %d; want %d", len(gotParams), len(tt.wantParams))
			}
			for i, p := range gotParams {
				if p != tt.wantParams[i] {
					t.Errorf("BuildLikeQuery() param[%d] = %v; want %v", i, p, tt.wantParams[i])
				}
			}
		})
	}
}
