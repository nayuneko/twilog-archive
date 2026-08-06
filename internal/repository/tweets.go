package repository

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"twilog-archive/internal/form"
	"twilog-archive/internal/model"
)

func Search(db *sqlx.DB, req *form.SearchRequest) ([]model.TweetsWithName, error) {
	word := strings.TrimSpace(req.SearchWord)
	if word == "" {
		return nil, nil
	}

	// 1. FTS5 Trigram MATCH 検索
	ftsQuery := "\"" + strings.ReplaceAll(word, "\"", "\"\"") + "\""
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
	if err := db.Select(&result, qFTS, paramsFTS...); err == nil {
		return result, nil
	}

	// 2. フォールバック: 通常の LIKE 検索
	paramsLIKE := []interface{}{"%" + word + "%"}
	qLIKE := "SELECT t.*, u.name FROM tweets t LEFT JOIN users u ON t.user_id = u.id WHERE full_text LIKE ?"
	if req.Pagination.LastID != nil {
		qLIKE += " AND t.id < ?"
		paramsLIKE = append(paramsLIKE, *req.Pagination.LastID)
	}
	qLIKE += " ORDER BY t.id DESC LIMIT 50"

	if err := db.Select(&result, qLIKE, paramsLIKE...); err != nil {
		return nil, err
	}
	return result, nil
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
