# XREA への Go CGI デプロイガイド

XREA（レンタルサーバー）上に Twilog Archive を Go CGI モードで配置・公開する手順です。

## 1. 概要・構成図

- **Webサーバー / 配信**: XREA Apache (`public_html/`)
- **APIサーバー**: Go CGI バイナリ (`index.cgi`)
- **データベース**: SQLite3 (`tweets.db` - 読み取り専用 `mode=ro`)
- **認証**: Basic 認証 (環境変数 `BASIC_AUTH_USER` / `BASIC_AUTH_PASS` 指定時のみ有効)

---

## 2. 自動デプロイ (Makefile コマンド)

SCP を利用して、ワンコマンドで全アセットやデータベースを自動転送・自動デプロイできます。

### 1) アプリ本体 (Webアセット・CGIバイナリ・.htaccess) のデプロイ
```bash
make deploy-xrea
```
※ 自動的に `npm run build` および Linux 用 CGI バイナリコンパイル (`bin/index.cgi`) が実行され、`atkg3a@s345.xrea.com:/virtual/atkg3a/public_html/x.nayuneko.jp` へ転送・パーミッション設定が行われます。

### 2) データベース (`tweets.db`) の安全な更新・差し替え
```bash
make deploy-db-xrea
```
※ **セキュリティ強化**: DBファイルは Web 非公開領域 (`/virtual/atkg3a/db/tweets.db`) へ安全に転送されます（ブラウザからの直接ダウンロードは不可能です）。

---

## 3. ファイル配置構造 (参考)

```text
/virtual/atkg3a/
├── db/
│   └── tweets.db              <-- 【Web非公開領域】ブラウザ直接アクセス不可で安全！
└── public_html/
    └── x.nayuneko.jp/
        ├── index.html         <-- web/dist/index.html
        ├── assets/            <-- web/dist/assets/
        ├── index.cgi          <-- bin/index.cgi (パーミッション: 755)
        └── .htaccess          <-- opts/xrea/.htaccess
```

### パーミッション設定 (重要)
- `index.cgi`: **`755`** (`chmod 755 index.cgi`)

---

## 4. Basic 認証の設定

`.htaccess` 内（または XREA の環境変数設定）でユーザー名・パスワードを設定します。

```apache
# opts/xrea/.htaccess
RewriteEngine On
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^(.*)$ index.cgi/$1 [L,QSA]

# Basic 認証をかける場合（環境変数が空の場合は認証オフ）
SetEnv BASIC_AUTH_USER "admin"
SetEnv BASIC_AUTH_PASS "your_secure_password"
SetEnv DB_PATH "/virtual/YOUR_XREA_USERNAME/public_html/data/db/tweets.db"
```

※ `BASIC_AUTH_USER` と `BASIC_AUTH_PASS` を指定すると、自動的にブラウザ標準の Basic 認証プロンプトが表示されるようになります。

---

## 5. データの更新方法 (インポート・差し替え)

ローカルPCで `make import` を実行して更新した最新の `tweets.db` を、FTP / SCP / rsync で XREA サーバー上の `data/db/tweets.db` へ転送・上書き差し替えするだけで更新が完了します。
