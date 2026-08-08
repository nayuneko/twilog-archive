package repository_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"twilog-archive/internal/repository"
)

func TestGetCalendarData(t *testing.T) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

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
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// テストデータの挿入
	_, err = db.Exec(`
		INSERT INTO tweets (id, created_at, created_date, screen_name, full_text, retweeted, replied, log_type) VALUES
		(1, '2023-10-25T12:00:00Z', '20231025', 'user', 'tweet 1', false, false, 1),
		(2, '2023-10-25T13:00:00Z', '20231025', 'user', 'tweet 2', false, false, 1),
		(3, '2023-10-24T10:00:00Z', '20231024', 'user', 'tweet 3', false, false, 1),
		(4, '2023-09-01T09:00:00Z', '20230901', 'user', 'tweet 4', false, false, 1),
		(5, '2022-12-31T23:59:59Z', '20221231', 'user', 'tweet 5', false, false, 1);
	`)
	if err != nil {
		t.Fatalf("failed to insert test tweets: %v", err)
	}

	cal, err := repository.GetCalendarData(db)
	if err != nil {
		t.Fatalf("GetCalendarData error: %v", err)
	}

	// 2023年
	if _, ok := cal["2023"]; !ok {
		t.Errorf("expected year 2023 in calendar data")
	}

	// 2023年10月
	october, ok := cal["2023"]["10"]
	if !ok {
		t.Errorf("expected month 10 in year 2023")
	} else {
		if len(october) != 2 || october[0] != 25 || october[1] != 24 {
			t.Errorf("unexpected days in 2023/10: %v", october)
		}
	}

	// 2023年9月
	september, ok := cal["2023"]["9"]
	if !ok {
		t.Errorf("expected month 9 in year 2023")
	} else {
		if len(september) != 1 || september[0] != 1 {
			t.Errorf("unexpected days in 2023/9: %v", september)
		}
	}

	// 2022年12月
	december, ok := cal["2022"]["12"]
	if !ok {
		t.Errorf("expected month 12 in year 2022")
	} else {
		if len(december) != 1 || december[0] != 31 {
			t.Errorf("unexpected days in 2022/12: %v", december)
		}
	}
}
