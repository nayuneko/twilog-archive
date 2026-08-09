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

	typeCond := BuildTypeFilterSQL(req.TweetTypeFilter)

	var total int
	qCountFTS := `SELECT COUNT(*) FROM tweets t JOIN tweets_fts f ON t.id = f.rowid WHERE tweets_fts MATCH ?`
	paramsCountFTS := []interface{}{ftsQuery}
	if typeCond != "" {
		qCountFTS += " AND " + typeCond
	}
	_ = db.Get(&total, qCountFTS, paramsCountFTS...)

	qFTS := `SELECT t.*, u.name 
	         FROM tweets t 
	         JOIN tweets_fts f ON t.id = f.rowid 
	         LEFT JOIN users u ON t.user_id = u.id 
	         WHERE tweets_fts MATCH ?`
	paramsFTS := []interface{}{ftsQuery}
	if typeCond != "" {
		qFTS += " AND " + typeCond
	}

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

	whereClause := likeCond
	if typeCond != "" {
		whereClause += " AND " + typeCond
	}

	qCountLIKE := "SELECT COUNT(*) FROM tweets t WHERE " + whereClause
	_ = db.Get(&total, qCountLIKE, likeParams...)

	qLIKE := "SELECT t.*, u.name FROM tweets t LEFT JOIN users u ON t.user_id = u.id WHERE " + whereClause
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

func Latest(db *sqlx.DB, lastID *string, filter form.TweetTypeFilter) ([]model.TweetsWithName, error) {
	var params []interface{}
	var conds []string

	if lastID != nil {
		conds = append(conds, "id < ?")
		params = append(params, *lastID)
	}

	typeCond := BuildTypeFilterSQL(filter)
	if typeCond != "" {
		conds = append(conds, typeCond)
	}

	q := "SELECT t.*, u.name FROM tweets t left join users u on t.user_id = u.id"
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " order by id desc limit 100"

	var result []model.TweetsWithName
	if err := db.Select(&result, q, params...); err != nil {
		return nil, err
	}
	return result, nil
}

func FindByDates(db *sqlx.DB, date string, filter form.TweetTypeFilter) ([]model.TweetsWithName, error) {
	var params []interface{}
	conds := []string{"created_date = ?"}
	params = append(params, date)

	typeCond := BuildTypeFilterSQL(filter)
	if typeCond != "" {
		conds = append(conds, typeCond)
	}

	q := "SELECT t.*, u.name FROM tweets t left join users u on t.user_id = u.id WHERE " + strings.Join(conds, " AND ") + " order by id desc limit 100"

	var result []model.TweetsWithName
	if err := db.Select(&result, q, params...); err != nil {
		return nil, err
	}
	return result, nil
}
