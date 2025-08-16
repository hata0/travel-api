package trip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTripID_NewTripID(t *testing.T) {
	idValue := "test-id-123"
	tripID := NewTripID(idValue)
	assert.Equal(t, idValue, tripID.value, "NewTripID は正しい値を持つ TripID を生成するべき")
}

func TestTripID_String(t *testing.T) {
	idValue := "test-id-456"
	tripID := NewTripID(idValue)
	assert.Equal(t, idValue, tripID.String(), "String() は正しい ID 値を返すべき")
}

func TestTripID_Equals(t *testing.T) {
	id1 := NewTripID("id-1")
	id2 := NewTripID("id-1")
	id3 := NewTripID("id-2")

	assert.True(t, id1.Equals(id2), "同じ値を持つ 2 つの TripID は等しいと判定されるべき")
	assert.False(t, id1.Equals(id3), "異なる値を持つ 2 つの TripID は等しくないと判定されるべき")
}
