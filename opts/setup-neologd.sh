#!/bin/bash
set -eux

# NEologdのインストール
git clone --depth 1 https://github.com/neologd/mecab-ipadic-neologd.git
cd mecab-ipadic-neologd

# 自動インストール（無駄な対話オプションはカット）
echo yes | ./bin/install-mecab-ipadic-neologd -n -y

# NEologdのパスを取得
NEOLOGD_PATH=$(./bin/install-mecab-ipadic-neologd -n -y 2>&1 | grep '/usr/lib/mecab/dic/mecab-ipadic-neologd' | head -n 1)

# mecabrcの設定を更新
sed -i "s|^dicdir.*|dicdir = ${NEOLOGD_PATH}|" /etc/mecabrc