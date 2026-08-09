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

	hybrid := BuildHybridSearchQuery(word, req.ExcludeWord, req.SearchType)
	// 検索トークンが1つも無い場合
	if !hybrid.HasFTS && len(hybrid.LikeConds) == 0 {
		return nil, 0, nil
	}

	typeCond := BuildTypeFilterSQL(req.TweetTypeFilter)
	hashtagCond, hashtagParams := BuildHashtagExcludeSQL(word, req.ExcludeWord)

	// FTS5 対象トークンがある場合: ハイブリッド検索（FTS5 MATCH + LIKE 補助条件）
	if hybrid.HasFTS {
		result, total, err := searchWithFTS(db, req, hybrid, typeCond, hashtagCond, hashtagParams)
		if err == nil {
			return result, total, nil
		}
		// テーブル不在 / FTS5未サポートの場合のみ LIKE 検索にフォールバック
		errMsg := err.Error()
		if !strings.Contains(errMsg, "no such table") && !strings.Contains(errMsg, "fts5") {
			return nil, 0, err
		}
	}

	// FTS5 対象トークンが無い場合、または FTS5 フォールバック: 全て LIKE 検索
	return searchWithLIKE(db, req, word, typeCond, hashtagCond, hashtagParams)
}

// searchWithFTS は FTS5 MATCH + LIKE 補助条件を組み合わせたハイブリッド検索を実行する
func searchWithFTS(db *sqlx.DB, req *form.SearchRequest, hybrid HybridSearchResult, typeCond string, hashtagCond string, hashtagParams []interface{}) ([]model.TweetsWithName, int, error) {
	var extraWhere []string
	var extraParams []interface{}

	// 2文字以下トークンの LIKE 条件を追加
	for _, cond := range hybrid.LikeConds {
		extraWhere = append(extraWhere, cond)
	}
	extraParams = append(extraParams, hybrid.LikeParams...)

	if typeCond != "" {
		extraWhere = append(extraWhere, typeCond)
	}
	if hashtagCond != "" {
		extraWhere = append(extraWhere, hashtagCond)
		extraParams = append(extraParams, hashtagParams...)
	}

	whereStr := ""
	if len(extraWhere) > 0 {
		whereStr = " AND " + strings.Join(extraWhere, " AND ")
	}

	// COUNT クエリ
	var total int
	qCount := `SELECT COUNT(*) FROM tweets t JOIN tweets_fts f ON t.id = f.rowid WHERE tweets_fts MATCH ?` + whereStr
	paramsCount := append([]interface{}{hybrid.FTSQuery}, extraParams...)
	_ = db.Get(&total, qCount, paramsCount...)

	// データ取得クエリ
	qFTS := `SELECT t.*, u.name 
	         FROM tweets t 
	         JOIN tweets_fts f ON t.id = f.rowid 
	         LEFT JOIN users u ON t.user_id = u.id 
	         WHERE tweets_fts MATCH ?` + whereStr

	paramsFTS := append([]interface{}{hybrid.FTSQuery}, extraParams...)

	if req.Pagination.LastID != nil {
		qFTS += " AND t.id < ?"
		paramsFTS = append(paramsFTS, *req.Pagination.LastID)
	}
	qFTS += " ORDER BY t.id DESC LIMIT 50"

	var result []model.TweetsWithName
	if err := db.Select(&result, qFTS, paramsFTS...); err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

// searchWithLIKE は全て LIKE 条件で検索を行う（FTS5 未使用 or フォールバック）
func searchWithLIKE(db *sqlx.DB, req *form.SearchRequest, word string, typeCond string, hashtagCond string, hashtagParams []interface{}) ([]model.TweetsWithName, int, error) {
	likeCond, likeParams := BuildLikeQuery(word, req.ExcludeWord, req.SearchType)
	if likeCond == "" {
		return nil, 0, nil
	}

	var likeWhere []string
	likeWhere = append(likeWhere, likeCond)
	if typeCond != "" {
		likeWhere = append(likeWhere, typeCond)
	}
	if hashtagCond != "" {
		likeWhere = append(likeWhere, hashtagCond)
		likeParams = append(likeParams, hashtagParams...)
	}

	whereClause := strings.Join(likeWhere, " AND ")

	var total int
	qCountLIKE := "SELECT COUNT(*) FROM tweets t WHERE " + whereClause
	_ = db.Get(&total, qCountLIKE, likeParams...)

	qLIKE := "SELECT t.*, u.name FROM tweets t LEFT JOIN users u ON t.user_id = u.id WHERE " + whereClause
	if req.Pagination.LastID != nil {
		qLIKE += " AND t.id < ?"
		likeParams = append(likeParams, *req.Pagination.LastID)
	}
	qLIKE += " ORDER BY t.id DESC LIMIT 50"

	var result []model.TweetsWithName
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
