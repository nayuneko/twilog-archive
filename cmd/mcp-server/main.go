package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	_ "github.com/mattn/go-sqlite3"

	"twilog-archive/internal/config"
	"twilog-archive/internal/form"
	"twilog-archive/internal/repository"
)

func resolveDBPath() string {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		return dbPath
	}
	execPath, err := os.Executable()
	if err == nil {
		projectDir := filepath.Dir(filepath.Dir(execPath))
		candidate := filepath.Join(projectDir, config.DBFile)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return config.DBFile
}

func main() {
	dbPath := resolveDBPath()
	db, err := sqlx.Open("sqlite3", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// MCP Server の作成
	s := server.NewMCPServer(
		"twilog-archive",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Tool 1: search_tweets
	searchTool := mcp.NewTool("search_tweets",
		mcp.WithDescription("Twilog/Xアーカイブの過去ツイートをキーワードで全文検索します (最新50件まで取得)"),
		mcp.WithString("query", mcp.Required(), mcp.Description("検索キーワード")),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		req := &form.SearchRequest{
			SearchWord: query,
		}
		tweets, err := repository.Search(db, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("DB検索エラー: %v", err)), nil
		}

		if len(tweets) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("キーワード '%s' に一致するツイートは見つかりませんでした。", query)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("検索結果: '%s' (%d件)\n\n", query, len(tweets)))
		for _, t := range tweets {
			name := config.MyName
			if t.Name != nil {
				name = *t.Name
			}
			sb.WriteString(fmt.Sprintf("[%s] %s (@%s) (ID: %d)\n%s\n---\n",
				t.CreatedAt.Format("2006-01-02 15:04:05"), name, t.ScreenName, t.ID, t.FullText))
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	// Tool 2: get_tweets_by_date
	dateTool := mcp.NewTool("get_tweets_by_date",
		mcp.WithDescription("指定した日付 (YYYYMMDD または YYYY-MM-DD) の過去ツイート一覧を取得します"),
		mcp.WithString("date", mcp.Required(), mcp.Description("日付 (例: 20230801 または 2023-08-01)")),
	)
	s.AddTool(dateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dateStr, err := request.RequireString("date")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		cleanDate := strings.ReplaceAll(dateStr, "-", "")
		tweets, err := repository.FindByDates(db, cleanDate)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("DB取得エラー: %v", err)), nil
		}

		if len(tweets) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("日付 '%s' のツイートは見つかりませんでした。", dateStr)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s のツイート一覧 (%d件):\n\n", dateStr, len(tweets)))
		for _, t := range tweets {
			name := config.MyName
			if t.Name != nil {
				name = *t.Name
			}
			sb.WriteString(fmt.Sprintf("[%s] %s (@%s) (ID: %d)\n%s\n---\n",
				t.CreatedAt.Format("15:04:05"), name, t.ScreenName, t.ID, t.FullText))
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	// Tool 3: get_latest_tweets
	latestTool := mcp.NewTool("get_latest_tweets",
		mcp.WithDescription("最新の過去ツイート一覧を取得します"),
	)
	s.AddTool(latestTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tweets, err := repository.Latest(db, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("DB取得エラー: %v", err)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("最新のツイート (%d件):\n\n", len(tweets)))
		for _, t := range tweets {
			name := config.MyName
			if t.Name != nil {
				name = *t.Name
			}
			sb.WriteString(fmt.Sprintf("[%s] %s (@%s) (ID: %d)\n%s\n---\n",
				t.CreatedAt.Format("2006-01-02 15:04:05"), name, t.ScreenName, t.ID, t.FullText))
		}

		return mcp.NewToolResultText(sb.String()), nil
	})

	// Server の標準入出力 (stdio) 起動
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
