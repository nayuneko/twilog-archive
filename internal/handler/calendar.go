package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"

	"twilog-archive/internal/repository"
)

func Calendar(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		calendarData, err := repository.GetCalendarData(db)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, calendarData)
	}
}
