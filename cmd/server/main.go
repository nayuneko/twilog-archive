package main

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
	"twilog-archive/internal/constant"
	"twilog-archive/internal/handler"
	"twilog-archive/web"
)

func main() {
	// SQLiteに接続
	db, err := sqlx.Open("sqlite3", constant.DBFile+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Echoのインスタンス作る
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

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

	// フロントエンド静的ファイル配信 & SPAフォールバック
	distFS, err := fs.Sub(web.Dist, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distFS))
		e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}

			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				fileServer.ServeHTTP(w, r)
				return
			}

			f, err := distFS.Open(path)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})))
	}

	// サーバー起動
	if err := e.Start(":10069"); err != nil {
		log.Fatal(err)
	}
}

