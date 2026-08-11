package server

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"twilog-archive/internal/handler"
	"twilog-archive/web"
)

// NewEngine に DB 接続を渡し、ルーティングや Basic 認証が組み込まれた Echo インスタンスを返します。
func NewEngine(db *sqlx.DB) *echo.Echo {
	e := echo.New()

	// Logger の出力先を Stderr に設定 (CGI の stdout 破壊を防止)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: os.Stderr,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 環境変数 BASIC_AUTH_USER と BASIC_AUTH_PASS が設定されている場合、Basic認証を適用
	authUser := os.Getenv("BASIC_AUTH_USER")
	if authUser == "" {
		authUser = os.Getenv("REDIRECT_BASIC_AUTH_USER")
	}
	authPass := os.Getenv("BASIC_AUTH_PASS")
	if authPass == "" {
		authPass = os.Getenv("REDIRECT_BASIC_AUTH_PASS")
	}
	if authUser != "" && authPass != "" {
		e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if username == authUser && password == authPass {
				return true, nil
			}
			return false, nil
		}))
	}

	// API ルーティング
	m := e.Group("/api/tweets")
	m.GET("/latest", handler.TweetsLatest(db))
	m.GET("/dates/:date", handler.TweetsDates(db))
	m.GET("/search/", handler.TweetsSearch(db))

	e.GET("/api/calendar", handler.Calendar(db))

	// フロントエンド静的ファイル配信 & SPAフォールバック
	distFS, err := fs.Sub(web.Dist, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distFS))
		e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}

			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				fileServer.ServeHTTP(w, r)
				return
			}

			f, err := distFS.Open(path)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})))
	}

	return e
}
