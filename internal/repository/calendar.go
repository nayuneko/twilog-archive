package repository

import (
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type CalendarData map[string]map[string][]int

func AllDates(db *sqlx.DB) ([]string, error) {
	q := "select distinct created_date from tweets order by created_date desc"

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

func GetCalendarData(db *sqlx.DB) (CalendarData, error) {
	dates, err := AllDates(db)
	if err != nil {
		return nil, err
	}
	cal := make(CalendarData)
	for _, date := range dates {
		if len(date) < 8 {
			continue
		}
		y := date[:4]
		m := strings.TrimLeft(date[4:6], "0")
		d, _ := strconv.Atoi(date[6:8])
		if _, ok := cal[y]; !ok {
			cal[y] = make(map[string][]int)
		}
		if _, ok := cal[y][m]; !ok {
			cal[y][m] = make([]int, 0)
		}
		cal[y][m] = append(cal[y][m], d)
	}
	return cal, nil
}
