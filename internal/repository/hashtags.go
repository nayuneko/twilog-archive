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

func HashtagsFindByTweetIDs(db *sqlx.DB, tweetIDs []int64) (map[int64][]model.Hashtags, error) {
	resultMap := make(map[int64][]model.Hashtags)
	if len(tweetIDs) == 0 {
		return resultMap, nil
	}
	query, args, err := sqlx.In("SELECT * FROM hashtags WHERE tweet_id IN (?) ORDER BY tweet_id, tag_index", tweetIDs)
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)
	var hashtagList []model.Hashtags
	if err := db.Select(&hashtagList, query, args...); err != nil {
		return nil, err
	}
	for _, h := range hashtagList {
		resultMap[h.TweetID] = append(resultMap[h.TweetID], h)
	}
	return resultMap, nil
}
