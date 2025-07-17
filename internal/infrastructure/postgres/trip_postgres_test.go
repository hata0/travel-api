package postgres

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTripPostgresRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := NewTripPostgresRepository(setupDB(t, ctx))

	id := domain.TripID("00000000-0000-0000-0000-000000000001")

	t.Run("正常系", func(t *testing.T) {
		trip, err := repo.FindByID(context.Background(), id)

		assert.NoError(t, err)
		assert.Equal(t, id, trip.ID)
		assert.Equal(t, "Trip to Tokyo", trip.Name)
	})

	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		id := domain.TripID("99999999-9999-9999-9999-999999999999")

		_, err := repo.FindByID(context.Background(), id)

		assert.ErrorIs(t, err, domain.ErrTripNotFound)
	})
}

func TestTripPostgresRepository_FindMany(t *testing.T) {
	ctx := context.Background()
	repo := NewTripPostgresRepository(setupDB(t, ctx))

	trips, err := repo.FindMany(context.Background())

	assert.NoError(t, err)
	assert.Len(t, trips, 2)
}

func TestTripPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewTripPostgresRepository(setupDB(t, ctx))

	// ドメインオブジェクトを作成
	id := domain.TripID("11112222-3333-4444-5555-666677778888")
	name := "name abc"
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(id, name, now, now)

	err := repo.Create(context.Background(), trip)
	require.NoError(t, err)

	trip, err = repo.FindByID(ctx, id)
	assert.NoError(t, err)

	assert.Equal(t, id, trip.ID)
	assert.Equal(t, name, trip.Name)
	assert.Equal(t, now, trip.CreatedAt)
	assert.Equal(t, now, trip.UpdatedAt)
}

func TestTripPostgresRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewTripPostgresRepository(setupDB(t, ctx))

	// ドメインオブジェクトを作成
	id := domain.TripID("00000000-0000-0000-0000-000000000001")
	name := "Updated Trip"
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(id, name, now, now)

	err := repo.Update(context.Background(), trip)
	require.NoError(t, err)

	trip, err = repo.FindByID(ctx, id)
	assert.NoError(t, err)

	assert.Equal(t, name, trip.Name)
	assert.Equal(t, now, trip.UpdatedAt)
}

func TestTripPostgresRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewTripPostgresRepository(setupDB(t, ctx))

	// ドメインオブジェクトを作成
	id := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(id, "Test Trip", now, now)

	err := repo.Delete(context.Background(), trip)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, id)
	require.Error(t, err)
}
