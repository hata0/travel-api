# リポジトリの実装方針

リポジトリ層は、ドメイン層と永続化層（データベースなど）の間の抽象化を提供し、ドメイン層が特定のデータベース実装に依存しないようにします。
ここでは、`internal/infrastructure/postgres/trip_postgres.go`と`internal/infrastructure/postgres/trip_postgres_test.go`を参考に、リポジトリの実装方針について具体的に説明します。

### 1. ドメイン層との分離

リポジトリは、`internal/domain`で定義されたインターフェース（例: `domain.TripRepository`）を実装します。これにより、ドメイン層はデータベースの具体的な実装詳細を知る必要がなく、ビジネスロジックに集中できます。

コンストラクタでは、`sqlc`が利用する`DBTX`インターフェース（`pgx.Tx`または`*pgxpool.Pool`のどちらも満たす）を受け取ります。これにより、通常のDBプールだけでなく、トランザクション内でもリポジトリの操作を実行できるようになり、柔軟性が向上します。

**例 (`trip_postgres.go`):**
```go
package postgres

import (
	"context"
	"travel-api/internal/domain"
	// ...
)

// TripPostgresRepository は domain.TripRepository インターフェースを実装します。
type TripPostgresRepository struct {
	queries *Queries // sqlcが生成したクエリ
}

// NewTripPostgresRepository は新しいリポジトリインスタンスを生成します。
func NewTripPostgresRepository(db DBTX) domain.TripRepository {
	return &TripPostgresRepository{
		queries: New(db),
	}
}
```

### 2. `sqlc`の活用と型マッピング

データベース操作には、`sqlc`によってSQLから自動生成されたGoコードを使用します。これにより、手書きの定型的なコードを減らし、型安全なデータベースアクセスを実現します。

リポジトリの重要な責務の一つが、ドメイン層で定義された型（例: `domain.TripID`, `time.Time`）と、`sqlc`が生成する`pgtype`パッケージの型（例: `pgtype.UUID`, `pgtype.Timestamptz`）との間のマッピングです。

-   **DBからの読み取り時**: `pgtype` からドメインの型へ変換します。このロジックは `mapToTrip` のようなプライベートなヘルパー関数にカプセル化すると良いでしょう。
-   **DBへの書き込み時**: ドメインの型から `pgtype` へ変換します。

**例 (`trip_postgres.go`):**
```go
// mapToTrip はDBのレコード(Trip)をドメインオブジェクト(domain.Trip)に変換します。
func (r *TripPostgresRepository) mapToTrip(record Trip) domain.Trip {
	var id domain.TripID
	if record.ID.Valid {
		// pgtype.UUID から domain.TripID へ
		id, _ = domain.NewTripID(record.ID.String())
	}
    // ...
	return domain.NewTrip(id, record.Name, createdAt, updatedAt)
}

// Create はドメインオブジェクト(domain.Trip)を受け取り、DBにレコードを作成します。
func (r *TripPostgresRepository) Create(ctx context.Context, trip domain.Trip) error {
	var validatedId pgtype.UUID
	// domain.TripID から pgtype.UUID へ
	_ = validatedId.Scan(trip.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	// time.Time から pgtype.Timestamptz へ
	_ = validatedCreatedAt.Scan(trip.CreatedAt)

    // ...

	// sqlcが生成したCreateTrip関数を呼び出す
	return r.queries.CreateTrip(ctx, CreateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		CreatedAt: validatedCreatedAt,
		UpdatedAt: validatedUpdatedAt,
	})
}
```

### 3. エラーハンドリング

データベース操作中に発生したエラーは、リポジトリ層でキャッチし、ドメイン層で定義された適切なエラー（例: `domain.ErrTripNotFound`）に変換して返します。これにより、上位層（ユースケース層やハンドラ層）は、`pgx.ErrNoRows`のようなデータベース固有のエラーに依存することなく、ビジネスロジックに基づいたエラーハンドリングを行えます。

**例 (`trip_postgres.go`):**
```go
func (r *TripPostgresRepository) FindByID(ctx context.Context, id domain.TripID) (domain.Trip, error) {
	// ...
	record, err := r.queries.GetTrip(ctx, validatedId)
	if err != nil {
		// pgx固有のエラーをドメインエラーに変換
		if err == pgx.ErrNoRows {
			return domain.Trip{}, domain.ErrTripNotFound
		} else {
			return domain.Trip{}, err
		}
	}
	// ...
}
```

### 4. テスト戦略

リポジトリ層のテストは、モックではなく実際のデータベース（テストコンテナ）に対して実行するインテグレーションテストです。これにより、SQLクエリやスキーマ定義を含めたデータベースとの連携が正しく機能することを保証します。

#### 4.1. テスト環境のセットアップ (`main_test.go`)

`testcontainers-go`ライブラリを使用して、テスト実行時に一時的なPostgreSQLコンテナを起動します。`TestMain`関数でテスト全体のライフサイクルを管理し、コンテナの起動、マイグレーションの実行、そして全テスト終了後のコンテナ破棄を行います。

**ポイント:**
-   **コンテナの起動**: `postgres.Run`でテスト用のDBコンテナを起動します。
-   **マイグレーション**: `golang-migrate`を使い、コンテナ起動後に最新のスキーマを適用します。
-   **スナップショット**: マイグレーション後のDBの状態をスナップショットとして保存します。これにより、各テストケースの実行前にDBを高速にクリーンな状態へ復元できます。

**例 (`main_test.go`):**
```go
func testMain(m *testing.M) int {
	ctx := context.Background()

	// 1. テストコンテナをセットアップ
	container, dbUrl, err := setupTestContainer(ctx)
	if err != nil {
		// ... エラー処理
	}
	testContainer = container
	testDbUrl = dbUrl

	// 2. テストを実行
	code := m.Run()

	// 3. コンテナを終了
	if err := testContainer.Terminate(ctx); err != nil {
		// ... エラー処理
	}

	return code
}

func setupTestContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	// PostgreSQL コンテナを起動
	container, err := postgres.Run(ctx, "postgres:17.4-bookworm", ...)
	// ...
	// 接続文字列を取得
	dbUrl, err := container.ConnectionString(ctx, "sslmode=disable")
	// ...
	// マイグレーションを実行
	mig, err := migrate.New("file://sql/migrations", dbUrl)
	// ...
	if err := mig.Up(); err != nil {
		// ...
	}
	// ...
	// スナップショットを作成
	if err := container.Snapshot(ctx, postgres.WithSnapshotName("test-db-snapshot")); err != nil {
		// ...
	}

	return container, dbUrl, nil
}
```

#### 4.2. テストケースの独立性

各テストケースは、他のテストケースから影響を受けないように、完全に独立しているべきです。`setupDB`のようなヘルパー関数を用意し、各テストの開始時にDB接続を確立し、終了時に`t.Cleanup`を使ってスナップショットを復元することで、これを実現します。

**例 (`trip_postgres_test.go`):**
```go
func setupDB(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()

	// テスト用のDBに接続
	db, err := pgx.Connect(ctx, testDbUrl)
	require.NoError(t, err)

	t.Cleanup(func() {
		// テスト終了時に接続を閉じる
		err := db.Close(ctx)
		require.NoError(t, err)
		// スナップショットを復元してDBをクリーンな状態に戻す
		err = testContainer.Restore(ctx)
		require.NoError(t, err)
	})

	return db
}

func TestTripPostgresRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	// 各テストの開始時にクリーンなDB接続を取得
	dbConn := setupDB(t, ctx)
	repo := NewTripPostgresRepository(dbConn)
    // ... テストロジック
}
```

#### 4.3. テストヘルパーの活用

テストコードの可読性と保守性を高めるために、繰り返し使用されるロジックはヘルパー関数として抽出します。

-   `createTestTrip`: テスト用のドメインオブジェクトを生成します。
-   `insertTestTrip`: テストデータをデータベースに直接挿入します。
-   `getTripFromDB`: テスト結果を検証するために、データベースから直接レコードを取得します。

**例 (`trip_postgres_test.go`):**
```go
// insertTestTrip はテスト用のTripドメインオブジェクトをデータベースに挿入するヘルパー関数
func insertTestTrip(t *testing.T, ctx context.Context, db DBTX, trip domain.Trip) {
	t.Helper()
	queries := New(db);
    // ... 型変換
	err := queries.CreateTrip(ctx, CreateTripParams{...})
	require.NoError(t, err)
}

func TestTripPostgresRepository_Create(t *testing.T) {
    // ...
	t.Run("正常系: 新しいレコードが作成される", func(t *testing.T) {
		trip := createTestTrip(t, "New Trip", now, now)

		err := repo.Create(ctx, trip)
		assert.NoError(t, err)

		// DBから直接取得して検証
		createdRecord, err := getTripFromDB(t, ctx, dbConn, trip.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, trip.Name, createdRecord.Name)
	})
}
```

#### 4.4. 網羅的なテスト

正常系（Happy Path）だけでなく、異常系やエッジケースも網羅的にテストします。

-   存在しないレコードの取得、更新、削除
-   重複するキーでの挿入（DBの制約違反）
-   空のテーブルに対する一覧取得

**例 (`trip_postgres_test.go`):**
```go
func TestTripPostgresRepository_FindByID(t *testing.T) {
    // ...
	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		id, err := domain.NewTripID(domain.NewUUID())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, id)

		assert.ErrorIs(t, err, domain.ErrTripNotFound)
	})
}
```