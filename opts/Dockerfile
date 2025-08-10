FROM mariadb:10.6

# 必要パッケージ
RUN apt update \
    && apt install -y mariadb-plugin-mroonga \
    && apt install -y software-properties-common lsb-release \
    && add-apt-repository -y universe \
    && add-apt-repository "deb http://security.ubuntu.com/ubuntu $(lsb_release --short --codename)-security main restricted" \
    && add-apt-repository -y ppa:groonga/ppa \
    && apt update \
    && apt install -y groonga-tokenizer-mecab \
    && apt clean \
    && rm -rf /var/lib/apt/lists/* \
    && ln -s /usr/share/mysql/mroonga/install.sql /docker-entrypoint-initdb.d

# NEologdの導入スクリプト
COPY opts/setup-neologd.sh /setup-neologd.sh
RUN bash /setup-neologd.sh && rm /setup-neologd.sh

# 初期化SQLを配置
COPY sql/schema.sql /docker-entrypoint-initdb.d/