package model

type Users struct {
	ID           int64  `db:"id"`
	Name         string `db:"name"`
	ScreenName   string `db:"screen_name"`
	LastStatusID int64  `db:"last_status_id"`
}
