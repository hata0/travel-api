package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTripID(t *testing.T) {
	t.Run("正常系: 有効なUUID", func(t *testing.T) {
		validUUID := uuid.New().String()
		tripID, err := NewTripID(validUUID)
		assert.NoError(t, err)
		assert.Equal(t, TripID{value: validUUID}, tripID)
	})

	t.Run("異常系: 無効なUUID", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		tripID, err := NewTripID(invalidUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
		assert.Equal(t, TripID{}, tripID)
	})

	t.Run("異常系: 空文字列", func(t *testing.T) {
		emptyUUID := ""
		tripID, err := NewTripID(emptyUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
		assert.Equal(t, TripID{}, tripID)
	})
}

func TestNewTrip(t *testing.T) {
	id, err := NewTripID(uuid.New().String())
	assert.NoError(t, err)
	name := "name abc"
	now := time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)

	trip := NewTrip(id, name, now, now)

	assert.Equal(t, id, trip.ID)
	assert.Equal(t, name, trip.Name)
	assert.True(t, trip.CreatedAt.Equal(now))
	assert.True(t, trip.UpdatedAt.Equal(now))
}

func TestTrip_Update(t *testing.T) {
	id, err := NewTripID(uuid.New().String())
	assert.NoError(t, err)
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
