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

func MediaFindByTweetIDs(db *sqlx.DB, tweetIDs []int64) (map[int64][]model.Media, error) {
	resultMap := make(map[int64][]model.Media)
	if len(tweetIDs) == 0 {
		return resultMap, nil
	}
	query, args, err := sqlx.In("SELECT * FROM media WHERE tweet_id IN (?) ORDER BY tweet_id, media_index", tweetIDs)
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)
	var mediaList []model.Media
	if err := db.Select(&mediaList, query, args...); err != nil {
		return nil, err
	}
	for _, m := range mediaList {
		resultMap[m.TweetID] = append(resultMap[m.TweetID], m)
	}
	return resultMap, nil
}
