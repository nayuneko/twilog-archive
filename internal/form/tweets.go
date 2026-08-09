package form

type (
	Pagination struct {
		LastID *string `json:"last_id"`
	}
	SearchRequest struct {
		Pagination
		SearchWord  string `json:"search_word"`
		ExcludeWord string `json:"exclude_word"`
		SearchType  string `json:"search_type"`
	}
)
