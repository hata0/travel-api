# travel-api

## 概要

旅行プランを管理するためのWeb APIです。

## 技術スタック

- 言語: Go 1.24.5
- Webフレームワーク: Gin
- データベース: PostgreSQL
- ORM/SQLジェネレータ: sqlc
- マイグレーションツール: golang-migrate/migrate
- テストツール:
  - testify: アサーションライブラリ
  - testcontainers-go: Dockerコンテナ管理 (テスト用)
  - uber-go/mock: 自動モック生成

## ディレクトリ構成

```
cmd/               # アプリケーションのエントリーポイント
internal/          # 内部的なアプリケーションロジック
  domain/          # コアビジネスロジックとエンティティ
  infrastructure/  # データベースなどの外部依存の実装
  interface/       # リクエストとレスポンスのハンドラ
  usecase/         # アプリケーション固有のビジネスルール
  config/          # アプリケーション設定
  injector/        # 依存性注入
  router/          # ルーティング
  server/          # サーバー起動ロジック
```

## 開発環境の構築

1.  **リポジトリをクローンします。**

    ```bash
    git clone git@github.com:hata0/travel-api.git
    cd travel-api
    ```

2.  **Dockerコンテナを起動します。**

    ```bash
    docker-compose up -d
    ```

3.  **APIサーバーを起動します。**

    ```bash
    go run cmd/server/main.go
    ```

    サーバーは `http://localhost:8080` で起動します。

## テスト

```bash
go test ./...
```

## ドキュメント

より詳細な開発ガイドラインや各層の実装方針については、以下のドキュメントを参照してください。

- [開発ガイドライン](./docs/README.md)
  - [API一覧](./docs/README.md#api一覧)
  - [作業手順](./docs/README.md#作業手順)
  - [共通ロジックの実装内容](./docs/README.md#共通ロジックの実装内容)
  - [cmdディレクトリ](./docs/cmd.md)
  - [configディレクトリ](./docs/config.md)
  - [ドメイン層の作成](./docs/domain.md)
  - [injectorディレクトリ](./docs/injector.md)
  - [インターフェース層の作成](./docs/interface.md)
  - [新規テーブル追加フロー](./docs/new_table_flow.md)
  - [リポジトリの実装方針](./docs/repository_guideline.md)
  - [ユースケース層の作成](./docs/usecase.md)
