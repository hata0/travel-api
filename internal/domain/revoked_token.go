package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/revoked_token.go travel-api/internal/domain RevokedTokenRepository
type RevokedTokenRepository interface {
	Create(ctx context.Context, token RevokedToken) error
	FindByJTI(ctx context.Context, jti string) (RevokedToken, error)
}

type RevokedTokenID struct {
	value string
}

func NewRevokedTokenID(id string) (RevokedTokenID, error) {
	if !IsValidUUID(id) {
		return RevokedTokenID{}, ErrInvalidUUID
	}
	return RevokedTokenID{value: id}, nil
}

func (id RevokedTokenID) String() string {
	return id.value
}

type RevokedToken struct {
	ID        RevokedTokenID
	TokenJTI  string
	ExpiresAt time.Time
	RevokedAt time.Time
}

func NewRevokedToken(id RevokedTokenID, tokenJTI string, expiresAt, revokedAt time.Time) RevokedToken {
	return RevokedToken{
		ID:        id,
		TokenJTI:  tokenJTI,
		ExpiresAt: expiresAt,
		RevokedAt: revokedAt,
	}
}
