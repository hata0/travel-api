package domain

import (
	"context"
	"time"
	"travel-api/internal/domain/shared/errors"
	"travel-api/internal/domain/shared/uuid"
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
	if !uuid.IsValidUUID(id) {
		return RevokedTokenID{}, errors.ErrInvalidUUID
	}
	return RevokedTokenID{value: id}, nil
}

func (id RevokedTokenID) String() string {
	return id.value
}

type RevokedToken struct {
	ID        RevokedTokenID
	UserID    UserID
	TokenJTI  string
	ExpiresAt time.Time
	RevokedAt time.Time
}

func NewRevokedToken(id RevokedTokenID, userID UserID, tokenJTI string, expiresAt, revokedAt time.Time) RevokedToken {
	return RevokedToken{
		ID:        id,
		UserID:    userID,
		TokenJTI:  tokenJTI,
		ExpiresAt: expiresAt,
		RevokedAt: revokedAt,
	}
}
