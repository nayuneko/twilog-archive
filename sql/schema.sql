-- sqlite3 data/db/tweets.db < sql/schema.sql

-- ツイート
CREATE TABLE tweets (
    id INTEGER PRIMARY KEY,
    created_at DATETIME NOT NULL,
    created_date TEXT NOT NULL,
    screen_name TEXT NOT NULL,
    full_text TEXT NOT NULL,
    retweeted BOOLEAN NOT NULL default false,
    replied BOOLEAN NOT NULL default false,
    log_type INTEGER NOT NULL, -- 1: twilog / 2: Xアーカイブ
    user_id INTEGER,
    embed_media_url TEXT -- x.com/xxx/status/yyy/photo/1 形式のURL (twilogのURL)、t.co形式のURL (XアーカイブのURL)
);

-- ユーザ情報
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    last_status_id INTEGER NOT NULL
);

-- メディア
CREATE TABLE media (
    tweet_id INTEGER NOT NULL,
    media_index INTEGER NOT NULL,
    media_url TEXT NOT NULL, -- 画像の直URL
    type TEXT, -- photo, video, animated_gif
    PRIMARY KEY (tweet_id, media_index),
    FOREIGN KEY (tweet_id) REFERENCES tweets(tweet_id) ON DELETE CASCADE
);

-- URL
CREATE TABLE urls (
    tweet_id INTEGER NOT NULL,
    url_index INTEGER NOT NULL,
    url TEXT NOT NULL,             -- 短縮URL (t.co)
    expanded_url TEXT,            -- 展開後URL
    display_url TEXT,             -- 表示用（例: example.com/...）
    PRIMARY KEY (tweet_id, url_index),
    FOREIGN KEY (tweet_id) REFERENCES tweets(tweet_id) ON DELETE CASCADE
);

-- ハッシュタグ
CREATE TABLE hashtags (
    tweet_id INTEGER NOT NULL,
    tag_index INTEGER NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (tweet_id, tag_index),
    FOREIGN KEY (tweet_id) REFERENCES tweets(tweet_id) ON DELETE CASCADE
);

CREATE INDEX idx_created_date_id ON tweets (created_date, id);

-- FTS5全文検索用テーブル
CREATE VIRTUAL TABLE tweets_fts USING fts5(
  id UNINDEXED,
  full_text,
  content='',
  tokenize='unicode61'
);