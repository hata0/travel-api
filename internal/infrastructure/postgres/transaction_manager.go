package postgres

import (
	"context"
	shared_errors "travel-api/internal/shared/errors"

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
		return shared_errors.NewInternalError("transaction failed", err)
	}
	defer tx.Rollback(ctx) // 関数が終了する際に常にロールバックを試みる

	err = fn(context.WithValue(ctx, txKey{}, tx)) // トランザクションをコンテキストに格納
	if err != nil {
		return err // fnがエラーを返したらロールバックされる
	}

	return tx.Commit(ctx) // エラーがなければコミット
}

func GetTxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
