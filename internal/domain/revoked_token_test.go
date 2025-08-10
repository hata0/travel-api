package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRevokedTokenID(t *testing.T) {
	t.Run("正常系: 有効なUUID", func(t *testing.T) {
		validUUID := uuid.New().String()
		revokedTokenID, err := NewRevokedTokenID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, RevokedTokenID{value: validUUID}, revokedTokenID)
	})

	t.Run("異常系: 無効なUUID", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		revokedTokenID, err := NewRevokedTokenID(invalidUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
		assert.Equal(t, RevokedTokenID{}, revokedTokenID)
	})

	t.Run("異常系: 空文字列", func(t *testing.T) {
		emptyUUID := ""
		revokedTokenID, err := NewRevokedTokenID(emptyUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
		assert.Equal(t, RevokedTokenID{}, revokedTokenID)
	})
}

func TestNewRevokedToken(t *testing.T) {
	id, err := NewRevokedTokenID(uuid.New().String())
	require.NoError(t, err)
	userID, err := NewUserID(uuid.New().String())
	assert.NoError(t, err)
	tokenJTI := "some-jti"
	expiresAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	revokedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	revokedToken := NewRevokedToken(id, userID, tokenJTI, expiresAt, revokedAt)

	assert.Equal(t, id, revokedToken.ID)
	assert.Equal(t, userID, revokedToken.UserID)
	assert.Equal(t, tokenJTI, revokedToken.TokenJTI)
	assert.True(t, revokedToken.ExpiresAt.Equal(expiresAt))
	assert.True(t, revokedToken.RevokedAt.Equal(revokedAt))
}
