package revokedtoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRevokedTokenID_NewRevokedTokenID(t *testing.T) {
	idValue := "test-revoked-token-id-123"
	revokedTokenID := NewRevokedTokenID(idValue)
	assert.Equal(t, idValue, revokedTokenID.value, "NewRevokedTokenID は正しい値を持つ RevokedTokenID を生成するべき")
}

func TestRevokedTokenID_String(t *testing.T) {
	idValue := "test-revoked-token-id-456"
	revokedTokenID := NewRevokedTokenID(idValue)
	assert.Equal(t, idValue, revokedTokenID.String(), "String() は正しい ID 値を返すべき")
}

func TestRevokedTokenID_Equals(t *testing.T) {
	id1 := NewRevokedTokenID("id-1")
	id2 := NewRevokedTokenID("id-1")
	id3 := NewRevokedTokenID("id-2")

	assert.True(t, id1.Equals(id2), "同じ値を持つ 2 つの RevokedTokenID は等しいと判定されるべき")
	assert.False(t, id1.Equals(id3), "異なる値を持つ 2 つの RevokedTokenID は等しくないと判定されるべき")
}
