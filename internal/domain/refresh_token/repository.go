package refreshtoken

import (
	"context"

	"github.com/hata0/travel-api/internal/domain/user"
)

//go:generate mockgen -destination mock/refresh_token.go github.com/hata0/travel-api/internal/domain/refreshtoken RefreshTokenRepository
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	Delete(ctx context.Context, id RefreshTokenID) error
	DeleteByUserID(ctx context.Context, userID user.UserID) error
}
