package postgres

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testContainer *postgres.PostgresContainer
	testDbUrl     string
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	ctx := context.Background()

	container, dbUrl, err := setupTestContainer(ctx)
	if err != nil {
		log.Printf("error setting up test container: %v", err)
		return 1
	}
	testContainer = container
	testDbUrl = dbUrl

	code := m.Run()

	// コンテナ終了
	if err := testContainer.Terminate(ctx); err != nil {
		log.Printf("error terminating test container: %v", err)
	}

	return code
}

func setupTestContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
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
	if err != nil {
		return nil, "", err
	}

	// 接続文字列を取得
	dbUrl, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", err
	}

	// マイグレーションを実行
	mig, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		return nil, "", err
	}
	if err := mig.Up(); err != nil {
		return nil, "", err
	}
	sourceErr, dbErr := mig.Close()
	if sourceErr != nil {
		return nil, "", sourceErr
	}
	if dbErr != nil {
		return nil, "", dbErr
	}

	// スナップショットを作成
	if err := container.Snapshot(ctx, postgres.WithSnapshotName("test-db-snapshot")); err != nil {
		return nil, "", err
	}

	return container, dbUrl, nil
}

func setupDB(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()

	db, err := pgx.Connect(ctx, testDbUrl)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close(ctx)
		require.NoError(t, err)
		// 各テストの終了時にスナップショットを復元
		err = testContainer.Restore(ctx)
		require.NoError(t, err)
	})

	return db
}
