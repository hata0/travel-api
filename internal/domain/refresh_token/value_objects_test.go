package refreshtoken

import (
	"testing"

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
