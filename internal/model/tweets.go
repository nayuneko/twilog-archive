package model

import "time"

type Tweets struct {
	ID            int64     `db:"id"`
	CreatedAt     time.Time `db:"created_at"`
	CreatedDate   string    `db:"created_date"`
	ScreenName    string    `db:"screen_name"`
	FullText      string    `db:"full_text"`
	Retweeted     bool      `db:"retweeted"`
	Replied       bool      `db:"replied"`
	LogType       string    `db:"log_type"`
	UserID        *int64    `db:"user_id"`
	EmbedMediaURL *string   `db:"embed_media_url"`
}

type TweetsWithName struct {
	Tweets
	Name *string `db:"name"`
}
