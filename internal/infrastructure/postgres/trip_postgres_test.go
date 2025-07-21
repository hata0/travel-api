package postgres

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/app_error"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestTrip はテスト用のTripドメインオブジェクトを生成するヘルパー関数
func createTestTrip(t *testing.T, name string, createdAt, updatedAt time.Time) domain.Trip {
	t.Helper()
	id, err := domain.NewTripID(uuid.New().String()) // 動的にUUIDを生成
	require.NoError(t, err)
	return domain.NewTrip(id, name, createdAt, updatedAt)
}

// insertTestTrip はテスト用のTripドメインオブジェクトをデータベースに挿入するヘルパー関数
func insertTestTrip(t *testing.T, ctx context.Context, db DBTX, trip domain.Trip) {
	t.Helper()
	queries := New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(trip.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	err := queries.CreateTrip(ctx, CreateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		CreatedAt: validatedCreatedAt,
		UpdatedAt: validatedUpdatedAt,
	})
	require.NoError(t, err)
}

// getTripFromDB はデータベースから直接Tripレコードを取得するヘルパー関数
func getTripFromDB(t *testing.T, ctx context.Context, db DBTX, idStr string) (Trip, error) {
	t.Helper()
	queries := New(db)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(idStr)

	return queries.GetTrip(ctx, validatedId)
}

func TestTripPostgresRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewTripPostgresRepository(dbConn)

	t.Run("正常系: レコードが存在する", func(t *testing.T) {
		name := "Trip to Tokyo"
		createdAt := time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC)
		updatedAt := time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC)
		trip := createTestTrip(t, name, createdAt, updatedAt)
		insertTestTrip(t, ctx, dbConn, trip)

		foundTrip, err := repo.FindByID(ctx, trip.ID)

		assert.NoError(t, err)
		assert.Equal(t, trip.ID.String(), foundTrip.ID.String())
		assert.Equal(t, trip.Name, foundTrip.Name)
		assert.True(t, trip.CreatedAt.Equal(foundTrip.CreatedAt))
		assert.True(t, trip.UpdatedAt.Equal(foundTrip.UpdatedAt))
	})

	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		id, err := domain.NewTripID(uuid.New().String())
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, id)

		assert.ErrorIs(t, err, app_error.ErrTripNotFound)
	})
}

func TestTripPostgresRepository_FindMany(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewTripPostgresRepository(dbConn)

	t.Run("正常系: 複数のレコードが存在する", func(t *testing.T) {
		trip1 := createTestTrip(t, "Trip A", time.Now().Truncate(time.Microsecond), time.Now().Truncate(time.Microsecond))
		trip2 := createTestTrip(t, "Trip B", time.Now().Truncate(time.Microsecond), time.Now().Truncate(time.Microsecond))
		insertTestTrip(t, ctx, dbConn, trip1)
		insertTestTrip(t, ctx, dbConn, trip2)

		trips, err := repo.FindMany(ctx)

		assert.NoError(t, err)
		assert.Len(t, trips, 2)

		// 取得したTripスライスをIDでマップに変換して検証
		tripMap := make(map[string]domain.Trip)
		for _, t := range trips {
			tripMap[t.ID.String()] = t
		}

		assert.Equal(t, trip1.ID.String(), tripMap[trip1.ID.String()].ID.String())
		assert.Equal(t, trip1.Name, tripMap[trip1.ID.String()].Name)
		assert.True(t, trip1.CreatedAt.Equal(tripMap[trip1.ID.String()].CreatedAt))
		assert.True(t, trip1.UpdatedAt.Equal(tripMap[trip1.ID.String()].UpdatedAt))

		assert.Equal(t, trip2.ID.String(), tripMap[trip2.ID.String()].ID.String())
		assert.Equal(t, trip2.Name, tripMap[trip2.ID.String()].Name)
		assert.True(t, trip2.CreatedAt.Equal(tripMap[trip2.ID.String()].CreatedAt))
		assert.True(t, trip2.UpdatedAt.Equal(tripMap[trip2.ID.String()].UpdatedAt))
	})

	t.Run("正常系: レコードが存在しない", func(t *testing.T) {
		// データベースをクリーンアップ
		_, err := dbConn.Exec(ctx, "DELETE FROM trips")
		require.NoError(t, err)

		trips, err := repo.FindMany(ctx)

		assert.NoError(t, err)
		assert.Empty(t, trips)
	})
}

func TestTripPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewTripPostgresRepository(dbConn)

	t.Run("正常系: 新しいレコードが作成される", func(t *testing.T) {
		name := "New Trip"
		now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		trip := createTestTrip(t, name, now, now)

		err := repo.Create(ctx, trip)
		assert.NoError(t, err)

		// DBから直接取得して検証
		createdRecord, err := getTripFromDB(t, ctx, dbConn, trip.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, trip.ID.String(), createdRecord.ID.String())
		assert.Equal(t, trip.Name, createdRecord.Name)
		assert.True(t, trip.CreatedAt.Equal(createdRecord.CreatedAt.Time))
		assert.True(t, trip.UpdatedAt.Equal(createdRecord.UpdatedAt.Time))
	})

	t.Run("異常系: 重複するIDで作成", func(t *testing.T) {
		name := "Existing Trip"
		now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		trip := createTestTrip(t, name, now, now)
		insertTestTrip(t, ctx, dbConn, trip) // 最初に挿入

		err := repo.Create(ctx, trip) // 同じIDで再度挿入
		assert.ErrorIs(t, err, app_error.ErrTripAlreadyExists)
	})
}

func TestTripPostgresRepository_Update(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewTripPostgresRepository(dbConn)

	originalName := "Original Trip"
	originalCreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	originalUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	originalTrip := createTestTrip(t, originalName, originalCreatedAt, originalUpdatedAt)
	insertTestTrip(t, ctx, dbConn, originalTrip)

	t.Run("正常系: レコードが更新される", func(t *testing.T) {
		updatedName := "Updated Trip Name"
		updatedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		// 更新対象のTripオブジェクトは、IDと更新したいフィールドを持つ
		updateTrip := createTestTrip(t, updatedName, originalCreatedAt, updatedAt)
		updateTrip.ID = originalTrip.ID // 既存のIDを設定

		err := repo.Update(ctx, updateTrip)
		assert.NoError(t, err)

		// DBから直接取得して検証
		foundRecord, err := getTripFromDB(t, ctx, dbConn, originalTrip.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, updatedName, foundRecord.Name)
		assert.True(t, updatedAt.Equal(foundRecord.UpdatedAt.Time))
		assert.True(t, originalCreatedAt.Equal(foundRecord.CreatedAt.Time)) // CreatedAtは変わらないことを確認
	})
}

func TestTripPostgresRepository_Delete(t *testing.T) {
	ctx := context.Background()
	dbConn := setupDB(t, ctx)
	repo := NewTripPostgresRepository(dbConn)

	name := "Trip to Delete"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	tripToDelete := createTestTrip(t, name, now, now)
	insertTestTrip(t, ctx, dbConn, tripToDelete)

	t.Run("正常系: レコードが削除される", func(t *testing.T) {
		err := repo.Delete(ctx, tripToDelete)
		assert.NoError(t, err)

		// DBから直接取得して削除されたことを検証
		_, err = getTripFromDB(t, ctx, dbConn, tripToDelete.ID.String())
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}
