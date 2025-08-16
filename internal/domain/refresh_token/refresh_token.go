package refreshtoken

import (
	"time"

	"github.com/hata0/travel-api/internal/domain/user"
)

type RefreshToken struct {
	id        RefreshTokenID
	userID    user.UserID
	token     string
	expiresAt time.Time
	createdAt time.Time
}

func NewRefreshToken(id RefreshTokenID, userID user.UserID, token string, expiresAt, createdAt time.Time) *RefreshToken {
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
func (rt *RefreshToken) UserID() user.UserID  { return rt.userID }
func (rt *RefreshToken) Token() string        { return rt.token }
func (rt *RefreshToken) ExpiresAt() time.Time { return rt.expiresAt }
func (rt *RefreshToken) CreatedAt() time.Time { return rt.createdAt }

func (rt *RefreshToken) IsExpired(now time.Time) bool {
	return now.After(rt.expiresAt)
}

func (rt *RefreshToken) Equals(other *RefreshToken) bool {
	if other == nil {
		return false
	}
	return rt.id.Equals(other.id)
}
