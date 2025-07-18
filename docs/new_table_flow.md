# 新規テーブル追加フロー

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