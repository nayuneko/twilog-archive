package model

type Media struct {
	TweetID   int64  `db:"tweet_id"`
	Index     int    `db:"media_index"`
	MediaURL  string `db:"media_url"`
	MediaType string `db:"type"`
}
