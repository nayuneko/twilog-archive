package handler

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"strconv"
	"strings"
	"twilog-archive/internal/repository"
)

type CalendarData map[string]map[string][]int

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func Calendar(db *sqlx.DB) echo.HandlerFunc {
	cacheName := "./data/static/calendar.json"
	return func(c echo.Context) error {
		var calendarData CalendarData
		if fileExists(cacheName) {
			f, _ := os.Open(cacheName)
			defer f.Close()
			if err := json.NewDecoder(f).Decode(&calendarData); err != nil {
				return err
			}
		} else {
			dates, err := repository.AllDates(db)
			if err != nil {
				return err
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
			f, _ := os.Create(cacheName)
			defer f.Close()
			json.NewEncoder(f).Encode(cal)
			calendarData = cal
		}
		return c.JSON(http.StatusOK, calendarData)
	}
}
