package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func beforeAll(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string) {
	// PostgreSQL コンテナを起動
	container, err := postgres.Run(
		ctx,
		"postgres:17.4-bookworm",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	// 接続文字列を取得
	dbUrl, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// マイグレーションを実行
	m, err := migrate.New(
		"file://sql/migrations",
		dbUrl,
	)
	require.NoError(t, err)
	err = m.Up()
	require.NoError(t, err)
	sourceErr, dbErr := m.Close()
	require.NoError(t, sourceErr)
	require.NoError(t, dbErr)

	// テストデータを挿入
	db, err := sql.Open("pgx", dbUrl)
	require.NoError(t, err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures"),
	)
	require.NoError(t, err)
	err = fixtures.Load()
	require.NoError(t, err)

	db.Close()

	// スナップショットを作成
	err = container.Snapshot(ctx, postgres.WithSnapshotName("test-db-snapshot"))
	require.NoError(t, err)

	return container, dbUrl
}

func setup(t *testing.T, ctx context.Context, container *postgres.PostgresContainer, dbUrl string) *Queries {
	t.Helper()

	connection, err := pgx.Connect(ctx, dbUrl)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := connection.Close(ctx)
		require.NoError(t, err)
		err = container.Restore(ctx)
		require.NoError(t, err)
	})

	queries := New(connection)

	return queries
}
