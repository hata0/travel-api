package database

import (
	"context"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupDB(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string) {
	ctr, err := postgres.Run(
		ctx,
		"postgres:17.4-bookworm",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	dbURL, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	m, err := migrate.New(
		"file://sql/migrations",
		dbURL,
	)
	require.NoError(t, err)
	err = m.Up()
	require.NoError(t, err)
	sourceErr, dbErr := m.Close()
	require.NoError(t, sourceErr)
	require.NoError(t, dbErr)

	err = ctr.Snapshot(ctx, postgres.WithSnapshotName("test-db-snapshot"))
	require.NoError(t, err)

	return ctr, dbURL
}

func setupQueries(t *testing.T, ctx context.Context, ctr *postgres.PostgresContainer, dbURL string) *Queries {
	conn, err := pgx.Connect(context.Background(), dbURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		conn.Close(context.Background())
		err := ctr.Restore(ctx)
		require.NoError(t, err)
	})

	queries := New(conn)

	return queries
}
