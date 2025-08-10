package repository

import (
	"github.com/jmoiron/sqlx"
	"twilog-archive/internal/model"
)

func URLsFindByTweetID(db *sqlx.DB, tweetID int64) ([]model.URLs, error) {
	q := "select * from urls where tweet_id = ? order by url_index"
	var result []model.URLs
	if err := db.Select(&result, q, tweetID); err != nil {
		return nil, err
	}
	return result, nil
}
