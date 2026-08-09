package repository

import (
	"testing"
)

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		searchType string
		want       string
	}{
		{
			name:       "Single word",
			input:      "golang",
			searchType: "and",
			want:       `"golang"`,
		},
		{
			name:       "Space separated AND",
			input:      "golang react",
			searchType: "and",
			want:       `"golang" AND "react"`,
		},
		{
			name:       "Full width space separated AND",
			input:      "golang　react",
			searchType: "and",
			want:       `"golang" AND "react"`,
		},
		{
			name:       "Space separated OR",
			input:      "golang react",
			searchType: "or",
			want:       `"golang" OR "react"`,
		},
		{
			name:       "Explicit OR keyword",
			input:      "golang OR react",
			searchType: "and",
			want:       `"golang" OR "react"`,
		},
		{
			name:       "NOT / hyphen exclusion",
			input:      "golang -python",
			searchType: "and",
			want:       `"golang" AND NOT "python"`,
		},
		{
			name:       "NOT keyword",
			input:      "golang NOT python",
			searchType: "and",
			want:       `"golang" AND NOT "python"`,
		},
		{
			name:       "Quoted phrase",
			input:      `"hello world" test`,
			searchType: "and",
			want:       `"hello world" AND "test"`,
		},
		{
			name:       "OR mode with NOT exclusion",
			input:      "golang react -python",
			searchType: "or",
			want:       `"golang" OR "react" OR NOT "python"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFTSQuery(tt.input, tt.searchType)
			if got != tt.want {
				t.Errorf("BuildFTSQuery(%q, %q) = %q; want %q", tt.input, tt.searchType, got, tt.want)
			}
		})
	}
}

func TestBuildLikeQuery(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		searchType string
		wantCond   string
		wantParams []interface{}
	}{
		{
			name:       "Space separated AND",
			input:      "golang react",
			searchType: "and",
			wantCond:   "(t.full_text LIKE ? AND t.full_text LIKE ?)",
			wantParams: []interface{}{"%golang%", "%react%"},
		},
		{
			name:       "Space separated OR",
			input:      "golang react",
			searchType: "or",
			wantCond:   "(t.full_text LIKE ? OR t.full_text LIKE ?)",
			wantParams: []interface{}{"%golang%", "%react%"},
		},
		{
			name:       "Exclusion NOT",
			input:      "golang -python",
			searchType: "and",
			wantCond:   "(t.full_text LIKE ? AND t.full_text NOT LIKE ?)",
			wantParams: []interface{}{"%golang%", "%python%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCond, gotParams := BuildLikeQuery(tt.input, tt.searchType)
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
