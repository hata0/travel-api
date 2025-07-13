package database

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestQueries(t *testing.T) {
	ctx := context.Background()

	ctr, dbURL := setupDB(t, ctx)
	queries := setupQueries(t, ctx, ctr, dbURL)

	t.Run("CreateTrip", func(t *testing.T) {
		var id pgtype.UUID
		err := id.Scan("11112222-3333-4444-5555-666677778888")
		require.NoError(t, err)

		var now pgtype.Timestamptz
		err = now.Scan(time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC))
		require.NoError(t, err)

		err = queries.CreateTrip(context.Background(), CreateTripParams{
			ID:        id,
			Name:      "name abc",
			CreatedAt: now,
			UpdatedAt: now,
		})
		require.NoError(t, err)
	})
}
