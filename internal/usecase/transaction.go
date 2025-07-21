package usecase

import "context"

// TransactionManager はトランザクションを管理するためのインターフェースです。
//
//go:generate mockgen -destination mock/transaction.go travel-api/internal/usecase TransactionManager
type TransactionManager interface {
	// RunInTx は指定された関数をデータベーストランザクション内で実行します。
	// 関数がエラーを返した場合、トランザクションはロールバックされます。
	// それ以外の場合、トランザクションはコミットされます。
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
