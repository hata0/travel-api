package domain

import (
	"testing"
	"time"

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

func TestNewTrip(t *testing.T) {
	id := NewTripID("trip-id-1")
	name := "Test Trip"
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	trip := NewTrip(id, name, createdAt, updatedAt)

	assert.NotNil(t, trip, "NewTrip は nil を返すべきではない")
	assert.Equal(t, id, trip.id, "NewTrip は正しい ID を設定するべき")
	assert.Equal(t, name, trip.name, "NewTrip は正しい name を設定するべき")
	assert.Equal(t, createdAt, trip.createdAt, "NewTrip は正しい createdAt を設定するべき")
	assert.Equal(t, updatedAt, trip.updatedAt, "NewTrip は正しい updatedAt を設定するべき")
}

func TestTrip_Getters(t *testing.T) {
	id := NewTripID("trip-id-2")
	name := "Another Trip"
	createdAt := time.Now().Add(-48 * time.Hour)
	updatedAt := time.Now().Add(-24 * time.Hour)

	trip := NewTrip(id, name, createdAt, updatedAt)

	assert.Equal(t, id, trip.ID(), "ID() は正しい ID を返すべき")
	assert.Equal(t, name, trip.Name(), "Name() は正しい name を返すべき")
	assert.Equal(t, createdAt, trip.CreatedAt(), "CreatedAt() は正しい createdAt を返すべき")
	assert.Equal(t, updatedAt, trip.UpdatedAt(), "UpdatedAt() は正しい updatedAt を返すべき")
}

func TestTrip_Update(t *testing.T) {
	id := NewTripID("trip-id-3")
	originalName := "Original Trip Name"
	originalCreatedAt := time.Now().Add(-72 * time.Hour)
	originalUpdatedAt := time.Now().Add(-48 * time.Hour)

	trip := NewTrip(id, originalName, originalCreatedAt, originalUpdatedAt)

	newName := "Updated Trip Name"
	newUpdatedAt := time.Now()

	updatedTrip := trip.Update(newName, newUpdatedAt)

	assert.NotNil(t, updatedTrip, "Update は新しい Trip インスタンスを返すべき")
	assert.Equal(t, id, updatedTrip.ID(), "Update は元の ID を保持すべき")
	assert.Equal(t, newName, updatedTrip.Name(), "Update は新しい name を設定すべき")
	assert.Equal(t, originalCreatedAt, updatedTrip.CreatedAt(), "Update は元の createdAt を保持すべき")
	assert.Equal(t, newUpdatedAt, updatedTrip.UpdatedAt(), "Update は新しい updatedAt を設定すべき")

	// 元の trip が変更されていないことを確認
	assert.Equal(t, originalName, trip.Name(), "元の Trip の name は変更されてはいけない")
	assert.Equal(t, originalUpdatedAt, trip.UpdatedAt(), "元の Trip の updatedAt は変更されてはいけない")
}

func TestTrip_Equals(t *testing.T) {
	id1 := NewTripID("trip-id-4")
	id2 := NewTripID("trip-id-5")
	now := time.Now()

	trip1 := NewTrip(id1, "Trip A", now, now)
	trip2 := NewTrip(id1, "Trip A", now, now) // trip1 と同じ ID
	trip3 := NewTrip(id2, "Trip B", now, now) // trip1 と異なる ID

	assert.True(t, trip1.Equals(trip2), "同じ ID を持つ 2 つの Trip は等しいと判定されるべき")
	assert.False(t, trip1.Equals(trip3), "異なる ID を持つ 2 つの Trip は等しくないと判定されるべき")
	assert.False(t, trip1.Equals(nil), "Trip は nil と等しいと判定されるべきではない")
}
