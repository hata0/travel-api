package injector

import (
	"travel-api/internal/domain"
	"travel-api/internal/infrastructure/postgres"
	"travel-api/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Services はドメインサービスの実装を提供する
type Services struct {
	clock              domain.Clock
	uuidGenerator      domain.UUIDGenerator
	transactionManager usecase.TransactionManager
}

// NewServices はサービスを初期化する
func NewServices(db *pgxpool.Pool) *Services {
	return &Services{
		clock:              &domain.SystemClock{},
		uuidGenerator:      &domain.DefaultUUIDGenerator{},
		transactionManager: postgres.NewTransactionManager(db),
	}
}

func (s *Services) Clock() domain.Clock {
	return s.clock
}

func (s *Services) UUIDGenerator() domain.UUIDGenerator {
	return s.uuidGenerator
}

func (s *Services) TransactionManager() usecase.TransactionManager {
	return s.transactionManager
}
