package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/revoked_token.go github.com/hata0/travel-api/internal/domain RevokedTokenRepository
type RevokedTokenRepository interface {
	Create(ctx context.Context, token *RevokedToken) error
	FindByJTI(ctx context.Context, jti string) (*RevokedToken, error)
}

type RevokedTokenID struct {
	value string
}

func NewRevokedTokenID(id string) RevokedTokenID {
	return RevokedTokenID{value: id}
}

func (id RevokedTokenID) String() string {
	return id.value
}

func (id RevokedTokenID) Equals(other RevokedTokenID) bool {
	return id.value == other.value
}

type RevokedToken struct {
	id        RevokedTokenID
	userID    UserID
	tokenJTI  string
	expiresAt time.Time
	revokedAt time.Time
}

func NewRevokedToken(id RevokedTokenID, userID UserID, tokenJTI string, expiresAt, revokedAt time.Time) *RevokedToken {
	return &RevokedToken{
		id:        id,
		userID:    userID,
		tokenJTI:  tokenJTI,
		expiresAt: expiresAt,
		revokedAt: revokedAt,
	}
}

// Getters
func (rt *RevokedToken) ID() RevokedTokenID   { return rt.id }
func (rt *RevokedToken) UserID() UserID       { return rt.userID }
func (rt *RevokedToken) TokenJTI() string     { return rt.tokenJTI }
func (rt *RevokedToken) ExpiresAt() time.Time { return rt.expiresAt }
func (rt *RevokedToken) RevokedAt() time.Time { return rt.revokedAt }

func (rt *RevokedToken) Equals(other *RevokedToken) bool {
	if other == nil {
		return false
	}
	return rt.id.Equals(other.id)
}
