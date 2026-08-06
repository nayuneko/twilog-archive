JS_FILES := $(wildcard data/tweets/*.js)
JSON_FILES := $(patsubst data/tweets/%.js,data/json/%.json,$(JS_FILES))

.PHONY: pre-parse clean-json

pre-parse: $(JSON_FILES)

run: bin/api
	bin/api

data/json/%.json: data/tweets/%.js
	node tools/tweets_parser.js $< $@

bin/api:
	go build -o bin/api cmd/api/main.go

bin/extract-archive: cmd/extract-archive/main.go
	go build -o $@ $^

bin/import-x-archive: cmd/import-x-archive/main.go
	go build -o $@ $^

bin/mcp-server: cmd/mcp-server/main.go
	go build -o $@ $^

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
