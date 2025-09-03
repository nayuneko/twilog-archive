package handler

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"net/http"
	"os"
	"twilog-archive/internal/constant"
	"twilog-archive/internal/utils"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func Calendar(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		calendarData, err := func() (utils.CalendarData, error) {
			if fileExists(constant.JsonCalendar) {
				var calendarData utils.CalendarData
				f, _ := os.Open(constant.JsonCalendar)
				defer f.Close()
				if err := json.NewDecoder(f).Decode(&calendarData); err != nil {
					return nil, err
				}
				return calendarData, nil
			}
			return utils.MakeCalendarData(db)
		}()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, calendarData)
	}
}
