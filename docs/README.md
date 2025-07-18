# 作業手順

## ドキュメント

- [ドメイン層の作成](./domain.md)

## 新規テーブル追加フロー
```

## 新規テーブル追加フロー

新しいテーブルをデータベースに追加し、アプリケーションで利用可能にするまでの手順です。

### 1. マイグレーションファイルの作成

`golang-migrate` を使用して、新しいテーブル定義用のマイグレーションファイルを作成します。

```bash
make migrate-new name=<migration_name>
```

`<migration_name>` には `create_users_table` のような、変更内容がわかる名前を指定してください。

これにより、`internal/infrastructure/postgres/sql/migrations/` ディレクトリに `up` と `down` の2つのSQLファイルが生成されます。

### 2. テーブル定義 (SQL) の記述

生成された `up` ファイルに `CREATE TABLE` 文を記述します。`down` ファイルには、テーブルを削除する `DROP TABLE` 文を記述してください。

**ベストプラクティス**:
- 実際の運用を想定し、アプリケーション側で値の正当性を担保する前提でスキーマを設計する: データベースレベルでの厳密なバリデーションよりも、アプリケーション層でのビジネスロジックに基づいたバリデーションを優先します。
- 必要な制約 (`NOT NULL` や `UNIQUE`) は必ず設定する
- 必須カラムには `DEFAULT` を設定せず、アプリケーションで明示的に値を渡す: アプリケーションのビジネスロジックとデータベースのスキーマ定義の乖離を防ぎ、データの一貫性をアプリケーション側で管理するためです。

**例 (`up` ファイル):**
```sql
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL, -- パスワードのハッシュ値を保存
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
```

**例 (`down` ファイル):**
```sql
DROP TABLE IF EXISTS users;
```

**アンチパターン**:
以下のような `sql` は作成しないでください。

- `TEXT` 型の代わりに `VARCHAR` を使用する:
   アプリ側で長さを担保できるため `TEXT` で十分です。不要に型を制限しないでください。
- `DEFAULT` を使用する:
  アプリケーションで値を必ずセットするべきです。
- `NOT NULL` 制約を省略する:
  意図しない `NULL` が混入し、データ不整合の原因となります。

### 3. クエリファイルの作成

`sqlc` がGoのコードを生成するために、SQLクエリを記述するファイルを作成します。

クエリファイルは `internal/infrastructure/postgres/sql/queries/` ディレクトリに配置してください。ファイル名は `<table_name>.sql` (例: `users.sql`) とします。

クエリには `sqlc` が認識できる特別なコメントを追加します。このコメントは、`sqlc` がGoの関数を生成する際のヒントとなります。

-   `-- name: <QueryName> :one`: 単一のレコードを返すクエリ。
-   `-- name: <QueryName> :many`: 複数のレコードを返すクエリ。
-   `-- name: <QueryName> :exec`: レコードを返さないクエリ（INSERT, UPDATE, DELETEなど）。

**例 (`users.sql`):**
```sql
-- name: CreateUser :exec
INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, created_at, updated_at FROM users
WHERE email = $1 LIMIT 1;
```

**ベストプラクティス**:
-   各クエリファイルは、特定のテーブルまたは関連する一連の操作に特化させます。
-   クエリ名は、そのクエリが何を行うかを明確に示します（例: `GetUserByID`, `ListUsers`, `CreateUser`）。
-   パラメータは `$1`, `$2` のようにプレースホルダを使用します。`sqlc` がこれらをGoの関数の引数にマッピングします。

### 4. Goコードの自動生成

以下のコマンドを実行して、記述したクエリからGoのコードを自動生成します。

```bash
sqlc generate
```

これにより、`sqlc.yml` の設定に従って、`internal/infrastructure/postgres/` ディレクトリに `users.sql.go` のようなファイルが生成され、Goのコードからデータベース操作が可能になります。

## リポジトリの実装方針

リポジトリ層は、ドメイン層と永続化層（データベースなど）の間の抽象化を提供し、ドメイン層が特定のデータベース実装に依存しないようにします。ここでは、`internal/infrastructure/postgres/trip_postgres.go`と`internal/infrastructure/postgres/trip_postgres_test.go`を参考に、リポジトリの実装方針について説明します。

### 1. ドメイン層との分離

-   リポジトリは、`internal/domain`で定義されたインターフェース（例: `domain.TripRepository`）を実装します。これにより、ドメイン層はデータベースの具体的な実装詳細を知る必要がなく、ビジネスロジックに集中できます。
-   リポジトリのコンストラクタは、`DBTX`インターフェース（`pgx.Tx`または`*pgxpool.Pool`）を受け取ることで、トランザクション管理を柔軟に行えるようにします。

### 2. `sqlc`の活用

-   データベース操作には`sqlc`によって自動生成されたコード（`internal/infrastructure/postgres/queries.go`など）を使用します。これにより、手書きのSQLクエリを減らし、型安全なデータベースアクセスを実現します。
-   `sqlc`が生成する`pgtype`パッケージの型（例: `pgtype.UUID`, `pgtype.Timestamptz`）と、ドメイン層の型（例: `domain.TripID`, `time.Time`）との間のマッピングは、リポジトリ層の責務となります。`mapToTrip`のようなヘルパー関数を使用して、この変換をカプセル化します。

### 3. エラーハンドリング

-   データベース操作中に発生したエラーは、適切なドメインエラー（例: `domain.ErrTripNotFound`）に変換して返します。これにより、上位層（ユースケース層やハンドラ層）はデータベース固有のエラーに依存することなく、ビジネスロジックに基づいたエラーハンドリングを行えます。
-   `pgx.ErrNoRows`のような特定のデータベースエラーは、`domain.ErrTripNotFound`に変換する典型的な例です。

### 4. テスト戦略

リポジトリ層のテストは、実際のデータベース（テストコンテナなど）を使用して行います。これにより、データベースとの統合が正しく機能していることを保証します。

-   **テストコンテナの利用**: `testcontainers-go`を使用して、テスト実行時に一時的なPostgreSQLコンテナを起動します。これにより、開発環境やCI/CD環境に依存しない、独立したテスト環境を構築できます。
-   **テストデータの独立性**: 各テストケースは、他のテストケースに影響を与えないように、独立したテストデータを使用します。テスト開始時にデータベースをクリーンな状態にし、必要なデータを各テストケース内で挿入します。
    -   `createTestTrip`: テスト用のドメインオブジェクトを生成するヘルパー関数。
    -   `insertTestTrip`: データベースにテストデータを挿入するヘルパー関数。
    -   `getTripFromDB`: データベースから直接データを取得し、テストの検証に使用するヘルパー関数。
-   **網羅的なテスト**: 正常系だけでなく、以下のような異常系もテストします。
    -   存在しないレコードの取得、更新、削除。
    -   重複するキーでの挿入（データベースの制約違反）。
-   **アサーションの具体性**: `github.com/stretchr/testify/assert`を使用して、期待される結果と実際の結果を厳密に比較します。特に`time.Time`のような型は、比較時の精度に注意し、必要に応じて`Truncate`メソッドを使用します。
-   **テストヘルパーの活用**: 繰り返し使用されるロジック（例: テストデータの生成、データベースへの挿入・取得）はヘルパー関数として抽出し、テストコードの可読性と保守性を高めます。
