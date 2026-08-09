package form

type (
	Pagination struct {
		LastID *string `json:"last_id"`
	}
	TweetTypeFilter struct {
		IncludeNormal bool `json:"include_normal"`
		IncludeReply  bool `json:"include_reply"`
		IncludeRT     bool `json:"include_rt"`
	}
	SearchRequest struct {
		Pagination
		TweetTypeFilter
		SearchWord  string `json:"search_word"`
		ExcludeWord string `json:"exclude_word"`
		SearchType  string `json:"search_type"`
	}
)
