# 作業手順

## ドメイン層の作成

ドメイン層は、アプリケーションの核となるビジネスロジックとエンティティを定義します。
ここでは、特定の永続化技術やフレームワークに依存しない、純粋なビジネスルールを記述します。

### 1. エンティティの定義

ドメインの核となるエンティティ（例: `Trip`）を定義します。
エンティティは、そのドメインオブジェクトが持つべき属性と、その属性に対する操作をメソッドとして持ちます。

-   **ファイルパス**: `internal/domain/<entity_name>.go` (例: `internal/domain/trip.go`)
-   **内容**:
    -   エンティティを表す構造体を定義します。
    -   エンティティのIDには、`TripID`のような型エイリアスを定義し、型安全性を高めます。
    -   エンティティの生成には、不変性を保つためのコンストラクタ関数（例: `NewTrip`）を定義します。
    -   エンティティの状態を変更する操作は、値レシーバのメソッドとして定義し、新しいインスタンスを返すことで不変性を維持します（例: `Trip.Update`）。

```go
// internal/domain/trip.go の例
type TripID string // IDの型エイリアス

type Trip struct {
	ID        TripID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTrip は新しいTripエンティティを作成します。
func NewTrip(id TripID, name string, createdAt time.Time, updatedAt time.Time) Trip {
	return Trip{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// Update はTripの名前と更新日時を更新し、新しいTripエンティティを返します。
func (t Trip) Update(name string, updatedAt time.Time) Trip {
	return Trip{
		ID:        t.ID,
		Name:      name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: updatedAt,
	}
}
```

### 2. リポジトリインターフェースの定義

エンティティの永続化操作を抽象化するリポジトリインターフェース（例: `TripRepository`）を定義します。
このインターフェースは、データベースなどの具体的な永続化層の実装からドメイン層を分離します。

-   **ファイルパス**: `internal/domain/<entity_name>.go` (エンティティと同じファイルに定義することが多い)
-   **内容**:
    -   CRUD (Create, Read, Update, Delete) 操作に対応するメソッドを定義します。
    -   各メソッドは `context.Context` を第一引数に取ります。
    -   エラーハンドリングのために、ドメイン固有のエラー（例: `ErrTripNotFound`）を返します。

```go
// internal/domain/trip.go の例
//go:generate mockgen -destination mock/trip.go travel-api/internal/domain TripRepository
type TripRepository interface {
	FindByID(ctx context.Context, id TripID) (Trip, error)
	FindMany(ctx context.Context) ([]Trip, error)
	Create(ctx context.Context, trip Trip) error
	Update(ctx context.Context, trip Trip) error
	Delete(ctx context.Context, trip Trip) error
}
```
`//go:generate mockgen` コメントを追加することで、`go generate` コマンド実行時にテスト用のモック実装が自動生成されます。

### 3. エラーの定義

発生する可能性のあるビジネスエラーをドメイン層に定義します。これにより、エラーの種類を明確にし、適切なエラーハンドリングを促します。

-   **共通エラー**: `internal/domain/error.go` に定義します。
-   **エンティティ固有のエラー**: 各エンティティのファイル（例: `internal/domain/trip.go`）に定義します。

**例 (`internal/domain/error.go`):**
```go
package domain

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
)
```

**例 (`internal/domain/trip.go`):**
```go
package domain

import "errors"

var (
	ErrTripNotFound = errors.New("trip not found")
)
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

クエリには `sqlc` が認識できる特別なコメントを追加します。例: `-- name: CreateUser :exec`

### 4. Goコードの自動生成

以下のコマンドを実行して、記述したクエリからGoのコードを自動生成します。

```bash
sqlc generate
```

これにより、`sqlc.yml` の設定に従って、`internal/infrastructure/postgres/` ディレクトリに `users.sql.go` のようなファイルが生成され、Goのコードからデータベース操作が可能になります。
