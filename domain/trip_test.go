package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTrip(t *testing.T) {
	id := TripId("abc123def")
	name := "name abc"
	now := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)

	trip := NewTrip(id, name, now, now)

	assert.Equal(t, id, trip.id)
	assert.Equal(t, name, trip.name)
	assert.True(t, trip.createdAt.Equal(now))
	assert.True(t, trip.updatedAt.Equal(now))
}

func TestTrip_Update(t *testing.T) {
	id := TripId("abc123def")
	name := "name abc"
	past := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)
	trip := NewTrip(id, name, past, past)

	updatedName := "new name abc"
	now := time.Date(2001, time.January, 1, 1, 1, 1, 1, time.UTC)
	updatedTrip := trip.Update(updatedName, now)

	assert.Equal(t, id, updatedTrip.id)
	assert.Equal(t, updatedName, updatedTrip.name)
	assert.True(t, updatedTrip.createdAt.Equal(past))
	assert.True(t, updatedTrip.updatedAt.Equal(now))
}
