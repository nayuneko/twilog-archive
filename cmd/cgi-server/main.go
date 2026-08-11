package main

import (
	"log"
	"net/http/cgi"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"twilog-archive/internal/config"
	"twilog-archive/internal/server"
)

func main() {
	log.SetOutput(os.Stderr)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = os.Getenv("REDIRECT_DB_PATH")
	}
	if dbPath == "" {
		// XREA 環境のデフォルト絶対パス
		defaultPath := "/virtual/atkg3a/db/tweets.db"
		if _, err := os.Stat(defaultPath); err == nil {
			dbPath = defaultPath
		} else {
			dbPath = config.DBFile
		}
	}

	// Read-Only モードで SQLite に接続 (CGI環境でのロック問題防止)
	db, err := sqlx.Open("sqlite3", dbPath+"?mode=ro&_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 共通エンジンを取得
	e := server.NewEngine(db)

	// CGI として Serve
	if err := cgi.Serve(e); err != nil {
		log.Fatal(err)
	}
}
