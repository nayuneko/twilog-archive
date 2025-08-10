package form

type (
	Pagination struct {
		LastID *string `json:"last_id"`
	}
	SearchRequest struct {
		Pagination
		SearchWord string `json:"search_word"`
	}
)
