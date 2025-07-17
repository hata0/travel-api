# travel-api

## 概要

旅行プランを管理するためのWeb APIです。

## 技術スタック

- Go 1.24.5
- Gin
- PostgreSQL
- sqlc

## ドキュメント

APIの仕様については、[docs/api.md](./docs/api.md) を参照してください。

## ディレクトリ構成

```
cmd/               # アプリケーションのエントリーポイント
internal/          # 内部的なアプリケーションロジック
  domain/          # ドメインモデルとビジネスロジック
  infrastructure/  # データベースなどの外部依存の実装
  interface/       # リクエストとレスポンスのハンドラ
  usecase/         # アプリケーション固有のビジネスルール
```

## 開発環境の構築

1. **リポジトリをクローンします。**

   ```bash
   git clone git@github.com:hata0/travel-api.git
   cd travel-api
   ```

2. **Dockerコンテナを起動します。**

   ```bash
   docker-compose up -d
   ```

3. **APIサーバーを起動します。**

   ```bash
   go run cmd/server/main.go
   ```

   サーバーは `http://localhost:8080` で起動します。

## テスト

```bash
go test ./...
```
