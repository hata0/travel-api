package di

import (
	"github.com/hata0/travel-api/internal/domain/shared/clock"
	"github.com/hata0/travel-api/internal/domain/shared/transaction_manager"
	"github.com/hata0/travel-api/internal/domain/shared/uuid"
	"github.com/hata0/travel-api/internal/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Services はドメインサービスの実装を提供する
type Services struct {
	clock              clock.Clock
	uuidGenerator      uuid.UUIDGenerator
	transactionManager transaction_manager.TransactionManager
}

// NewServices はサービスを初期化する
func NewServices(db *pgxpool.Pool) *Services {
	return &Services{
		clock:              &clock.SystemClock{},
		uuidGenerator:      &uuid.DefaultUUIDGenerator{},
		transactionManager: postgres.NewTransactionManager(db),
	}
}

func (s *Services) Clock() clock.Clock {
	return s.clock
}

func (s *Services) UUIDGenerator() uuid.UUIDGenerator {
	return s.uuidGenerator
}

func (s *Services) TransactionManager() transaction_manager.TransactionManager {
	return s.transactionManager
}
