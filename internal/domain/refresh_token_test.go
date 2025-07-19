package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewRefreshTokenID(t *testing.T) {
	t.Run("正常系: 有効なUUID", func(t *testing.T) {
		validUUID := uuid.New().String()
		refreshTokenID, err := NewRefreshTokenID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, RefreshTokenID{value: validUUID}, refreshTokenID)
	})

	t.Run("異常系: 無効なUUID", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		refreshTokenID, err := NewRefreshTokenID(invalidUUID)
		assert.ErrorIs(t, err, ErrInvalidUUID)
		assert.Equal(t, RefreshTokenID{}, refreshTokenID)
	})

	t.Run("異常系: 空文字列", func(t *testing.T) {
		emptyUUID := ""
		refreshTokenID, err := NewRefreshTokenID(emptyUUID)
		assert.ErrorIs(t, err, ErrInvalidUUID)
		assert.Equal(t, RefreshTokenID{}, refreshTokenID)
	})
}

func TestNewRefreshToken(t *testing.T) {
	id, err := NewRefreshTokenID(uuid.New().String())
	assert.NoError(t, err)
	userID, err := NewUserID(uuid.New().String())
	assert.NoError(t, err)
	token := "some-refresh-token"
	expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	refreshToken := NewRefreshToken(id, userID, token, expiresAt, createdAt)

	assert.Equal(t, id, refreshToken.ID)
	assert.Equal(t, userID, refreshToken.UserID)
	assert.Equal(t, token, refreshToken.Token)
	assert.True(t, refreshToken.ExpiresAt.Equal(expiresAt))
	assert.True(t, refreshToken.CreatedAt.Equal(createdAt))
}
