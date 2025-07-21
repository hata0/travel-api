package domain

import (
	"context"
	"time"
	"travel-api/internal/domain/shared/uuid"
)

//go:generate mockgen -destination mock/refresh_token.go travel-api/internal/domain RefreshTokenRepository
type RefreshTokenRepository interface {
	Create(ctx context.Context, token RefreshToken) error
	FindByToken(ctx context.Context, token string) (RefreshToken, error)
	Delete(ctx context.Context, token RefreshToken) error
	DeleteByUserID(ctx context.Context, userID UserID) error
}

type RefreshTokenID struct {
	value string
}

func NewRefreshTokenID(id string) (RefreshTokenID, error) {
	if !uuid.IsValidUUID(id) {
		return RefreshTokenID{}, ErrInvalidUUID
	}
	return RefreshTokenID{value: id}, nil
}

func (id RefreshTokenID) String() string {
	return id.value
}

type RefreshToken struct {
	ID        RefreshTokenID
	UserID    UserID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewRefreshToken(id RefreshTokenID, userID UserID, token string, expiresAt, createdAt time.Time) RefreshToken {
	return RefreshToken{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}
}
