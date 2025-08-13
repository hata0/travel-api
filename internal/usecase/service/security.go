package service

import (
	"context"

	"github.com/hata0/travel-api/internal/domain"
)

//go:generate mockgen -destination mock/time.go github.com/hata0/travel-api/internal/usecase/service SecurityService
type SecurityService interface {
	CheckRevokedToken(ctx context.Context, token string) error
	HandleTokenReuseAttack(ctx context.Context, userID domain.UserID) error
}
