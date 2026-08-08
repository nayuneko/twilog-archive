package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"twilog-archive/internal/config"
	"twilog-archive/internal/model"
	"twilog-archive/internal/xdata"
)

const (
	updateTwilog = false
	batchSize    = 500
)

type insertData struct {
	tweet    *model.Tweets
	users    []model.Users
	media    []model.Media
	urls     []model.URLs
	hashtags []model.Hashtags
}

func setupDB(db *sqlx.DB) error {
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

func insertBatch(tx *sql.Tx, batch []*insertData) (int64, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	var totalTweetsInserted int64

	// 1. tweets (chunk size 50, 10 params per row = 500 params)
	const tweetChunkSize = 50
	for i := 0; i < len(batch); i += tweetChunkSize {
		end := i + tweetChunkSize
		if end > len(batch) {
			end = len(batch)
		}
		chunk := batch[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT OR IGNORE INTO tweets VALUES ")
		args := make([]interface{}, 0, len(chunk)*10)

		for idx, d := range chunk {
			if idx > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("(?,?,?,?,?,?,?,?,?,?)")
			args = append(args,
				d.tweet.ID,
				d.tweet.CreatedAt.Format(time.RFC3339),
				d.tweet.CreatedAt.In(time.Local).Format("20060102"),
				d.tweet.ScreenName,
				d.tweet.FullText,
				d.tweet.Retweeted,
				d.tweet.Replied,
				model.LogTypeXArchive,
				d.tweet.UserID,
				d.tweet.EmbedMediaURL,
			)
		}

		res, err := tx.Exec(sb.String(), args...)
		if err != nil {
			return 0, fmt.Errorf("tweetsのバッチ追加に失敗: %w", err)
		}
		rows, _ := res.RowsAffected()
		totalTweetsInserted += rows
	}

	// 2. users (chunk size 100, 3 params per row = 300 params)
	var allUsers []model.Users
	for _, d := range batch {
		allUsers = append(allUsers, d.users...)
	}
	if len(allUsers) > 0 {
		const userChunkSize = 100
		for i := 0; i < len(allUsers); i += userChunkSize {
			end := i + userChunkSize
			if end > len(allUsers) {
				end = len(allUsers)
			}
			chunk := allUsers[i:end]

			var sb strings.Builder
			sb.WriteString("INSERT OR IGNORE INTO users (id, name, last_status_id) VALUES ")
			args := make([]interface{}, 0, len(chunk)*3)

			for idx, u := range chunk {
				if idx > 0 {
					sb.WriteString(",")
				}
				sb.WriteString("(?,?,?)")
				args = append(args, u.ID, u.Name, u.LastStatusID)
			}
			sb.WriteString(" ON CONFLICT(id) DO UPDATE SET name = excluded.name, last_status_id = excluded.last_status_id WHERE excluded.last_status_id > users.last_status_id")

			if _, err := tx.Exec(sb.String(), args...); err != nil {
				return 0, fmt.Errorf("usersのバッチ追加に失敗: %w", err)
			}
		}
	}

	// 3. media (chunk size 100, 4 params per row = 400 params)
	var allMedia []model.Media
	for _, d := range batch {
		allMedia = append(allMedia, d.media...)
	}
	if len(allMedia) > 0 {
		const mediaChunkSize = 100
		for i := 0; i < len(allMedia); i += mediaChunkSize {
			end := i + mediaChunkSize
			if end > len(allMedia) {
				end = len(allMedia)
			}
			chunk := allMedia[i:end]

			var sb strings.Builder
			sb.WriteString("INSERT OR IGNORE INTO media VALUES ")
			args := make([]interface{}, 0, len(chunk)*4)

			for idx, m := range chunk {
				if idx > 0 {
					sb.WriteString(",")
				}
				sb.WriteString("(?,?,?,?)")
				args = append(args, m.TweetID, m.Index, m.MediaURL, m.MediaType)
			}

			if _, err := tx.Exec(sb.String(), args...); err != nil {
				return 0, fmt.Errorf("mediaのバッチ追加に失敗: %w", err)
			}
		}
	}

	// 4. urls (chunk size 100, 5 params per row = 500 params)
	var allURLs []model.URLs
	for _, d := range batch {
		allURLs = append(allURLs, d.urls...)
	}
	if len(allURLs) > 0 {
		const urlChunkSize = 100
		for i := 0; i < len(allURLs); i += urlChunkSize {
			end := i + urlChunkSize
			if end > len(allURLs) {
				end = len(allURLs)
			}
			chunk := allURLs[i:end]

			var sb strings.Builder
			sb.WriteString("INSERT OR IGNORE INTO urls VALUES ")
			args := make([]interface{}, 0, len(chunk)*5)

			for idx, u := range chunk {
				if idx > 0 {
					sb.WriteString(",")
				}
				sb.WriteString("(?,?,?,?,?)")
				args = append(args, u.TweetID, u.Index, u.URL, u.ExpandURL, u.DisplayURL)
			}

			if _, err := tx.Exec(sb.String(), args...); err != nil {
				return 0, fmt.Errorf("urlsのバッチ追加に失敗: %w", err)
			}
		}
	}

	// 5. hashtags (chunk size 100, 3 params per row = 300 params)
	var allHashtags []model.Hashtags
	for _, d := range batch {
		allHashtags = append(allHashtags, d.hashtags...)
	}
	if len(allHashtags) > 0 {
		const hashtagChunkSize = 100
		for i := 0; i < len(allHashtags); i += hashtagChunkSize {
			end := i + hashtagChunkSize
			if end > len(allHashtags) {
				end = len(allHashtags)
			}
			chunk := allHashtags[i:end]

			var sb strings.Builder
			sb.WriteString("INSERT OR IGNORE INTO hashtags VALUES ")
			args := make([]interface{}, 0, len(chunk)*3)

			for idx, h := range chunk {
				if idx > 0 {
					sb.WriteString(",")
				}
				sb.WriteString("(?,?,?)")
				args = append(args, h.TweetID, h.Index, h.Tag)
			}

			if _, err := tx.Exec(sb.String(), args...); err != nil {
				return 0, fmt.Errorf("hashtagsのバッチ追加に失敗: %w", err)
			}
		}
	}

	return totalTweetsInserted, nil
}

func createTweets(t *xdata.Tweet) *model.Tweets {
	screenName := config.MyScreenName
	fullText := t.FullText

	// RT @xxxx: から始まるツイートの場合RT
	if strings.HasPrefix(t.FullText, "RT @") {
		// "RT @" を取り除く
		rest := t.FullText[4:]
		// ":" の位置を探す
		if idx := strings.Index(rest, ":"); idx != -1 {
			screenName = rest[:idx]
			fullText = rest[idx+2:]
		}
	}
	retweeted := screenName != config.MyScreenName

	tweets := &model.Tweets{
		ID:         int64(t.ID),
		CreatedAt:  t.CreatedAt.Time(),
		ScreenName: screenName,
		FullText:   fullText,
		Retweeted:  retweeted,
		Replied:    t.InReplyToUserID != nil,
	}
	return tweets
}

func createUsers(t *xdata.Tweet, tweets *model.Tweets) []model.Users {
	if len(t.Entities.UserMentions) == 0 {
		return nil
	}
	r := make([]model.Users, 0, len(t.Entities.UserMentions))
	for _, u := range t.Entities.UserMentions {
		uid := int64(u.ID)
		r = append(r, model.Users{
			ID:           uid,
			Name:         u.Name,
			LastStatusID: int64(t.ID),
		})
		if tweets.Retweeted && tweets.ScreenName == u.ScreenName {
			tweets.UserID = &uid
		}
	}
	return r
}

func createMedia(t *xdata.Tweet, tweets *model.Tweets) []model.Media {
	media := func() []xdata.Media {
		if len(t.ExtendedEntities.Media) > 0 {
			return t.ExtendedEntities.Media
		}
		if len(t.Entities.Media) > 0 {
			return t.Entities.Media
		}
		return nil
	}()
	if media == nil {
		return nil
	}
	r := make([]model.Media, 0, len(media))
	for idx, m := range media {
		r = append(r, model.Media{
			TweetID:   int64(t.ID),
			Index:     idx + 1,
			MediaURL:  m.MediaURLHttps,
			MediaType: m.Type,
		})
	}
	m0 := media[0]
	if m0.URL != "" && strings.HasSuffix(tweets.FullText, m0.URL) {
		tweets.EmbedMediaURL = &m0.URL
	} else if m0.ExpandedURL != "" && strings.HasSuffix(tweets.FullText, m0.ExpandedURL) {
		tweets.EmbedMediaURL = &m0.ExpandedURL
	}
	return r
}

func createUrls(t *xdata.Tweet) []model.URLs {
	if len(t.Entities.URLs) == 0 {
		return nil
	}
	r := make([]model.URLs, 0, len(t.Entities.URLs))
	for idx, u := range t.Entities.URLs {
		r = append(r, model.URLs{
			TweetID:    int64(t.ID),
			Index:      idx + 1,
			URL:        u.URL,
			ExpandURL:  u.ExpandedURL,
			DisplayURL: u.DisplayURL,
		})
	}
	return r
}

func createHashtags(t *xdata.Tweet) []model.Hashtags {
	if len(t.Entities.Hashtags) == 0 {
		return nil
	}
	r := make([]model.Hashtags, 0, len(t.Entities.Hashtags))
	for idx, h := range t.Entities.Hashtags {
		r = append(r, model.Hashtags{
			TweetID: int64(t.ID),
			Index:   idx + 1,
			Tag:     h.Text,
		})
	}
	return r
}

// importTweetsFromReader io.Reader (JSまたはJSONデータ) からツイートをストリーミングデコードしてインポート
func importTweetsFromReader(db *sqlx.DB, r io.Reader) (int64, error) {
	jsReader := xdata.NewStripJSPrefixReader(r)
	dec := json.NewDecoder(jsReader)

	t, err := dec.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, nil
		}
		return 0, fmt.Errorf("JSONトークンの読み込み失敗: %w", err)
	}

	delim, ok := t.(json.Delim)
	if !ok || delim != '[' {
		return 0, fmt.Errorf("JSON配列の開始 '[' が見つかりません: %v", t)
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	batch := make([]*insertData, 0, batchSize)
	var totalCount int64

	for dec.More() {
		var tw xdata.TweetWrapper
		if err := dec.Decode(&tw); err != nil {
			return 0, fmt.Errorf("JSONデコード失敗: %w", err)
		}

		tweetsModel := createTweets(&tw.Tweet)
		users := createUsers(&tw.Tweet, tweetsModel)
		media := createMedia(&tw.Tweet, tweetsModel)
		urls := createUrls(&tw.Tweet)
		hashtags := createHashtags(&tw.Tweet)

		batch = append(batch, &insertData{
			tweet:    tweetsModel,
			users:    users,
			media:    media,
			urls:     urls,
			hashtags: hashtags,
		})

		if len(batch) >= batchSize {
			inserted, err := insertBatch(tx, batch)
			if err != nil {
				return 0, err
			}
			totalCount += inserted
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		inserted, err := insertBatch(tx, batch)
		if err != nil {
			return 0, err
		}
		totalCount += inserted
	}

	// 配列の閉じカッコ ']' を消費
	_, _ = dec.Token()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return totalCount, nil
}

func importTweetsFromFile(db *sqlx.DB, path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return importTweetsFromReader(db, file)
}

func isTweetJSFile(filename string) bool {
	base := filepath.Base(filename)
	return base == "tweets.js" || (strings.HasPrefix(base, "tweets-part") && strings.HasSuffix(base, ".js"))
}

// importTweetsFromZip ZIPファイルから解凍せずに直接ツイートデータを読み込んでインポート
func importTweetsFromZip(db *sqlx.DB, zipPath string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("ZIPファイルオープン失敗: %w", err)
	}
	defer zr.Close()

	var totalImported int64
	for _, zf := range zr.File {
		if isTweetJSFile(zf.Name) {
			rc, err := zf.Open()
			if err != nil {
				return fmt.Errorf("ZIP内ファイル読込失敗 (%s): %w", zf.Name, err)
			}
			rows, err := importTweetsFromReader(db, rc)
			_ = rc.Close()
			if err != nil {
				return fmt.Errorf("インポート失敗 (%s): %w", zf.Name, err)
			}
			fmt.Printf("インポート完了: %s（%d件）\n", zf.Name, rows)
			totalImported += rows
		}
	}

	if totalImported == 0 {
		fmt.Println("対象のツイートデータ (tweets.js / tweets-part*.js) がZIP内に見つかりませんでした。")
	}

	return finishImport(db)
}

// importTweetsFromDir ディレクトリ配下のjson/jsファイルをインポート (互換性用)
func importTweetsFromDir(db *sqlx.DB, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		ext := filepath.Ext(name)
		if !entry.IsDir() && (ext == ".json" || ext == ".js") {
			if name == "tweet-headers.json" || name == "tweet-headers.js" || name == "like.json" || name == "like.js" {
				continue
			}
			fullPath := filepath.Join(dirPath, name)
			rows, err := importTweetsFromFile(db, fullPath)
			if err != nil {
				return err
			}
			fmt.Printf("インポート完了: %s（%d件）\n", fullPath, rows)
		}
	}

	return finishImport(db)
}

func finishImport(db *sqlx.DB) error {
	// 自分のIDを追加
	if _, err := db.Exec("INSERT OR IGNORE INTO users VALUES (?, ?, 0)", config.MyUserID, config.MyName); err != nil {
		return err
	}

	if updateTwilog {
		if err := updateTwilogDate(db); err != nil {
			return err
		}
	}
	fmt.Println("インポート完了！")
	return nil
}

func updateTwilogDate(db *sqlx.DB) error {
	csvPath := "./data/csv/nayuneko-250707.csv"

	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare("UPDATE tweets SET created_at = ?, created_date = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("CSVの読み込みエラー: %w", err)
		}
		if len(record) < 5 {
			continue
		}

		idStr := record[0]
		dateStr := record[2]
		logType := record[4]

		if logType != "1" {
			continue
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}

		createdAt, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local)
		if err != nil {
			return fmt.Errorf("スキップ: 行 %s（日時パース失敗）: %w", idStr, err)
		}

		createdDate := createdAt.Format("20060102")

		if _, err := stmt.Exec(
			createdAt.In(time.UTC).Format(time.RFC3339),
			createdDate,
			id,
		); err != nil {
			return fmt.Errorf("UPDATE失敗: id=%d: %w", id, err)
		}
	}

	return tx.Commit()
}

func main() {
	db, err := sqlx.Open("sqlite3", config.DBFile)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	if err := setupDB(db); err != nil {
		log.Printf("DB PRAGMA設定警告: %v", err)
	}

	var targetPath string
	if len(os.Args) >= 2 {
		targetPath = os.Args[1]
	}

	if targetPath != "" {
		stat, err := os.Stat(targetPath)
		if err != nil {
			log.Fatalf("指定されたパスが存在しないかエラー: %v", err)
		}
		if !stat.IsDir() && strings.HasSuffix(strings.ToLower(targetPath), ".zip") {
			if err := importTweetsFromZip(db, targetPath); err != nil {
				log.Fatal(err)
			}
			return
		} else if stat.IsDir() {
			if err := importTweetsFromDir(db, targetPath); err != nil {
				log.Fatal(err)
			}
			return
		} else {
			// 単一ファイルの場合
			rows, err := importTweetsFromFile(db, targetPath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("インポート完了: %s（%d件）\n", targetPath, rows)
			if err := finishImport(db); err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	// 引数なしの場合はデフォルトで JsonDir または TweetsDir からインポート
	if _, err := os.Stat(config.JsonDir); err == nil {
		if err := importTweetsFromDir(db, config.JsonDir); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Fatal("使い方: import-x-archive <x-archive.zip | データディレクトリ>")
}
