package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"twilog-archive/internal/config"
	"twilog-archive/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

var (
	re      = regexp.MustCompile(`https://x\.com/([^/]+)/status/`)
	reMedia = regexp.MustCompile(`https://(?:x|twitter)\.com/[^/]+/status/\d+/(?:photo|video)/\d+`)
)

func setupDB(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA cache_size = -64000;",
		"PRAGMA temp_store = MEMORY;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("PRAGMA設定失敗 (%s): %w", p, err)
		}
	}
	return nil
}

func main() {
	csvPath := config.CSVFile
	if len(os.Args) >= 2 && os.Args[1] != "" {
		csvPath = os.Args[1]
	}
	dbPath := config.DBFile

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	if err := setupDB(db); err != nil {
		log.Printf("DB PRAGMA設定警告: %v", err)
	}

	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("CSVオープン失敗: %v", err)
	}
	defer f.Close()

	var csvReader io.Reader = f
	if strings.HasSuffix(csvPath, ".gz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			log.Fatalf("GZIPオープン失敗: %v", err)
		}
		defer gzr.Close()
		csvReader = gzr
	}

	reader := csv.NewReader(csvReader)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // 可変長レコード対応

	startTime := time.Now()
	log.Printf("Twilog CSV インポート開始: %s", csvPath)

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("トランザクション開始失敗: %v", err)
	}
	defer tx.Rollback()

	userStmt, err := db.Prepare("SELECT id FROM users WHERE screen_name = ? LIMIT 1")
	if err != nil {
		log.Fatalf("ユーザー検索ステートメント作成失敗: %v", err)
	}
	defer userStmt.Close()

	q := `INSERT INTO tweets (id, created_at, created_date, screen_name, full_text, retweeted, log_type, user_id, embed_media_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET full_text = CASE WHEN length(excluded.full_text) > length(tweets.full_text) THEN excluded.full_text ELSE tweets.full_text END, screen_name = excluded.screen_name, retweeted = excluded.retweeted, embed_media_url = COALESCE(excluded.embed_media_url, tweets.embed_media_url), user_id = COALESCE(tweets.user_id, excluded.user_id)`
	stmt, err := tx.Prepare(q)
	if err != nil {
		log.Fatalf("ステートメント作成失敗: %v", err)
	}
	defer stmt.Close()

	var insertedCount, skippedCount int
	batchSize := 5000

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skippedCount++
			continue
		}

		if len(record) < 5 {
			skippedCount++
			continue // 欠落行はスキップ
		}

		idStr := record[0]
		url := record[1]
		dateStr := record[2]
		text := record[3]
		logType := record[4]

		// ログタイプは1:ツイート(RT含む)、2:いいね、3:ブックマーク
		if logType != model.TwilogLogTypeTweet {
			skippedCount++
			continue
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			skippedCount++
			continue
		}

		// 投稿日時のパース
		createdAt, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			skippedCount++
			continue
		}

		// RT判定
		match := re.FindStringSubmatch(url)
		if len(match) < 2 {
			skippedCount++
			continue
		}
		screenName := match[1]
		if screenName == config.MyScreenName && strings.HasPrefix(text, "RT @") {
			rest := text[4:]
			if idx := strings.Index(rest, ":"); idx != -1 {
				screenName = rest[:idx]
			}
		}
		retweeted := screenName != config.MyScreenName

		var userID *int64
		var userRowID int64
		if err := userStmt.QueryRow(screenName).Scan(&userRowID); err == nil {
			userID = &userRowID
		}

		var embedMediaURL *string
		if matches := reMedia.FindAllString(text, -1); len(matches) > 0 {
			lastMatch := matches[len(matches)-1]
			if strings.HasSuffix(strings.TrimSpace(text), lastMatch) {
				embedMediaURL = &lastMatch
			}
		}

		createdDate := createdAt.In(time.Local).Format("20060102")

		_, err = stmt.Exec(
			id,
			createdAt.Format(time.RFC3339),
			createdDate,
			screenName,
			text,
			retweeted,
			model.LogTypeTwilog,
			userID,
			embedMediaURL,
		)
		if err != nil {
			log.Printf("INSERT失敗: id=%d: %v", id, err)
			skippedCount++
			continue
		}

		insertedCount++
		if insertedCount%batchSize == 0 {
			if err := stmt.Close(); err != nil {
				log.Fatalf("ステートメントクローズ失敗: %v", err)
			}
			if err := tx.Commit(); err != nil {
				log.Fatalf("コミット失敗: %v", err)
			}

			tx, err = db.Begin()
			if err != nil {
				log.Fatalf("トランザクション開始失敗: %v", err)
			}
			stmt, err = tx.Prepare(q)
			if err != nil {
				log.Fatalf("ステートメント作成失敗: %v", err)
			}

			log.Printf("進捗: %d 件処理済み...", insertedCount)
		}
	}

	if err := stmt.Close(); err != nil {
		log.Fatalf("ステートメントクローズ失敗: %v", err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("コミット失敗: %v", err)
	}

	log.Printf("インポート完了！ 処理件数: %d 件, スキップ件数: %d 件 (所要時間: %v)", insertedCount, skippedCount, time.Since(startTime))
}

