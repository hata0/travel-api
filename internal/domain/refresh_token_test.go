package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenID_NewRefreshTokenID(t *testing.T) {
	idValue := "test-refresh-token-id-123"
	refreshTokenID := NewRefreshTokenID(idValue)
	assert.Equal(t, idValue, refreshTokenID.value, "NewRefreshTokenID は正しい値を持つ RefreshTokenID を生成するべき")
}

func TestRefreshTokenID_String(t *testing.T) {
	idValue := "test-refresh-token-id-456"
	refreshTokenID := NewRefreshTokenID(idValue)
	assert.Equal(t, idValue, refreshTokenID.String(), "String() は正しい ID 値を返すべき")
}

func TestRefreshTokenID_Equals(t *testing.T) {
	id1 := NewRefreshTokenID("id-1")
	id2 := NewRefreshTokenID("id-1")
	id3 := NewRefreshTokenID("id-2")

	assert.True(t, id1.Equals(id2), "同じ値を持つ 2 つの RefreshTokenID は等しいと判定されるべき")
	assert.False(t, id1.Equals(id3), "異なる値を持つ 2 つの RefreshTokenID は等しくないと判定されるべき")
}

func TestNewRefreshToken(t *testing.T) {
	id := NewRefreshTokenID("refresh-token-id-1")
	userID := NewUserID("user-id-1")
	token := "some-refresh-token-string"
	expiresAt := time.Now().Add(24 * time.Hour)
	createdAt := time.Now()

	refreshToken := NewRefreshToken(id, userID, token, expiresAt, createdAt)

	assert.NotNil(t, refreshToken, "NewRefreshToken は nil を返すべきではない")
	assert.Equal(t, id, refreshToken.id, "NewRefreshToken は正しい ID を設定するべき")
	assert.Equal(t, userID, refreshToken.userID, "NewRefreshToken は正しい UserID を設定するべき")
	assert.Equal(t, token, refreshToken.token, "NewRefreshToken は正しい Token を設定するべき")
	assert.Equal(t, expiresAt, refreshToken.expiresAt, "NewRefreshToken は正しい ExpiresAt を設定するべき")
	assert.Equal(t, createdAt, refreshToken.createdAt, "NewRefreshToken は正しい CreatedAt を設定するべき")
}

func TestRefreshToken_Getters(t *testing.T) {
	id := NewRefreshTokenID("refresh-token-id-2")
	userID := NewUserID("user-id-2")
	token := "another-refresh-token-string"
	expiresAt := time.Now().Add(48 * time.Hour)
	createdAt := time.Now().Add(24 * time.Hour)

	refreshToken := NewRefreshToken(id, userID, token, expiresAt, createdAt)

	assert.Equal(t, id, refreshToken.ID(), "ID() は正しい ID を返すべき")
	assert.Equal(t, userID, refreshToken.UserID(), "UserID() は正しい UserID を返すべき")
	assert.Equal(t, token, refreshToken.Token(), "Token() は正しい Token を返すべき")
	assert.Equal(t, expiresAt, refreshToken.ExpiresAt(), "ExpiresAt() は正しい ExpiresAt を返すべき")
	assert.Equal(t, createdAt, refreshToken.CreatedAt(), "CreatedAt() は正しい CreatedAt を返すべき")
}

func TestRefreshToken_Equals(t *testing.T) {
	id1 := NewRefreshTokenID("refresh-token-id-3")
	id2 := NewRefreshTokenID("refresh-token-id-4")
	userID := NewUserID("user-id-3")
	now := time.Now()

	refreshToken1 := NewRefreshToken(id1, userID, "token-A", now.Add(1*time.Hour), now)
	refreshToken2 := NewRefreshToken(id1, userID, "token-A", now.Add(1*time.Hour), now) // refreshToken1 と同じ ID
	refreshToken3 := NewRefreshToken(id2, userID, "token-B", now.Add(2*time.Hour), now.Add(1*time.Hour)) // refreshToken1 と異なる ID

	assert.True(t, refreshToken1.Equals(refreshToken2), "同じ ID を持つ 2 つの RefreshToken は等しいと判定されるべき")
	assert.False(t, refreshToken1.Equals(refreshToken3), "異なる ID を持つ 2 つの RefreshToken は等しくないと判定されるべき")
	assert.False(t, refreshToken1.Equals(nil), "RefreshToken は nil と等しいと判定されるべきではない")
}
