JS_FILES := $(wildcard data/tweets/*.js)
JSON_FILES := $(patsubst data/tweets/%.js,data/json/%.json,$(JS_FILES))

.PHONY: pre-parse clean

pre-parse: $(JSON_FILES)

data/json/%.json: data/tweets/%.js
	node tools/tweets_parser.js $< $@

clean:
	rm -f data/json/*.json
