# トランザクション管理

このドキュメントでは、アプリケーションにおけるデータベーストランザクション管理の実装について説明します。トランザクションを導入することで、複数のデータベース操作がアトミック（不可分）に実行されることを保証し、データの一貫性と信頼性を高めます。

## 目的

-   **データの一貫性**: 複数の関連するデータベース操作がすべて成功するか、すべて失敗してロールバックされるかのどちらかであることを保証します。
-   **アトミック性**: 複雑なビジネスロジックにおいて、部分的なデータ変更を防ぎ、データベースの状態を常に整合性の取れたものに保ちます。
-   **テスト容易性**: ユースケース層がトランザクションの開始・終了を制御し、リポジトリ層がトランザクションの有無を透過的に扱うことで、各層の責務が明確になり、テストが容易になります。

## 主要なコンポーネント

### 1. `domain.TransactionManager` インターフェース

トランザクション管理の抽象化を定義します。ユースケース層はこのインターフェースに依存し、具体的な実装からは分離されます。

-   **ファイルパス**: `internal/domain/transaction.go`

```go
package domain

import "context"

// TransactionManager はトランザクションを管理するためのインターフェースです。
type TransactionManager interface {
	// RunInTx は指定された関数をデータベーストランザクション内で実行します。
	// 関数がエラーを返した場合、トランザクションはロールバックされます。
	// それ以外の場合、トランザクションはコミットされます。
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### 2. `postgres.TransactionManager` 実装

`domain.TransactionManager`インターフェースのPostgreSQL向け実装です。`pgxpool.Pool`を使用してトランザクションを開始し、コンテキストにトランザクションオブジェクトを格納します。

-   **ファイルパス**: `internal/infrastructure/postgres/transaction_manager.go`

```go
package postgres

import (
	"context"
	"travel-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// TransactionManager はpgxpool.Poolを使用したトランザクションマネージャーです。
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager は新しいTransactionManagerのインスタンスを作成します。
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// RunInTx は指定された関数をデータベーストランザクション内で実行します。
func (tm *TransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return domain.NewInternalServerError(err)
	}
	defer tx.Rollback(ctx) // 関数が終了する際に常にロールバックを試みる

	err = fn(context.WithValue(ctx, txKey{}, tx)) // トランザクションをコンテキストに格納
	if err != nil {
		return err // fnがエラーを返したらロールバックされる
	}

	return tx.Commit(ctx) // エラーがなければコミット
}

// GetTxFromContext はコンテキストから pgx.Tx オブジェクトを取得します。
// トランザクションがコンテキストに存在しない場合は false を返します。
func GetTxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
```

### 3. `postgres.BaseRepository` と `getQueries` メソッド

リポジトリがトランザクションの有無を透過的に判断し、適切な`sqlc.Queries`インスタンスを使用するための共通ロジックを提供します。すべてのPostgresリポジトリはこの`BaseRepository`を埋め込みます。

-   **ファイルパス**: `internal/infrastructure/postgres/base_repository.go`

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

## ワークフロー

1.  **ユースケース層でのトランザクション開始**: 複数のデータベース操作をアトミックに実行したいユースケース（例: ユーザーログイン時のトークン保存と削除）では、`AuthInteractor`などのインタラクターが`i.transactionManager.RunInTx(ctx, func(txCtx context.Context) error { ... })`を呼び出します。
2.  **トランザクションのコンテキストへの格納**: `RunInTx`メソッドは、新しいデータベーストランザクションを開始し、そのトランザクションオブジェクトを`txCtx`（トランザクションコンテキスト）に格納して、コールバック関数に渡します。
3.  **リポジトリ層での透過的なトランザクション利用**: コールバック関数内でリポジトリのメソッド（例: `i.userRepository.FindByEmail(txCtx, email)`）が呼び出されると、リポジトリ内部の`getQueries`メソッドが`txCtx`をチェックします。
4.  **適切な`Queries`インスタンスの取得**: `getQueries`は`GetTxFromContext`を使用して`txCtx`からトランザクションオブジェクトを抽出し、それを使って`sqlc.New(tx)`を呼び出し、トランザクション対応の`Queries`インスタンスを生成します。トランザクションが存在しない場合は、通常のDBプールを使用する`Queries`インスタンスを返します。
5.  **アトミックなデータベース操作**: リポジトリメソッドは、取得した`Queries`インスタンスを使用してデータベース操作を実行します。これにより、すべての操作が同じトランザクション内で実行されることが保証されます。
6.  **トランザクションのコミット/ロールバック**: `RunInTx`メソッドは、コールバック関数がエラーを返さずに終了した場合にトランザクションをコミットし、エラーを返した場合にロールバックします。

## 利点

-   **簡潔なユースケースロジック**: ユースケース層はトランザクションの開始、コミット、ロールバックといった定型的な処理から解放され、純粋なビジネスロジックに集中できます。
-   **透過的なリポジトリ操作**: リポジトリ層は、呼び出し元がトランザクション内で実行しているかどうかを意識することなく、常に適切なデータベース操作を実行できます。
-   **堅牢なエラーハンドリング**: トランザクションマネージャーがエラー発生時のロールバックを自動的に処理するため、エラーハンドリングが簡素化され、バグのリスクが低減します。
-   **高いテスト容易性**: ユースケース層のテストでは`TransactionManager`をモック化し、`RunInTx`が渡された関数を実行するように設定することで、データベースに依存しない単体テストが可能です。
