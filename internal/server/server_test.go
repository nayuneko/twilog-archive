package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
)

func TestBasicAuthMiddleware(t *testing.T) {
	// メモリ内SQLite
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	os.Setenv("BASIC_AUTH_USER", "admin")
	os.Setenv("BASIC_AUTH_PASS", "secret")
	defer func() {
		os.Unsetenv("BASIC_AUTH_USER")
		os.Unsetenv("BASIC_AUTH_PASS")
	}()

	e := NewEngine(db)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// 認証なしのリクエスト ➔ 401 Unauthorized
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	// 正しい認証ヘッダー付きリクエスト ➔ 200 OK
	reqAuth := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqAuth.SetBasicAuth("admin", "secret")
	recAuth := httptest.NewRecorder()
	e.ServeHTTP(recAuth, reqAuth)

	if recAuth.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recAuth.Code)
	}
}
