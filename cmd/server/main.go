package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"twilog-archive/internal/config"
	"twilog-archive/internal/server"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = config.DBFile
	}

	// SQLiteに接続
	db, err := sqlx.Open("sqlite3", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 共通エンジンを作成
	e := server.NewEngine(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "10069"
	}

	// サーバー起動
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
