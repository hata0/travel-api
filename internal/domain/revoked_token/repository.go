package revokedtoken

import "context"

//go:generate mockgen -destination mock/revoked_token.go github.com/hata0/travel-api/internal/domain/revokedtoken RevokedTokenRepository
type RevokedTokenRepository interface {
	Create(ctx context.Context, token *RevokedToken) error
	FindByJTI(ctx context.Context, jti string) (*RevokedToken, error)
}
