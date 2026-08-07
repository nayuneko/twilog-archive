JS_FILES := $(wildcard data/tweets/*.js)
JSON_FILES := $(patsubst data/tweets/%.js,data/json/%.json,$(JS_FILES))

.PHONY: import run clean build-web dev pre-parse clean-json

GO_BUILD := CGO_ENABLED=1 CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go build -tags fts5

build-web:
	cd web && npm run build

pre-parse: $(JSON_FILES)

run: bin/server
	bin/server

dev: bin/server
	(cd web && npm run dev) & bin/server

data/json/%.json: data/tweets/%.js
	node tools/tweets_parser.js $< $@

bin/server: cmd/server/main.go build-web
	$(GO_BUILD) -o bin/server cmd/server/main.go

bin/extract-archive: cmd/extract-archive/main.go
	go build -o $@ $^

bin/import-x-archive: cmd/import-x-archive/main.go
	go build -o $@ $^

bin/mcp-server: cmd/mcp-server/main.go
	go build -o $@ $^

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

import-x-archive: bin/import-x-archive
	./bin/import-x-archive

import: extract-archive pre-parse import-x-archive

clean-import: clean-data import
