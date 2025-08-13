package service

import (
	"github.com/hata0/travel-api/internal/domain"
)

//go:generate mockgen -destination mock/time.go github.com/hata0/travel-api/internal/usecase/service TokenService
type TokenService interface {
	GenerateAccessToken(userID domain.UserID) (string, error)
	GenerateRefreshToken() (string, error)
}
