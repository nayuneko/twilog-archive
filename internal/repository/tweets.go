package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"twilog-archive/internal/form"
	"twilog-archive/internal/model"
)

func Search(db *sqlx.DB, req *form.SearchRequest) ([]model.TweetsWithName, error) {
	params := []interface{}{fmt.Sprintf("%%%s%%", req.SearchWord)}
	/*
			q := `SELECT * FROM tweets
		JOIN (SELECT id FROM tweets WHERE match text against (? in boolean mode)`
	*/
	q := "SELECT t.*, u.name FROM tweets t left join users u on t.user_id = u.id"
	q += " WHERE full_text like ?"
	if req.Pagination.LastID != nil {
		q += " AND id < ?"
		params = append(params, *req.Pagination.LastID)
	}
	/*
			q += ` ORDER BY id DESC LIMIT 100) t ON t.id = s.id
		ORDER BY id DESC`
	*/
	q += ` ORDER BY id DESC`
	var result []model.TweetsWithName
	if err := db.Select(&result, q, params...); err != nil {
		return nil, err
	}
	return result, nil
	//	return getTweets(db, q, params...)
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
