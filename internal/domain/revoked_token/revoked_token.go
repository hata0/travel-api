package revokedtoken

import (
	"time"

	"github.com/hata0/travel-api/internal/domain/user"
)

type RevokedToken struct {
	id        RevokedTokenID
	userID    user.UserID
	tokenJTI  string
	expiresAt time.Time
	revokedAt time.Time
}

func NewRevokedToken(id RevokedTokenID, userID user.UserID, tokenJTI string, expiresAt, revokedAt time.Time) *RevokedToken {
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
func (rt *RevokedToken) UserID() user.UserID  { return rt.userID }
func (rt *RevokedToken) TokenJTI() string     { return rt.tokenJTI }
func (rt *RevokedToken) ExpiresAt() time.Time { return rt.expiresAt }
func (rt *RevokedToken) RevokedAt() time.Time { return rt.revokedAt }

func (rt *RevokedToken) Equals(other *RevokedToken) bool {
	if other == nil {
		return false
	}
	return rt.id.Equals(other.id)
}
