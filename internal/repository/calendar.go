package repository

import (
	"github.com/jmoiron/sqlx"
)

func AllDates(db *sqlx.DB) ([]string, error) {
	q := "select created_date from tweets group by created_date order by created_date desc"

	var result []string
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
