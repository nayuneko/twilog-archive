package main

import (
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const testSchema = `
CREATE TABLE tweets (
    id INTEGER PRIMARY KEY,
    created_at DATETIME NOT NULL,
    created_date TEXT NOT NULL,
    screen_name TEXT NOT NULL,
    full_text TEXT NOT NULL,
    retweeted BOOLEAN NOT NULL default false,
    replied BOOLEAN NOT NULL default false,
    log_type INTEGER NOT NULL,
    user_id INTEGER,
    embed_media_url TEXT
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    screen_name TEXT,
    last_status_id INTEGER NOT NULL
);

CREATE TABLE media (
    tweet_id INTEGER NOT NULL,
    media_index INTEGER NOT NULL,
    media_url TEXT NOT NULL,
    type TEXT,
    PRIMARY KEY (tweet_id, media_index)
);

CREATE TABLE urls (
    tweet_id INTEGER NOT NULL,
    url_index INTEGER NOT NULL,
    url TEXT NOT NULL,
    expanded_url TEXT,
    display_url TEXT,
    PRIMARY KEY (tweet_id, url_index)
);

CREATE TABLE hashtags (
    tweet_id INTEGER NOT NULL,
    tag_index INTEGER NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (tweet_id, tag_index)
);
`

func TestIsTweetJSFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"tweets.js", true},
		{"tweets-part1.js", true},
		{"tweets-part99.js", true},
		{"like.js", false},
		{"tweet-headers.json", false},
		{"other.txt", false},
	}

	for _, tt := range tests {
		if got := isTweetJSFile(tt.filename); got != tt.expected {
			t.Errorf("isTweetJSFile(%q) = %v, want %v", tt.filename, got, tt.expected)
		}
	}
}

func TestImportTweetsFromReader(t *testing.T) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("DBオープン失敗: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("スキーマ作成失敗: %v", err)
	}

	sampleJS := `window.YTD.tweets.part0 = [
  {
    "tweet": {
      "id": "1000000000000000001",
      "created_at": "Sat Aug 02 05:49:11 +0000 2025",
      "full_text": "Hello world #test https://t.co/xyz123",
      "entities": {
        "hashtags": [{"text": "test", "indices": ["12", "17"]}],
        "user_mentions": [{"id": "9999", "screen_name": "user9999", "name": "User 9999", "indices": ["0", "5"]}],
        "urls": [{"url": "https://t.co/xyz123", "expanded_url": "https://example.com", "display_url": "example.com", "indices": ["18", "38"]}],
        "media": []
      }
    }
  },
  {
    "tweet": {
      "id": "1000000000000000002",
      "created_at": "Sat Aug 02 06:00:00 +0000 2025",
      "full_text": "Second tweet with photo https://t.co/pic123",
      "extended_entities": {
        "media": [{"id": "2001", "media_url_https": "https://pbs.twimg.com/media/abc.jpg", "type": "photo", "url": "https://t.co/pic123"}]
      }
    }
  }
];`

	count, err := importTweetsFromReader(db, strings.NewReader(sampleJS))
	if err != nil {
		t.Fatalf("importTweetsFromReader 失敗: %v", err)
	}

	if count != 2 {
		t.Errorf("インポート件数 = %d, want 2", count)
	}

	var tweetCount int
	if err := db.Get(&tweetCount, "SELECT COUNT(*) FROM tweets"); err != nil {
		t.Fatalf("tweetsカウント取得失敗: %v", err)
	}
	if tweetCount != 2 {
		t.Errorf("DBのtweets件数 = %d, want 2", tweetCount)
	}

	var hashtagCount int
	if err := db.Get(&hashtagCount, "SELECT COUNT(*) FROM hashtags"); err != nil {
		t.Fatalf("hashtagsカウント取得失敗: %v", err)
	}
	if hashtagCount != 1 {
		t.Errorf("DBのhashtags件数 = %d, want 1", hashtagCount)
	}

	var mediaCount int
	if err := db.Get(&mediaCount, "SELECT COUNT(*) FROM media"); err != nil {
		t.Fatalf("mediaカウント取得失敗: %v", err)
	}
	if mediaCount != 1 {
		t.Errorf("DBのmedia件数 = %d, want 1", mediaCount)
	}

	var userCount int
	if err := db.Get(&userCount, "SELECT COUNT(*) FROM users"); err != nil {
		t.Fatalf("usersカウント取得失敗: %v", err)
	}
	if userCount != 1 {
		t.Errorf("DBのusers件数 = %d, want 1", userCount)
	}
}


