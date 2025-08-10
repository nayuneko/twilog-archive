package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
	"twilog-archive/internal/constant"

	_ "github.com/mattn/go-sqlite3"
)

var (
	re = regexp.MustCompile(`https://x\.com/([^/]+)/status/`)
)

func main() {
	// ファイルパス・DBパスの設定
	csvPath := "./data/csv/nayuneko-250707.csv"
	dbPath := "./data/db/tweets.db"

	// SQLiteに接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// CSVオープン
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // 可変長レコード対応

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		if len(record) < 5 {
			continue // 欠落行はスキップ
		}

		idStr := record[0]
		url := record[1]
		dateStr := record[2]
		text := record[3]
		logType := record[4]

		// ログタイプは1:ツイート(RT含む)、2:いいね、3:ブックマーク
		if logType != "1" {
			continue
		}

		id, _ := strconv.ParseInt(idStr, 10, 64)

		// 投稿日時のパース
		createdAt, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			log.Printf("スキップ: 行 %s（日時パース失敗）: %v\n", idStr, err)
			continue
		}

		// RT判定
		var retweeted bool
		var screenName string
		match := re.FindStringSubmatch(url)
		if len(match) >= 2 {
			screenName = match[1]
			retweeted = screenName != constant.MyScreenName
		} else {
			fmt.Println("マッチしませんでした")
			continue
		}

		createdDate := createdAt.In(time.Local).Format("20060102")

		q := `INSERT OR IGNORE INTO tweets (id, created_at, created_date, screen_name, full_text, retweeted, log_type) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(
			q,
			id,
			createdAt.Format(time.RFC3339),
			createdDate,
			screenName,
			text,
			retweeted,
			constant.LogTypeTwilog,
		)
		if err != nil {
			log.Printf("スキップ: 行 %s（INSERT失敗）: %v\n", idStr, err)
		}
	}
	fmt.Println("インポート完了！")
}
