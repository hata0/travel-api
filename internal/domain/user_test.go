package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewUserID(t *testing.T) {
	t.Run("正常系: 有効なUUID", func(t *testing.T) {
		validUUID := uuid.New().String()
		userID, err := NewUserID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, UserID{value: validUUID}, userID)
	})

	t.Run("異常系: 無効なUUID", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		userID, err := NewUserID(invalidUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
		assert.Equal(t, UserID{}, userID)
	})

	t.Run("異常系: 空文字列", func(t *testing.T) {
		emptyUUID := ""
		userID, err := NewUserID(emptyUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
		assert.Equal(t, UserID{}, userID)
	})
}

func TestNewUser(t *testing.T) {
	id, err := NewUserID(uuid.New().String())
	assert.NoError(t, err)
	username := "testuser"
	email := "test@example.com"
	passwordHash := "hashedpassword"
	now := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)

	user := NewUser(id, username, email, passwordHash, now, now)

	assert.Equal(t, id, user.ID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.True(t, user.CreatedAt.Equal(now))
	assert.True(t, user.UpdatedAt.Equal(now))
}
