package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTrip(t *testing.T) {
	id := TripID("abc123def")
	name := "name abc"
	now := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)

	trip := NewTrip(id, name, now, now)

	assert.Equal(t, id, trip.ID)
	assert.Equal(t, name, trip.Name)
	assert.True(t, trip.CreatedAt.Equal(now))
	assert.True(t, trip.UpdatedAt.Equal(now))
}

func TestTrip_Update(t *testing.T) {
	id := TripID("abc123def")
	name := "name abc"
	past := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)
	trip := NewTrip(id, name, past, past)

	updatedName := "new name abc"
	now := time.Date(2001, time.January, 1, 1, 1, 1, 1, time.UTC)
	updatedTrip := trip.Update(updatedName, now)

	assert.Equal(t, id, updatedTrip.ID)
	assert.Equal(t, updatedName, updatedTrip.Name)
	assert.True(t, updatedTrip.CreatedAt.Equal(past))
	assert.True(t, updatedTrip.UpdatedAt.Equal(now))
}
