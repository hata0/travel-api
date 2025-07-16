package database

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueries_GetTrip(t *testing.T) {
	ctx := context.Background()
	queries := setup(t, ctx)

	t.Run("正常系", func(t *testing.T) {
		var id pgtype.UUID
		err := id.Scan("00000000-0000-0000-0000-000000000001")
		require.NoError(t, err)

		trip, err := queries.GetTrip(ctx, id)
		require.NoError(t, err)

		assert.Equal(t, id, trip.ID)
		assert.Equal(t, "Trip to Tokyo", trip.Name)
	})

	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		var id pgtype.UUID
		err := id.Scan("99999999-9999-9999-9999-999999999999")
		require.NoError(t, err)

		_, err = queries.GetTrip(ctx, id)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestQueries_ListTrips(t *testing.T) {
	ctx := context.Background()

	queries := setup(t, ctx)

	trips, err := queries.ListTrips(ctx)
	require.NoError(t, err)

	assert.Len(t, trips, 2)
}

func TestQueries_CreateTrip(t *testing.T) {
	ctx := context.Background()

	queries := setup(t, ctx)

	var id pgtype.UUID
	err := id.Scan("11112222-3333-4444-5555-666677778888")
	require.NoError(t, err)

	name := "name abc"

	// TIMESTAMPTZ型のため、ミリ秒以下は無視される
	var now pgtype.Timestamptz
	err = now.Scan(time.Date(2000, time.January, 1, 1, 1, 1, 0, time.Local))
	require.NoError(t, err)

	err = queries.CreateTrip(ctx, CreateTripParams{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	require.NoError(t, err)

	trip, err := queries.GetTrip(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, id, trip.ID)
	assert.Equal(t, name, trip.Name)
	assert.Equal(t, now, trip.CreatedAt)
	assert.Equal(t, now, trip.UpdatedAt)
}

func TestQueries_UpdateTrip(t *testing.T) {
	ctx := context.Background()

	queries := setup(t, ctx)

	var id pgtype.UUID
	err := id.Scan("00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	// UpdateTripメソッドはnameを更新しない場合でも、引数に指定しなければ空の値に更新されてしまうので注意
	name := "Trip to Tokyo"

	var now pgtype.Timestamptz
	err = now.Scan(time.Date(2002, time.January, 1, 1, 1, 1, 0, time.Local))
	require.NoError(t, err)

	err = queries.UpdateTrip(ctx, UpdateTripParams{
		ID:        id,
		Name:      name,
		UpdatedAt: now,
	})
	require.NoError(t, err)

	trip, err := queries.GetTrip(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, name, trip.Name)
	assert.Equal(t, now, trip.UpdatedAt)
}

func TestQueries_DeleteTrip(t *testing.T) {
	ctx := context.Background()

	queries := setup(t, ctx)

	var id pgtype.UUID
	err := id.Scan("00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	err = queries.DeleteTrip(ctx, id)
	require.NoError(t, err)

	_, err = queries.GetTrip(ctx, id)
	require.Error(t, err)
}
