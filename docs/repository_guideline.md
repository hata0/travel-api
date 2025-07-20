# リポジトリの実装方針

リポジトリ層は、ドメイン層と永続化層（データベースなど）の間の抽象化を提供し、ドメイン層が特定のデータベース実装に依存しないようにします。
ここでは、`internal/infrastructure/postgres/trip_postgres.go`と`internal/infrastructure/postgres/trip_postgres_test.go`を参考に、リポジトリの実装方針について具体的に説明します。

### 1. ドメイン層との分離とトランザクションの透過的利用

リポジトリは、`internal/domain`で定義されたインターフェース（例: `domain.TripRepository`）を実装します。これにより、ドメイン層はデータベースの具体的な実装詳細を知る必要がなく、ビジネスロジックに集中できます。

トランザクションの透過的な利用を可能にするため、すべてのPostgresリポジトリは共通の`BaseRepository`を埋め込みます。`BaseRepository`は、`sqlc`が利用する`DBTX`インターフェース（`pgx.Tx`または`*pgxpool.Pool`のどちらも満たす）を受け取り、コンテキストからトランザクションの有無を判断して適切な`Queries`インスタンスを提供する`getQueries`メソッドを提供します。

これにより、リポジトリのコンストラクタはDBプールまたはトランザクションを受け入れ、リポジトリ内の各メソッドはトランザクションの有無を意識することなくデータベース操作を実行できます。

**例 (`internal/infrastructure/postgres/base_repository.go`):**
```go
package postgres

import "context"

// BaseRepository はすべてのPostgresリポジトリに共通の機能を提供します。
type BaseRepository struct {
	db DBTX
}

// NewBaseRepository は新しいBaseRepositoryのインスタンスを作成します。
func NewBaseRepository(db DBTX) *BaseRepository {
	return &BaseRepository{db: db}
}

// getQueries はコンテキストからトランザクションを取得し、それに応じたQueriesインスタンスを返します。
// ユースケース層でトランザクションが開始されている場合、そのトランザクションを使用します。
// そうでない場合は、通常のDBプールを使用します。
func (r *BaseRepository) getQueries(ctx context.Context) *Queries {
	if tx, ok := GetTxFromContext(ctx); ok {
		return New(tx) // トランザクションがあればトランザクション対応のQueriesを返す
	}
	return New(r.db) // なければ通常のDBプール対応のQueriesを返す
}
```

**例 (`trip_postgres.go`):**
```go
package postgres

import (
	"context"
	"travel-api/internal/domain"
	// ...
)

// TripPostgresRepository は domain.TripRepository インターフェースを実装します。
// BaseRepository を埋め込むことで、共通のgetQueriesメソッドを利用できます。
type TripPostgresRepository struct {
	*BaseRepository
}

// NewTripPostgresRepository は新しいリポジトリインスタンスを生成します。
func NewTripPostgresRepository(db DBTX) domain.TripRepository {
	return &TripPostgresRepository{
		BaseRepository: NewBaseRepository(db),
	}
}
```

### 2. `sqlc`の活用と型マッピング

データベース操作には、`sqlc`によってSQLから自動生成されたGoコードを使用します。これにより、手書きの定型的なコードを減らし、型安全なデータベースアクセスを実現します。

リポジトリの重要な責務の一つが、ドメイン層で定義された型（例: `domain.TripID`, `time.Time`）と、`sqlc`が生成する`pgtype`パッケージの型（例: `pgtype.UUID`, `pgtype.Timestamptz`）との間のマッピングです。

各リポジトリメソッド内では、`r.getQueries(ctx)`を呼び出すことで、現在のコンテキストにトランザクションが存在するかどうかに応じて、適切な`Queries`インスタンス（トランザクション対応または非トランザクション対応）が自動的に取得されます。これにより、リポジトリのメソッドはトランザクションの有無を意識することなく、常に正しい`Queries`インスタンスを使用してデータベース操作を実行できます。

-   **DBからの読み取り時**: `pgtype` からドメインの型へ変換します。このロジックは `mapToTrip` のようなプライベートなヘルパー関数にカプセル化します。
-   **DBへの書き込み時**: ドメインの型から `pgtype` へ変換します。

**例 (`trip_postgres.go`):**
```go
// mapToTrip はDBのレコード(Trip)をドメインオブジェクト(domain.Trip)に変換します。
func (r *TripPostgresRepository) mapToTrip(record Trip) domain.Trip {
	var id domain.TripID
	if record.ID.Valid {
		id, _ = domain.NewTripID(record.ID.String())
	}

	var createdAt time.Time
	if record.CreatedAt.Valid {
		createdAt = record.CreatedAt.Time
	}

	var updatedAt time.Time
	if record.UpdatedAt.Valid {
		updatedAt = record.UpdatedAt.Time
	}

	return domain.NewTrip(
		id,
		record.Name,
		createdAt,
		updatedAt,
	)
}

// Create はドメインオブジェクト(domain.Trip)を受け取り、DBにレコードを作成します。
func (r *TripPostgresRepository) Create(ctx context.Context, trip domain.Trip) error {
	queries := r.getQueries(ctx) // getQueries を使用して適切な Queries インスタンスを取得

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(trip.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	// sqlcが生成したCreateTrip関数を呼び出す
	if err := queries.CreateTrip(ctx, CreateTripParams{ // 取得したqueriesを使用
		ID:        validatedId,
		Name:      trip.Name,
		CreatedAt: validatedCreatedAt,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
        // ... エラーハンドリング ...
    }
    return nil
}
```

### 3. エラーハンドリング

データベース操作中に発生したエラーは、リポジトリ層でハンドリングし、ドメイン層で定義された適切なエラーに変換して返します。これにより、上位層（ユースケース層）は、データベース固有のエラー型に依存することなく、ビジネスロジックに基づいたエラーハンドリングを行えます。

-   **Read (`Find...`系メソッド)**: レコードが存在しないことを示す `pgx.ErrNoRows` を、`domain.ErrTripNotFound` のようなドメイン固有の「Not Found」エラーに変換します。
-   **Create (`Create`メソッド)**: 主キーやユニークキーの重複違反 (`pgconn.PgError` の `Code: "23505"`) を検知し、`domain.ErrTripAlreadyExists` のようなドメイン固有の「Already Exists」エラーに変換します。
-   **Update/Delete**: これらのメソッドでは、対象レコードが存在しない場合でも `sqlc` はエラーを返しません。ユースケース層で事前に存在確認を行う設計のため、リポジトリ層ではエラーハンドリングは不要です。
-   **その他のエラー**: 上記以外の予期せぬデータベースエラーは、すべて `domain.NewInternalServerError(err)` でラップし、根本原因を保持しつつ、上位層には一貫した内部サーバーエラーとして報告します。

**例 (`trip_postgres.go`):**
```go
// FindByIDでのエラーハンドリング
func (r *TripPostgresRepository) FindByID(ctx context.Context, id domain.TripID) (domain.Trip, error) {
	queries := r.getQueries(ctx) // getQueries を使用して適切な Queries インスタンスを取得
	var validatedId pgtype.UUID
	_ = validatedId.Scan(id.String())

	record, err := queries.GetTrip(ctx, validatedId) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Trip{}, domain.ErrTripNotFound // Not Foundエラーに変換
		} else {
			return domain.Trip{}, domain.NewInternalServerError(err) // その他のエラーをラップ
		}
	}
	return r.mapToTrip(record), nil
}

// Createでのエラーハンドリング
func (r *TripPostgresRepository) Create(ctx context.Context, trip domain.Trip) error {
    queries := r.getQueries(ctx) // getQueries を使用して適切な Queries インスタンスを取得
    // ...
	if err := queries.CreateTrip(ctx, params); err != nil { // 取得したqueriesを使用
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			return domain.ErrTripAlreadyExists // Already Existsエラーに変換
		}
		return domain.NewInternalServerError(err) // その他のエラーをラップ
	}
	return nil
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
		log.Printf("error setting up test container: %v", err)
		return 1
	}
	testContainer = container
	testDbUrl = dbUrl

	// 2. テストを実行
	code := m.Run()

	// 3. コンテナを終了
	if err := testContainer.Terminate(ctx); err != nil {
		log.Printf("error terminating test container: %v", err)
	}

	return code
}
```

#### 4.2. テストケースの独立性

各テストケースは、他のテストケースから影響を受けないように、完全に独立しているべきです。`setupDB`のようなヘルパー関数を用意し、各テストの開始時にDB接続を確立し、終了時に`t.Cleanup`を使ってスナップショットを復元することで、これを実現します。

**例 (`trip_postgres_test.go`):**
```go
func setupDB(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()

	db, err := pgx.Connect(ctx, testDbUrl)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close(ctx)
		require.NoError(t, err)
		// スナップショットを復元してDBをクリーンな状態に戻す
		err = testContainer.Restore(ctx)
		require.NoError(t, err)
	})

	return db
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
    // ... 型変換とDBへの挿入 ...
	err := queries.CreateTrip(ctx, CreateTripParams{...})
	require.NoError(t, err)
}
```

#### 4.4. 網羅的なテスト

正常系（Happy Path）だけでなく、リポジトリが返しうるすべてのエラーパターンを網羅的にテストします。

-   **`FindByID`**: 正常にレコードが取得できるケースと、レコードが存在しない場合に`domain.ErrTripNotFound`が返るケースをテストします。
-   **`FindMany`**: レコードが複数存在する場合と、1件も存在しない場合に空のスライスが返るケースをテストします。
-   **`Create`**: 正常にレコードが作成できるケースと、主キーが重複した場合に`domain.ErrTripAlreadyExists`が返るケースをテストします。
-   **`Update`/`Delete`**: 正常にレコードが更新・削除できるケースをテストします。ユースケース層で存在確認を行うため、リポジトリ層のテストでは対象レコードが存在する正常系のみをテスト対象とします。

**例 (`trip_postgres_test.go`):**
```go
func TestTripPostgresRepository_FindByID(t *testing.T) {
    // ...
	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		id, err := domain.NewTripID(uuid.New().String())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, id)

		assert.ErrorIs(t, err, domain.ErrTripNotFound)
	})
}

func TestTripPostgresRepository_Create(t *testing.T) {
    // ...
	t.Run("異常系: 重複するIDで作成", func(t *testing.T) {
		// ...
		trip := createTestTrip(t, "Existing Trip", now, now)
		insertTestTrip(t, ctx, dbConn, trip) // 最初に挿入

		err := repo.Create(ctx, trip) // 同じIDで再度挿入
		assert.ErrorIs(t, err, domain.ErrTripAlreadyExists)
	})
}