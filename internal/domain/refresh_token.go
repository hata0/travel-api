package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/refresh_token.go github.com/hata0/travel-api/internal/domain RefreshTokenRepository
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)
	Delete(ctx context.Context, id RefreshTokenID) error
	DeleteByUserID(ctx context.Context, userID UserID) error
}

type RefreshTokenID struct {
	value string
}

func NewRefreshTokenID(id string) RefreshTokenID {
	return RefreshTokenID{value: id}
}

func (id RefreshTokenID) String() string {
	return id.value
}

func (id RefreshTokenID) Equals(other RefreshTokenID) bool {
	return id.value == other.value
}

type RefreshToken struct {
	id        RefreshTokenID
	userID    UserID
	token     string
	expiresAt time.Time
	createdAt time.Time
}

func NewRefreshToken(id RefreshTokenID, userID UserID, token string, expiresAt, createdAt time.Time) *RefreshToken {
	return &RefreshToken{
		id:        id,
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

// Getters
func (rt *RefreshToken) ID() RefreshTokenID   { return rt.id }
func (rt *RefreshToken) UserID() UserID       { return rt.userID }
func (rt *RefreshToken) Token() string        { return rt.token }
func (rt *RefreshToken) ExpiresAt() time.Time { return rt.expiresAt }
func (rt *RefreshToken) CreatedAt() time.Time { return rt.createdAt }

func (rt *RefreshToken) Equals(other *RefreshToken) bool {
	if other == nil {
		return false
	}
	return rt.id.Equals(other.id)
}
