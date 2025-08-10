package model

type URLs struct {
	TweetID    int64  `db:"tweet_id"`
	Index      int    `db:"url_index"`
	URL        string `db:"url"`
	ExpandURL  string `db:"expanded_url"`
	DisplayURL string `db:"display_url"`
}
