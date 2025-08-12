package services

import "time"

//go:generate mockgen -destination mock/time.go github.com/hata0/travel-api/internal/usecase/services TimeProvider
type TimeProvider interface {
	Now() time.Time
}
