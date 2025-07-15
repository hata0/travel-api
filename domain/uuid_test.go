package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUUID(t *testing.T) {
	uuid := NewUUID()
	assert.True(t, IsValidUUID(uuid))
}

func TestIsValidUUID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		uuid := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
		assert.True(t, IsValidUUID(uuid))
	})

	t.Run("invalid UUID", func(t *testing.T) {
		uuid := "not-a-valid-uuid"
		assert.False(t, IsValidUUID(uuid))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.False(t, IsValidUUID(""))
	})

	t.Run("non-v4 UUID", func(t *testing.T) {
		// Example of a v1 UUID
		uuid := "6fa459ea-ee8a-11e7-80c1-9a214cf093ae"
		assert.False(t, IsValidUUID(uuid))
	})
}
