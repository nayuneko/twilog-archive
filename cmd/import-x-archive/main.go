package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"twilog-archive/internal/constant"
	"twilog-archive/internal/model"
	"twilog-archive/internal/utils"
	"twilog-archive/internal/xdata"
)

const (
	updateTwilog = false
)

type stmtMap map[string]*sql.Stmt
type insertData struct {
	tweet    *model.Tweets
	users    []model.Users
	media    []model.Media
	urls     []model.URLs
	hashtags []model.Hashtags
}

func createStatement(db *sqlx.DB) (stmtMap, error) {
	r := map[string]*sql.Stmt{}
	for _, s := range []struct {
		name string
		q    string
	}{
		{
			name: "tweets",
			q:    "INSERT OR IGNORE INTO tweets VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		},
		{
			name: "users",
			q: `INSERT OR IGNORE INTO users (id, name, last_status_id)
VALUES (?, ?, ?)
ON CONFLICT(id)
DO UPDATE SET name = excluded.name, last_status_id = excluded.last_status_id
   WHERE excluded.last_status_id > users.last_status_id`,
		},
		{
			name: "media",
			q:    "INSERT OR IGNORE INTO media VALUES (?, ?, ?, ?)",
		},
		{
			name: "urls",
			q:    "INSERT OR IGNORE INTO urls VALUES (?, ?, ?, ?, ?)",
		},
		{
			name: "hashtags",
			q:    "INSERT OR IGNORE INTO hashtags VALUES (?, ?, ?)",
		},
	} {
		stmt, err := db.Prepare(s.q)
		if err != nil {
			return nil, fmt.Errorf("%sステートメントの作成に失敗: %w", s.name, err)
		}
		r[s.name] = stmt
	}
	return r, nil
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

func importTweetsFromFile(db *sqlx.DB, path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var tweets []xdata.TweetWrapper
	if err := json.NewDecoder(file).Decode(&tweets); err != nil {
		return 0, err
	}

	sm, err := createStatement(db)
	if err != nil {
		return 0, err
	}

	var rows int64
	for _, tw := range tweets {
		tweets := createTweets(&tw.Tweet)
		users := createUsers(&tw.Tweet, tweets)
		media := createMedia(&tw.Tweet, tweets)
		if rows, err = insertAll(sm, &insertData{
			tweets,
			users,
			media,
			createUrls(&tw.Tweet),
			createHashtags(&tw.Tweet),
		}); err != nil {
			return 0, err
		}
		//_, _ = stmtFTS.Exec(t.IDStr, t.FullText)
	}
	return rows, nil
}

// importTweets jsonディレクトリにあるtweets.jsonをすべてインポート
func importTweets() error {

	db, err := sqlx.Open("sqlite3", constant.DBFile)
	if err != nil {
		return err
	}
	defer db.Close()

	entries, err := os.ReadDir(constant.JsonDir)
	if err != nil {
		return err
	}

	//var header []xdata.TweetHeaderWrapper
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && filepath.Ext(name) == ".json" {
			fullPath := filepath.Join(constant.JsonDir, name)
			if name == "tweet-headers.json" {
				/*
					if d, err := loadTweetHeaderFromFile(fullPath); err != nil {
						return err
					} else {
						header = d
					}
				*/
				continue
			} else if name == "like.json" {
				continue
			} else {
				if rows, err := importTweetsFromFile(db, fullPath); err != nil {
					return err
				} else {
					fmt.Printf("インポート完了: %s（%d件）\n", fullPath, rows)
				}
			}
		}
	}
	// 自分のIDを追加
	if _, err := db.Exec("INSERT OR IGNORE INTO users VALUES (?, ?, 0)", constant.MyUserID, constant.MyName); err != nil {
		return err
	}
	//

	// twilogの時刻で更新
	if _, err := utils.MakeCalendarData(db); err != nil {
		return err
	}
	if updateTwilog {
		// twilogの時刻で更新
		if err := updateTwilogDate(db); err != nil {
			return err
		}
	}
	fmt.Println("インポート完了！")
	return nil
}

func updateTwilogDate(db *sqlx.DB) error {
	// ファイルパス・DBパスの設定
	csvPath := "./data/csv/nayuneko-250707.csv"

	// CSVオープン
	f, err := os.Open(csvPath)
	if err != nil {
		return err
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
		dateStr := record[2]
		logType := record[4]

		// ログタイプは1:ツイート(RT含む)、2:いいね、3:ブックマーク
		if logType != "1" {
			continue
		}

		id, _ := strconv.ParseInt(idStr, 10, 64)

		// 投稿日時のパース（twilogの日時はJST）
		createdAt, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local)
		if err != nil {
			return fmt.Errorf("スキップ: 行 %s（日時パース失敗）: %w", idStr, err)
		}

		createdDate := createdAt.Format("20060102")

		q := "UPDATE tweets SET created_at = ?, created_date = ? WHERE id = ?"
		_, err = db.Exec(
			q,
			createdAt.In(time.UTC).Format(time.RFC3339),
			createdDate,
			id,
		)
		if err != nil {
			return fmt.Errorf("スキップ: 行 %s（INSERT失敗）: %w", idStr, err)
		}
	}
	return nil
}

func main() {
	if err := importTweets(); err != nil {
		log.Fatal(err)
	}
}
