package model

type Hashtags struct {
	TweetID int64  `db:"tweet_id"`
	Index   int    `db:"tag_index"`
	Tag     string `db:"tag"`
}
