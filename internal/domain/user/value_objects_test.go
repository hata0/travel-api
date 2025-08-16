package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserID_NewUserID(t *testing.T) {
	idValue := "test-user-id-123"
	userID := NewUserID(idValue)
	assert.Equal(t, idValue, userID.value, "NewUserID は正しい値を持つ UserID を生成するべき")
}

func TestUserID_String(t *testing.T) {
	idValue := "test-user-id-456"
	userID := NewUserID(idValue)
	assert.Equal(t, idValue, userID.String(), "String() は正しい ID 値を返すべき")
}

func TestUserID_Equals(t *testing.T) {
	id1 := NewUserID("user-id-1")
	id2 := NewUserID("user-id-1")
	id3 := NewUserID("user-id-2")

	assert.True(t, id1.Equals(id2), "同じ値を持つ 2 つの UserID は等しいと判定されるべき")
	assert.False(t, id1.Equals(id3), "異なる値を持つ 2 つの UserID は等しくないと判定されるべき")
}
