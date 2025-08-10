package repository

import (
	"github.com/jmoiron/sqlx"
	"twilog-archive/internal/model"
)

func MediaFindByTweetID(db *sqlx.DB, tweetID int64) ([]model.Media, error) {
	q := "select * from media where tweet_id = ? order by media_index"
	var result []model.Media
	if err := db.Select(&result, q, tweetID); err != nil {
		return nil, err
	}
	return result, nil
}
