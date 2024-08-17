# goのバージョン
FROM golang:1.22.5-alpine3.20

# OSのインストール済みのパッケージをバージョンアップ&必要なパッケージをインストール
RUN apk upgrade --update && \
    apk --no-cache add git gcc musl-dev

# ワーキングディレクトリの設定
WORKDIR /go/src/app/

# ホストPCの[./src]フォルダーを[/go/src/app/]フォルダーにコピー
ADD ./src .

# アプリケーションをビルド
RUN go build -o /go/bin/app .

# コンテナ起動時に実行するコマンド
CMD ["/go/bin/app"]
