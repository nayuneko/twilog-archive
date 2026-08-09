package repository

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"twilog-archive/internal/form"
	"twilog-archive/internal/model"
)

func Search(db *sqlx.DB, req *form.SearchRequest) ([]model.TweetsWithName, int, error) {
	word := strings.TrimSpace(req.SearchWord)
	if word == "" {
		return nil, 0, nil
	}

	// 1. FTS5 Trigram MATCH 検索
	ftsQuery := BuildFTSQuery(word, req.ExcludeWord, req.SearchType)
	if ftsQuery == "" {
		return nil, 0, nil
	}

	var total int
	qCountFTS := `SELECT COUNT(*) FROM tweets_fts WHERE tweets_fts MATCH ?`
	_ = db.Get(&total, qCountFTS, ftsQuery)

	qFTS := `SELECT t.*, u.name 
	         FROM tweets t 
	         JOIN tweets_fts f ON t.id = f.rowid 
	         LEFT JOIN users u ON t.user_id = u.id 
	         WHERE tweets_fts MATCH ?`
	paramsFTS := []interface{}{ftsQuery}

	if req.Pagination.LastID != nil {
		qFTS += " AND t.id < ?"
		paramsFTS = append(paramsFTS, *req.Pagination.LastID)
	}
	qFTS += " ORDER BY t.id DESC LIMIT 50"

	var result []model.TweetsWithName
	err := db.Select(&result, qFTS, paramsFTS...)
	if err == nil {
		return result, total, nil
	}

	// テーブル不在 / FTS5未サポートの場合のみ LIKE 検索にフォールバック
	errMsg := err.Error()
	if !strings.Contains(errMsg, "no such table") && !strings.Contains(errMsg, "fts5") {
		return nil, 0, err
	}

	// 2. フォールバック: 通常の LIKE 検索
	likeCond, likeParams := BuildLikeQuery(word, req.ExcludeWord, req.SearchType)
	if likeCond == "" {
		return nil, 0, nil
	}

	qCountLIKE := "SELECT COUNT(*) FROM tweets t WHERE " + likeCond
	_ = db.Get(&total, qCountLIKE, likeParams...)

	qLIKE := "SELECT t.*, u.name FROM tweets t LEFT JOIN users u ON t.user_id = u.id WHERE " + likeCond
	if req.Pagination.LastID != nil {
		qLIKE += " AND t.id < ?"
		likeParams = append(likeParams, *req.Pagination.LastID)
	}
	qLIKE += " ORDER BY t.id DESC LIMIT 50"

	if err := db.Select(&result, qLIKE, likeParams...); err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func Latest(db *sqlx.DB, lastID *string) ([]model.TweetsWithName, error) {
	var params []interface{}
	q := "SELECT t.*, u.name FROM tweets t left join users u on t.user_id = u.id"
	if lastID != nil {
		q += " WHERE id < ?"
		params = append(params, *lastID)
	}
	q += " order by id desc limit 100"
	var result []model.TweetsWithName
	if err := db.Select(&result, q, params...); err != nil {
		return nil, err
	}
	return result, nil
}

func FindByDates(db *sqlx.DB, date string) ([]model.TweetsWithName, error) {
	q := "SELECT t.*, u.name FROM tweets t left join users u on t.user_id = u.id"
	q += " WHERE created_date = ? order by id desc limit 100"
	var result []model.TweetsWithName
	if err := db.Select(&result, q, date); err != nil {
		return nil, err
	}
	return result, nil
}
