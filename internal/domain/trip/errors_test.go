package trip_test

import (
	"testing"

	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/domain/trip"
	"github.com/stretchr/testify/assert"
)

func TestNewTripNotFoundError(t *testing.T) {
	err := trip.NewTripNotFoundError()

	assert.NotNil(t, err)
	assert.IsType(t, &apperr.AppError{}, err)
	assert.Equal(t, trip.CodeTripNotFound, err.Code())
	assert.Equal(t, "Trip not found", err.Message())
}
