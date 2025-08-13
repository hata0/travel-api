package service

import "time"

//go:generate mockgen -destination mock/time.go github.com/hata0/travel-api/internal/usecase/service TimeService
type TimeService interface {
	Now() time.Time
}
