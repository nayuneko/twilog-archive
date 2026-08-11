.PHONY: import import-twilog run clean build-web dev

GO_BUILD := CGO_ENABLED=1 CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go build -tags fts5

build-web:
	cd web && npm run build

run: bin/server
	bin/server

dev: bin/server
	(cd web && npm run dev) & bin/server

bin/server: cmd/server/main.go build-web
	$(GO_BUILD) -o bin/server cmd/server/main.go

bin/extract-archive: cmd/extract-archive/main.go
	$(GO_BUILD) -o $@ $^

bin/import-x-archive: cmd/import-x-archive/main.go
	$(GO_BUILD) -o $@ $^

bin/import-twilog: cmd/import-twilog/main.go
	$(GO_BUILD) -o $@ $^

bin/mcp-server: cmd/mcp-server/main.go
	$(GO_BUILD) -o $@ $^

bin/fix-unescape: cmd/fix-unescape/main.go
	$(GO_BUILD) -o $@ $^

chat-gemini: bin/mcp-server
	python3 tools/chat_gemini.py

bin/index.cgi: cmd/cgi-server/main.go build-web
	CC="zig cc -target x86_64-linux-musl -fno-sanitize=undefined" GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go build -tags fts5 -o bin/index.cgi cmd/cgi-server/main.go

build-cgi: bin/index.cgi

clean:
	rm -f bin/*

extract-archive: bin/extract-archive
	./bin/extract-archive $(ZIP)

import: bin/import-x-archive
	./bin/import-x-archive $(ZIP)
	$(MAKE) optimize-db

import-twilog: bin/import-twilog
	./bin/import-twilog $(CSV)
	$(MAKE) optimize-db

optimize-db:
	@echo "===> SQLite DB を最適化中 (WAL統合 & VACUUM & FTS5最適化)..."
	@sqlite3 data/db/tweets.db "PRAGMA wal_checkpoint(TRUNCATE); VACUUM; INSERT INTO tweets_fts(tweets_fts) VALUES('optimize');"
	@rm -f data/db/tweets.db-wal data/db/tweets.db-shm
	@echo "===> DB 最適化が完了しました！"

# --- XREA デプロイ自動化 ---
XREA_USER ?= atkg3a
XREA_HOST ?= s345.xrea.com
XREA_PATH ?= /virtual/atkg3a/public_html/x.nayuneko.jp
XREA_DB_PATH ?= /virtual/atkg3a/db

deploy-xrea: build-cgi
	@echo "===> XREA ($(XREA_USER)@$(XREA_HOST):$(XREA_PATH)) へ Webアセット・CGIバイナリ・.htaccess を転送中..."
	ssh $(XREA_USER)@$(XREA_HOST) "mkdir -p $(XREA_PATH)"
	scp -r web/dist/* $(XREA_USER)@$(XREA_HOST):$(XREA_PATH)/
	scp bin/index.cgi $(XREA_USER)@$(XREA_HOST):$(XREA_PATH)/index.cgi
	ssh $(XREA_USER)@$(XREA_HOST) "chmod 755 $(XREA_PATH)/index.cgi"
	scp opts/xrea/.htaccess $(XREA_USER)@$(XREA_HOST):$(XREA_PATH)/.htaccess
	ssh $(XREA_USER)@$(XREA_HOST) "htpasswd -b -c $(XREA_PATH)/.htpasswd nayuneko nayu1279"
	@echo "===> アプリケーションのデプロイが完了しました！"

deploy-db-xrea: optimize-db
	@echo "===> XREA ($(XREA_USER)@$(XREA_HOST):$(XREA_DB_PATH)) [Web非公開エリア] へ tweets.db を安全転送中..."
	ssh $(XREA_USER)@$(XREA_HOST) "mkdir -p $(XREA_DB_PATH)"
	scp data/db/tweets.db $(XREA_USER)@$(XREA_HOST):$(XREA_DB_PATH)/tweets.db.tmp
	ssh $(XREA_USER)@$(XREA_HOST) "mv $(XREA_DB_PATH)/tweets.db.tmp $(XREA_DB_PATH)/tweets.db"
	@echo "===> データベースのアップデートが完了しました！"


