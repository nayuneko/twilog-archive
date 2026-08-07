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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"twilog-archive/internal/constant"
	"twilog-archive/internal/model"
	"twilog-archive/internal/utils"
	"twilog-archive/internal/xdata"
)

const (
	updateTwilog = false
)

var (
	reTweetsPart = regexp.MustCompile(`^tweets-part(\d+)\.js$`)
)

type stmtMap map[string]*sql.Stmt
type insertData struct {
	tweet    *model.Tweets
	users    []model.Users
	media    []model.Media
	urls     []model.URLs
	hashtags []model.Hashtags
}

func insertAll(sm stmtMap, d *insertData) (int64, error) {
	// tweets
	r, err := sm["tweets"].Exec(
		d.tweet.ID,
		d.tweet.CreatedAt.Format(time.RFC3339),
		d.tweet.CreatedAt.In(time.Local).Format("20060102"),
		d.tweet.ScreenName,
		d.tweet.FullText,
		d.tweet.Retweeted,
		d.tweet.Replied,
		constant.LogTypeXArchive,
		d.tweet.UserID,
		d.tweet.EmbedMediaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("tweetsの追加に失敗: id = %d, %w", d.tweet.ID, err)
	}
	rows, _ := r.RowsAffected()
	//users
	for _, u := range d.users {
		if _, err := sm["users"].Exec(
			u.ID,
			u.Name,
			d.tweet.ID,
		); err != nil {
			return 0, fmt.Errorf("usersの追加に失敗: id = %d, uid = %d, %w", d.tweet.ID, u.ID, err)
		}
	}
	// media
	for _, m := range d.media {
		if _, err := sm["media"].Exec(
			d.tweet.ID,
			m.Index,
			m.MediaURL,
			m.MediaType,
		); err != nil {
			return 0, fmt.Errorf("mediaの追加に失敗: id = %d, idx = %d, %w", d.tweet.ID, m.Index, err)
		}
	}
	// urls
	for _, u := range d.urls {
		if _, err := sm["urls"].Exec(
			d.tweet.ID,
			u.Index,
			u.URL,
			u.ExpandURL,
			u.DisplayURL,
		); err != nil {
			return 0, fmt.Errorf("urlの追加に失敗: id = %d, idx = %d, %w", d.tweet.ID, u.Index, err)
		}
	}
	// hashtags
	for _, h := range d.hashtags {
		if _, err := sm["hashtags"].Exec(
			d.tweet.ID,
			h.Index,
			h.Tag,
		); err != nil {
			return 0, fmt.Errorf("hashtagの追加に失敗: id = %d, idx = %d, %w", d.tweet.ID, h.Index, err)
		}
	}
	return rows, nil
}

func createTweets(t *xdata.Tweet) *model.Tweets {

	screenName := constant.MyScreenName
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
	retweeted := screenName != constant.MyScreenName

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
	if t.Entities.UserMentions == nil || len(t.Entities.UserMentions) == 0 {
		return nil
	}
	var r []model.Users
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
		if t.ExtendedEntities.Media != nil && len(t.ExtendedEntities.Media) > 0 {
			return t.ExtendedEntities.Media
		}
		if t.Entities.Media != nil && len(t.Entities.Media) > 0 {
			return t.Entities.Media
		}
		return nil
	}()
	if media == nil {
		return nil
	}
	var r []model.Media
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
	if t.Entities.URLs == nil || len(t.Entities.URLs) == 0 {
		return nil
	}
	var r []model.URLs
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
	if t.Entities.Hashtags == nil || len(t.Entities.Hashtags) == 0 {
		return nil
	}
	var r []model.Hashtags
	for idx, h := range t.Entities.Hashtags {
		r = append(r, model.Hashtags{
			TweetID: int64(t.ID),
			Index:   idx + 1,
			Tag:     h.Text,
		})
	}
	return r
}

func (sm stmtMap) Close() {
	for _, stmt := range sm {
		_ = stmt.Close()
	}
}

// importTweetsFromReader io.Reader (JSまたはJSONデータ) からツイートをデコードしてインポート
func importTweetsFromReader(db *sqlx.DB, r io.Reader) (int64, error) {
	jsReader := xdata.NewStripJSPrefixReader(r)

	var tweets []xdata.TweetWrapper
	if err := json.NewDecoder(jsReader).Decode(&tweets); err != nil {
		return 0, err
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	sm := make(stmtMap)
	for _, s := range []struct {
		name string
		q    string
	}{
		{name: "tweets", q: "INSERT OR IGNORE INTO tweets VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"},
		{name: "users", q: `INSERT OR IGNORE INTO users (id, name, last_status_id) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET name = excluded.name, last_status_id = excluded.last_status_id WHERE excluded.last_status_id > users.last_status_id`},
		{name: "media", q: "INSERT OR IGNORE INTO media VALUES (?, ?, ?, ?)"},
		{name: "urls", q: "INSERT OR IGNORE INTO urls VALUES (?, ?, ?, ?, ?)"},
		{name: "hashtags", q: "INSERT OR IGNORE INTO hashtags VALUES (?, ?, ?)"},
	} {
		stmt, err := tx.Prepare(s.q)
		if err != nil {
			return 0, fmt.Errorf("%sステートメントの作成に失敗: %w", s.name, err)
		}
		sm[s.name] = stmt
	}
	defer sm.Close()

	var count int64
	for _, tw := range tweets {
		tweetsModel := createTweets(&tw.Tweet)
		users := createUsers(&tw.Tweet, tweetsModel)
		media := createMedia(&tw.Tweet, tweetsModel)
		insertedRows, err := insertAll(sm, &insertData{
			tweetsModel,
			users,
			media,
			createUrls(&tw.Tweet),
			createHashtags(&tw.Tweet),
		})
		if err != nil {
			return 0, err
		}
		count += insertedRows
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
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
	return base == "tweets.js" || reTweetsPart.MatchString(base)
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
	if _, err := db.Exec("INSERT OR IGNORE INTO users VALUES (?, ?, 0)", constant.MyUserID, constant.MyName); err != nil {
		return err
	}

	// twilogのカレンダーデータ生成
	if _, err := utils.MakeCalendarData(db); err != nil {
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
	db, err := sqlx.Open("sqlite3", constant.DBFile)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

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
	if _, err := os.Stat(constant.JsonDir); err == nil {
		if err := importTweetsFromDir(db, constant.JsonDir); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Fatal("使い方: import-x-archive <x-archive.zip | データディレクトリ>")
}
