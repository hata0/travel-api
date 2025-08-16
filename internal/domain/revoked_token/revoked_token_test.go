package revokedtoken

import (
	"testing"
	"time"

	"github.com/hata0/travel-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestNewRevokedToken(t *testing.T) {
	id := NewRevokedTokenID("revoked-token-id-1")
	userID := user.NewUserID("user-id-1")
	tokenJTI := "jti-123"
	expiresAt := time.Now().Add(1 * time.Hour)
	revokedAt := time.Now()

	revokedToken := NewRevokedToken(id, userID, tokenJTI, expiresAt, revokedAt)

	assert.NotNil(t, revokedToken, "NewRevokedToken は nil を返すべきではない")
	assert.Equal(t, id, revokedToken.id, "NewRevokedToken は正しい ID を設定するべき")
	assert.Equal(t, userID, revokedToken.userID, "NewRevokedToken は正しい UserID を設定するべき")
	assert.Equal(t, tokenJTI, revokedToken.tokenJTI, "NewRevokedToken は正しい TokenJTI を設定するべき")
	assert.Equal(t, expiresAt, revokedToken.expiresAt, "NewRevokedToken は正しい ExpiresAt を設定するべき")
	assert.Equal(t, revokedAt, revokedToken.revokedAt, "NewRevokedToken は正しい RevokedAt を設定するべき")
}

func TestRevokedToken_Getters(t *testing.T) {
	id := NewRevokedTokenID("revoked-token-id-2")
	userID := user.NewUserID("user-id-2")
	tokenJTI := "jti-456"
	expiresAt := time.Now().Add(2 * time.Hour)
	revokedAt := time.Now().Add(1 * time.Hour)

	revokedToken := NewRevokedToken(id, userID, tokenJTI, expiresAt, revokedAt)

	assert.Equal(t, id, revokedToken.ID(), "ID() は正しい ID を返すべき")
	assert.Equal(t, userID, revokedToken.UserID(), "UserID() は正しい UserID を返すべき")
	assert.Equal(t, tokenJTI, revokedToken.TokenJTI(), "TokenJTI() は正しい TokenJTI を返すべき")
	assert.Equal(t, expiresAt, revokedToken.ExpiresAt(), "ExpiresAt() は正しい ExpiresAt を返すべき")
	assert.Equal(t, revokedAt, revokedToken.RevokedAt(), "RevokedAt() は正しい RevokedAt を返すべき")
}

func TestRevokedToken_Equals(t *testing.T) {
	id1 := NewRevokedTokenID("revoked-token-id-3")
	id2 := NewRevokedTokenID("revoked-token-id-4")
	userID := user.NewUserID("user-id-3")
	now := time.Now()

	revokedToken1 := NewRevokedToken(id1, userID, "jti-A", now.Add(1*time.Hour), now)
	revokedToken2 := NewRevokedToken(id1, userID, "jti-A", now.Add(1*time.Hour), now)                  // revokedToken1 と同じ ID
	revokedToken3 := NewRevokedToken(id2, userID, "jti-B", now.Add(2*time.Hour), now.Add(1*time.Hour)) // revokedToken1 と異なる ID

	assert.True(t, revokedToken1.Equals(revokedToken2), "同じ ID を持つ 2 つの RevokedToken は等しいと判定されるべき")
	assert.False(t, revokedToken1.Equals(revokedToken3), "異なる ID を持つ 2 つの RevokedToken は等しくないと判定されるべき")
	assert.False(t, revokedToken1.Equals(nil), "RevokedToken は nil と等しいと判定されるべきではない")
}
