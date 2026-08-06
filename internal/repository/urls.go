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

func URLsFindByTweetIDs(db *sqlx.DB, tweetIDs []int64) (map[int64][]model.URLs, error) {
	resultMap := make(map[int64][]model.URLs)
	if len(tweetIDs) == 0 {
		return resultMap, nil
	}
	query, args, err := sqlx.In("SELECT * FROM urls WHERE tweet_id IN (?) ORDER BY tweet_id, url_index", tweetIDs)
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)
	var urlList []model.URLs
	if err := db.Select(&urlList, query, args...); err != nil {
		return nil, err
	}
	for _, u := range urlList {
		resultMap[u.TweetID] = append(resultMap[u.TweetID], u)
	}
	return resultMap, nil
}
