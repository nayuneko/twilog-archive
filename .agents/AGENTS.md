# Project Context & Agent Rules (twilog-archive)

このリポジトリは、Twilog CSV データおよび X (旧Twitter) 公式アーカイブ zip からツイートデータを SQLite DB にインポートし、Web UI や MCP (Model Context Protocol) サーバー経由で検索・閲覧するためのシステムです。

## 技術スタック
- **Backend / CLI**: Go 1.25+ (`labstack/echo`, `jmoiron/sqlx`, `mattn/go-sqlite3`)
- **Database**: SQLite3 (`data/db/tweets.db`) + FTS5 (Trigram Tokenizer で全文検索)
- **Frontend**: React 18 / TypeScript / Vite / Tailwind CSS (`web/`)
- **AI Integration**: MCP Server (`cmd/mcp-server`), Gemini Script (`tools/chat_gemini.py`)

## ディレクトリ構造マップ
- **`cmd/`**: エントリポイント
  - `server/`: APIサーバー (Echo Web サーバー)
  - `import-twilog/`: Twilog CSV インポート CLI
  - `import-x-archive/`: X (Twitter) zip アーカイブ インポート CLI
  - `extract-archive/`: zip 内のメディア等の解凍・配置用 CLI
  - `mcp-server/`: MCP サーバー CLI
  - `fix-unescape/`: DB内データのHTMLアンエスケープ一括修正 CLI
- **`internal/`**: バックエンドコアパッケージ
  - `config/`: 環境設定
  - `handler/`: Echo HTTP API ハンドラー
  - `repository/`: SQLite データアクセス層 (FTS5 検索、CRUD)
  - `model/`: データモデル (Tweet, User, Media, URL, Hashtag)
  - `xdata/`: X アーカイブデータ解析処理
  - `text/`: HTML アンエスケープ等のテキスト処理ユーティリティ
- **`sql/`**: DBスキーマ (`schema.sql`)
- **`web/`**: フロントエンドソース (Vite + React TS)
- **`tools/`**: 補助スクリプト (Gemini チャット連携等)

## 主要ビルド & 開発コマンド (Makefile)
- `make dev`: フロントエンド開発サーバー (Vite) + バックエンドサーバー (Echo) を同時起動
- `make run`: バックエンドサーバー起動 (bin/server)
- `make build-web`: フロントエンドのビルド (`web/dist` 生成)
- `make import ZIP=path/to/archive.zip`: Xアーカイブのインポート
- `make import-twilog CSV=path/to/twilog.csv`: Twilog CSVのインポート

## 実装上の重要注意点
1. **SQLite FTS5**: `CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go build -tags fts5` がビルド時に必要（Makefile参照）。
2. **HTML Entity / アンエスケープ**: ツイート本文は `&lt;` `&gt;` `&amp;` などの文字実体参照が含まれる場合があり、`internal/text` パッケージでアンエスケープ処理を行う。
3. **データソースの種類**: `log_type = 1` (Twilog) / `log_type = 2` (Xアーカイブ)。
