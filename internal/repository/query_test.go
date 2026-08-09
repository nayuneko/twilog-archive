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

func TestIsFTSCompatible(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"杏奈", false},         // 2文字 (日本語)
		{"亜利沙", true},        // 3文字 (日本語)
		{"ab", false},           // 2文字 (ASCII)
		{"abc", true},           // 3文字 (ASCII)
		{"a", false},            // 1文字
		{"春香", false},          // 2文字 (日本語)
		{"可愛すぎる", true},      // 5文字 (日本語)
		{"", false},             // 空文字
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := isFTSCompatible(tt.text)
			if got != tt.want {
				t.Errorf("isFTSCompatible(%q) = %v; want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestBuildHybridSearchQuery(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		excludeInput    string
		searchType      string
		wantFTS         string
		wantHasFTS      bool
		wantLikeConds   int
		wantLikeParams  int
	}{
		{
			name:           "All tokens FTS compatible",
			input:          "亜利沙 golang",
			excludeInput:   "",
			searchType:     "and",
			wantFTS:        `("亜利沙" AND "golang")`,
			wantHasFTS:     true,
			wantLikeConds:  0,
			wantLikeParams: 0,
		},
		{
			name:           "Mixed: short + long tokens (杏奈 亜利沙)",
			input:          "杏奈 亜利沙",
			excludeInput:   "",
			searchType:     "and",
			wantFTS:        `"亜利沙"`,
			wantHasFTS:     true,
			wantLikeConds:  1,
			wantLikeParams: 1,
		},
		{
			name:           "All tokens short (杏奈 春香)",
			input:          "杏奈 春香",
			excludeInput:   "",
			searchType:     "and",
			wantFTS:        "",
			wantHasFTS:     false,
			wantLikeConds:  2,
			wantLikeParams: 2,
		},
		{
			name:           "Single long token",
			input:          "亜利沙",
			excludeInput:   "",
			searchType:     "and",
			wantFTS:        `"亜利沙"`,
			wantHasFTS:     true,
			wantLikeConds:  0,
			wantLikeParams: 0,
		},
		{
			name:           "Single short token",
			input:          "杏奈",
			excludeInput:   "",
			searchType:     "and",
			wantFTS:        "",
			wantHasFTS:     false,
			wantLikeConds:  1,
			wantLikeParams: 1,
		},
		{
			name:           "Mixed with exclude (long exclude)",
			input:          "杏奈",
			excludeInput:   "リツイート",
			searchType:     "and",
			wantFTS:        "",
			wantHasFTS:     false,
			wantLikeConds:  2,  // LIKE for 杏奈 + NOT LIKE for リツイート
			wantLikeParams: 2,
		},
		{
			name:           "FTS with short exclude",
			input:          "亜利沙",
			excludeInput:   "杏奈",
			searchType:     "and",
			wantFTS:        `"亜利沙"`,
			wantHasFTS:     true,
			wantLikeConds:  1,  // NOT LIKE for 杏奈
			wantLikeParams: 1,
		},
		{
			name:           "Empty input",
			input:          "",
			excludeInput:   "",
			searchType:     "and",
			wantFTS:        "",
			wantHasFTS:     false,
			wantLikeConds:  0,
			wantLikeParams: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildHybridSearchQuery(tt.input, tt.excludeInput, tt.searchType)
			if got.FTSQuery != tt.wantFTS {
				t.Errorf("FTSQuery = %q; want %q", got.FTSQuery, tt.wantFTS)
			}
			if got.HasFTS != tt.wantHasFTS {
				t.Errorf("HasFTS = %v; want %v", got.HasFTS, tt.wantHasFTS)
			}
			if len(got.LikeConds) != tt.wantLikeConds {
				t.Errorf("LikeConds count = %d; want %d (conds: %v)", len(got.LikeConds), tt.wantLikeConds, got.LikeConds)
			}
			if len(got.LikeParams) != tt.wantLikeParams {
				t.Errorf("LikeParams count = %d; want %d (params: %v)", len(got.LikeParams), tt.wantLikeParams, got.LikeParams)
			}
		})
	}
}

