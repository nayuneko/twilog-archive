package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"twilog-archive/internal/handler"
)

func main() {

	// DBパスの設定
	dbPath := "./data/db/tweets.db"

	// SQLiteに接続
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Echoのインスタンス作る
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ルーティング
	m := e.Group("/api/tweets")
	m.GET("/latest", handler.TweetsLatest(db))
	m.GET("/dates/:date", handler.TweetsDates(db))
	m.GET("/search/", handler.TweetsSearch(db))
	/*
		s := e.Group("/api/statuses")
		s.GET("/:status_id", router.StatusDetail(dbMap, statusRepo, entityRepo))
	*/
	e.GET("/api/calendar", handler.Calendar(db))

	// サーバー起動
	if err := e.Start(":10069"); err != nil {
		log.Fatal(err)
	}
}
