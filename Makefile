JS_FILES := $(wildcard data/tweets/*.js)
JSON_FILES := $(patsubst data/tweets/%.js,data/json/%.json,$(JS_FILES))

.PHONY: pre-parse clean-json

pre-parse: $(JSON_FILES)

data/json/%.json: data/tweets/%.js
	node tools/tweets_parser.js $< $@

bin/api:
	go build -o bin/api cmd/api/main.go

bin/extract-archive: cmd/extract-archive/main.go
	go build -o $@ $^

bin/import-x-archive: cmd/import-x-archive/main.go
	go build -o $@ $^

clean:
	rm -f bin/*

clean-json:
	rm -f data/json/*.json

import: clean-json bin/extract-archive pre-parse bin/import-x-archive
	./bin/extract-archive
	./bin/import-x-archive