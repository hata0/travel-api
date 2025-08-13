package service

import (
	"time"

	"github.com/hata0/travel-api/internal/domain"
)

//go:generate mockgen -destination mock/time.go github.com/hata0/travel-api/internal/usecase/service TokenService
type TokenService interface {
	GenerateAccessToken(userID domain.UserID, expiresAt time.Time, secret, issuer string, issuedAt time.Time) (string, error)
	GenerateRefreshToken() string
	ParseAccessToken(token, secret string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID    domain.UserID
	ExpiresAt time.Time
	IssuedAt  time.Time
	Issuer    string
}
