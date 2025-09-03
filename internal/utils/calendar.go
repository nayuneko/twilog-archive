package utils

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"os"
	"strconv"
	"strings"
	"twilog-archive/internal/constant"
	"twilog-archive/internal/repository"
)

type CalendarData map[string]map[string][]int

func MakeCalendarData(db *sqlx.DB) (CalendarData, error) {
	dates, err := repository.AllDates(db)
	if err != nil {
		return nil, err
	}
	cal := make(CalendarData)
	for _, date := range dates {
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
	f, _ := os.Create(constant.JsonCalendar)
	defer f.Close()
	if err := json.NewEncoder(f).Encode(cal); err != nil {
		return nil, err
	}
	return cal, nil
}
