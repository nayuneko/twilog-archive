package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"twilog-archive/internal/config"
	"twilog-archive/internal/text"
)

func main() {
	dbPath := config.DBFile
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	log.Printf("HTMLエンティティ修正開始: %s", dbPath)
	startTime := time.Now()

	rows, err := db.Query("SELECT id, full_text FROM tweets WHERE full_text LIKE '%&%'")
	if err != nil {
		log.Fatalf("SELECTクエリ失敗: %v", err)
	}
	defer rows.Close()

	type item struct {
		id   int64
		text string
	}
	var targets []item

	for rows.Next() {
		var id int64
		var textVal string
		if err := rows.Scan(&id, &textVal); err != nil {
			log.Fatalf("Scan失敗: %v", err)
		}
		unescaped := text.UnescapeHTML(textVal)
		if unescaped != textVal {
			targets = append(targets, item{id: id, text: unescaped})
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rowsエラー: %v", err)
	}

	log.Printf("修正対象件数: %d 件", len(targets))
	if len(targets) == 0 {
		log.Println("修正対象はありませんでした。")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("トランザクション開始失敗: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE tweets SET full_text = ? WHERE id = ?")
	if err != nil {
		log.Fatalf("UPDATEステートメント作成失敗: %v", err)
	}
	defer stmt.Close()

	updatedCount := 0
	for _, t := range targets {
		if _, err := stmt.Exec(t.text, t.id); err != nil {
			log.Fatalf("UPDATE実行失敗 (id=%d): %v", t.id, err)
		}
		updatedCount++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("コミット失敗: %v", err)
	}

	log.Printf("HTMLエンティティ修正完了！ 修正件数: %d 件 (所要時間: %v)", updatedCount, time.Since(startTime))
}
