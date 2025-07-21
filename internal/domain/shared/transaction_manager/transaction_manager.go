package transaction_manager

import (
	"context"
)

// TransactionManager はトランザクションを管理するためのインターフェースです。
//
//go:generate mockgen -destination mock/transaction_manager.go travel-api/internal/domain/shared/transaction_manager TransactionManager
type TransactionManager interface {
	// RunInTx は指定された関数をデータベーストランザクション内で実行します。
	// 関数がエラーを返した場合、トランザクションはロールバックされます。
	// それ以外の場合、トランザクションはコミットされます。
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
