package repository

import (
	"github.com/jmoiron/sqlx"
	"twilog-archive/internal/model"
)

func HashtagsFindByTweetID(db *sqlx.DB, tweetID int64) ([]model.Hashtags, error) {
	q := "select * from hashtags where tweet_id = ? order by tag_index"
	var result []model.Hashtags
	if err := db.Select(&result, q, tweetID); err != nil {
		return nil, err
	}
	return result, nil
}
