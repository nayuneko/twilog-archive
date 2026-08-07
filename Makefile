.PHONY: import run clean build-web dev clean-js clean-json clean-data

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

bin/mcp-server: cmd/mcp-server/main.go
	$(GO_BUILD) -o $@ $^

chat-gemini: bin/mcp-server
	python3 tools/chat_gemini.py

clean:
	rm -f bin/*

clean-js:
	rm -f data/tweets/*.js

clean-json:
	rm -f data/json/*.json

clean-data: clean-js clean-json

extract-archive: bin/extract-archive
	./bin/extract-archive $(ZIP)

import: bin/import-x-archive
	./bin/import-x-archive $(ZIP)
