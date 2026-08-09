package main

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"twilog-archive/internal/form"
	"twilog-archive/internal/repository"
)

func TestResolveDBPath(t *testing.T) {
	t.Run("DB_PATH env set", func(t *testing.T) {
		customPath := "/tmp/custom_tweets.db"
		os.Setenv("DB_PATH", customPath)
		defer os.Unsetenv("DB_PATH")

		got := resolveDBPath()
		if got != customPath {
			t.Errorf("resolveDBPath() = %v, want %v", got, customPath)
		}
	})

	t.Run("DB_PATH unset default fallback", func(t *testing.T) {
		os.Unsetenv("DB_PATH")
		got := resolveDBPath()
		if got == "" {
			t.Error("resolveDBPath() returned empty string")
		}
	})
}

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
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
		last_status_id INTEGER NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// ダミーデータの追加
	_, err = db.Exec(`
		INSERT INTO tweets VALUES (1, '2023-08-01 12:00:00', '20230801', 'nayuneko', '美味しいラーメンを食べた', 0, 0, 1, 87211693, NULL);
		INSERT INTO tweets VALUES (2, '2023-08-01 13:00:00', '20230801', 'nayuneko', 'うどんも美味しかった', 0, 0, 1, 87211693, NULL);
		INSERT INTO users VALUES (87211693, 'なゆ', 2);
	`)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	return db
}

func TestRepositorySearchAndFindByDates(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("Search Keyword Hit", func(t *testing.T) {
		req := &form.SearchRequest{SearchWord: "ラーメン"}
		results, totalCount, err := repository.Search(db, req)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if totalCount != 1 {
			t.Errorf("got totalCount %d, want 1", totalCount)
		}
		if len(results) != 1 {
			t.Fatalf("got %d tweets, want 1", len(results))
		}
		if results[0].FullText != "美味しいラーメンを食べた" {
			t.Errorf("got %q, want '美味しいラーメンを食べた'", results[0].FullText)
		}
	})

	defaultFilter := form.TweetTypeFilter{IncludeNormal: true, IncludeReply: true, IncludeRT: true}

	t.Run("FindByDates", func(t *testing.T) {
		results, err := repository.FindByDates(db, "20230801", defaultFilter)
		if err != nil {
			t.Fatalf("FindByDates failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("got %d tweets, want 2", len(results))
		}
	})

	t.Run("Latest", func(t *testing.T) {
		results, err := repository.Latest(db, nil, defaultFilter)
		if err != nil {
			t.Fatalf("Latest failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("got %d tweets, want 2", len(results))
		}
	})
}
