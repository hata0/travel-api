package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/hata0/travel-api/internal/infrastructure/postgres/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTripPostgresRepository_NewTripPostgresRepository はNewTripPostgresRepository関数のテスト
func TestTripPostgresRepository_NewTripPostgresRepository(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx) // テスト用DB接続の取得

	// 新しいリポジトリインスタンスがnilではないことを確認
	repo := NewTripPostgresRepository(db)
	assert.NotNil(t, repo, "リポジトリインスタンスがnilであってはならない")
}

// TestTripPostgresRepository_FindByID はFindByIDメソッドのテスト
func TestTripPostgresRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx) // テスト用DB接続の取得
	repo := NewTripPostgresRepository(db)
	queries := postgres.New(db)
	mapper := mapper.NewPostgreSQLTypeMapper()

	// テストデータ準備
	now := time.Now().UTC().Truncate(time.Microsecond)
	testTripID := domain.NewTripID(uuid.New().String())
	testTripName := "テスト旅行"

	pgUUID, err := mapper.ToUUID(testTripID.String())
	require.NoError(t, err)
	pgCreatedAt, err := mapper.ToTimestamp(now)
	require.NoError(t, err)
	pgUpdatedAt, err := mapper.ToTimestamp(now)
	require.NoError(t, err)

	// データベースに直接テストデータを挿入
	err = queries.CreateTrip(ctx, postgres.CreateTripParams{
		ID:        pgUUID,
		Name:      testTripName,
		CreatedAt: pgCreatedAt,
		UpdatedAt: pgUpdatedAt,
	})
	require.NoError(t, err, "テストデータの挿入に失敗")

	t.Run("存在するIDでTripを取得できること", func(t *testing.T) {
		foundTrip, err := repo.FindByID(ctx, testTripID)
		require.NoError(t, err, "FindByIDでエラーが発生してはならない")
		assert.NotNil(t, foundTrip, "取得したTripがnilであってはならない")
		assert.Equal(t, testTripID, foundTrip.ID(), "TripIDが一致すること")
		assert.Equal(t, testTripName, foundTrip.Name(), "TripNameが一致すること")
		assert.WithinDuration(t, now, foundTrip.CreatedAt(), time.Second, "CreatedAtがほぼ一致すること")
		assert.WithinDuration(t, now, foundTrip.UpdatedAt(), time.Second, "UpdatedAtがほぼ一致すること")
	})

	t.Run("存在しないIDでTripを取得しようとするとErrTripNotFoundが返されること", func(t *testing.T) {
		nonExistentID := domain.NewTripID(uuid.New().String())
		_, err := repo.FindByID(ctx, nonExistentID)
		assert.True(t, errors.Is(err, apperr.ErrTripNotFound), "ErrTripNotFoundが返されるべき")
	})

	t.Run("不正な形式のIDでTripを取得しようとするとInternalErrorが返されること", func(t *testing.T) {
		invalidID := domain.NewTripID("invalid-uuid")
		_, err := repo.FindByID(ctx, invalidID)
		assert.True(t, errors.Is(err, apperr.NewInternalError("", nil)), "InternalErrorが返されるべき")
	})
}

// TestTripPostgresRepository_FindMany はFindManyメソッドのテスト
func TestTripPostgresRepository_FindMany(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx) // テスト用DB接続の取得
	repo := NewTripPostgresRepository(db)
	queries := postgres.New(db)
	mapper := mapper.NewPostgreSQLTypeMapper()

	t.Run("Tripが一つも存在しない場合、空のリストが返されること", func(t *testing.T) {
		foundTrips, err := repo.FindMany(ctx)
		require.NoError(t, err, "FindManyでエラーが発生してはならない")
		assert.Len(t, foundTrips, 0, "空のリストが返されるべき")
	})

	t.Run("複数のTripが存在する場合、すべて取得できること", func(t *testing.T) {
		// テストデータ準備
		now := time.Now().UTC().Truncate(time.Microsecond)
		trip1ID := domain.NewTripID(uuid.New().String())
		trip2ID := domain.NewTripID(uuid.New().String())

		pgUUID1, err := mapper.ToUUID(trip1ID.String())
		require.NoError(t, err)
		pgUUID2, err := mapper.ToUUID(trip2ID.String())
		require.NoError(t, err)
		pgCreatedAt, err := mapper.ToTimestamp(now)
		require.NoError(t, err)
		pgUpdatedAt, err := mapper.ToTimestamp(now)
		require.NoError(t, err)

		// データベースに直接テストデータを挿入
		err = queries.CreateTrip(ctx, postgres.CreateTripParams{
			ID:        pgUUID1,
			Name:      "旅行1",
			CreatedAt: pgCreatedAt,
			UpdatedAt: pgUpdatedAt,
		})
		require.NoError(t, err)
		err = queries.CreateTrip(ctx, postgres.CreateTripParams{
			ID:        pgUUID2,
			Name:      "旅行2",
			CreatedAt: pgCreatedAt,
			UpdatedAt: pgUpdatedAt,
		})
		require.NoError(t, err)

		foundTrips, err := repo.FindMany(ctx)
		require.NoError(t, err, "FindManyでエラーが発生してはならない")
		assert.Len(t, foundTrips, 2, "2つのTripが取得されるべき")

		// 取得したTripのIDをセットに変換して比較
		foundIDs := make(map[domain.TripID]bool)
		for _, trip := range foundTrips {
			foundIDs[trip.ID()] = true
		}
		assert.True(t, foundIDs[trip1ID], "Trip1が含まれるべき")
		assert.True(t, foundIDs[trip2ID], "Trip2が含まれるべき")
	})
}

// TestTripPostgresRepository_Create はCreateメソッドのテスト
func TestTripPostgresRepository_Create(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx) // テスト用DB接続の取得
	repo := NewTripPostgresRepository(db)
	queries := postgres.New(db)
	mapper := mapper.NewPostgreSQLTypeMapper()

	t.Run("新しいTripを正常に作成できること", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		newTripID := domain.NewTripID(uuid.New().String())
		newTripName := "新規作成旅行"
		newTrip := domain.NewTrip(newTripID, newTripName, now, now)

		err := repo.Create(ctx, newTrip)
		require.NoError(t, err, "Createでエラーが発生してはならない")

		// データベースから直接取得して検証
		pgUUID, err := mapper.ToUUID(newTripID.String())
		require.NoError(t, err)
		record, err := queries.GetTrip(ctx, pgUUID)
		require.NoError(t, err, "作成したTripがDBに存在しない")

		actualID, err := mapper.FromUUID(record.ID)
		require.NoError(t, err)
		assert.Equal(t, newTripID.String(), actualID, "IDが一致すること")
		assert.Equal(t, newTripName, record.Name, "Nameが一致すること")
		actualCreatedAt, err := mapper.FromTimestamp(record.CreatedAt)
		require.NoError(t, err)
		assert.WithinDuration(t, now, actualCreatedAt, time.Second, "CreatedAtがほぼ一致すること")
		actualUpdatedAt, err := mapper.FromTimestamp(record.UpdatedAt)
		require.NoError(t, err)
		assert.WithinDuration(t, now, actualUpdatedAt, time.Second, "UpdatedAtがほぼ一致すること")
	})

	t.Run("nilのTripを渡すとInternalErrorが返されること", func(t *testing.T) {
		err := repo.Create(ctx, nil)
		assert.True(t, errors.Is(err, apperr.NewInternalError("", nil)), "InternalErrorが返されるべき")
	})

	t.Run("重複するIDでTripを作成しようとするとエラーが返されること", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		duplicateTripID := domain.NewTripID(uuid.New().String())
		trip1 := domain.NewTrip(duplicateTripID, "最初の旅行", now, now)
		trip2 := domain.NewTrip(duplicateTripID, "重複旅行", now, now)

		err := repo.Create(ctx, trip1)
		require.NoError(t, err)

		err = repo.Create(ctx, trip2)
		assert.Error(t, err, "重複IDでの作成はエラーになるべき")
		assert.True(t, errors.Is(err, apperr.NewInternalError("", nil)), "InternalErrorが返されるべき") // PostgreSQLの重複キーエラーがInternalErrorにラップされる
	})
}

// TestTripPostgresRepository_Update はUpdateメソッドのテスト
func TestTripPostgresRepository_Update(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx) // テスト用DB接続の取得
	repo := NewTripPostgresRepository(db)
	queries := postgres.New(db)
	mapper := mapper.NewPostgreSQLTypeMapper()

	// テストデータ準備
	originalNow := time.Now().UTC().Truncate(time.Microsecond)
	updateTripID := domain.NewTripID(uuid.New().String())
	originalTripName := "更新前旅行"
	originalTrip := domain.NewTrip(updateTripID, originalTripName, originalNow, originalNow)

	err := repo.Create(ctx, originalTrip)
	require.NoError(t, err, "テストデータの作成に失敗")

	t.Run("既存のTripを正常に更新できること", func(t *testing.T) {
		updatedName := "更新後旅行"
		updatedNow := time.Now().UTC().Truncate(time.Microsecond)
		updatedTrip := domain.NewTrip(updateTripID, updatedName, originalNow, updatedNow) // CreatedAtは変わらない

		err := repo.Update(ctx, updatedTrip)
		require.NoError(t, err, "Updateでエラーが発生してはならない")

		// データベースから直接取得して検証
		pgUUID, err := mapper.ToUUID(updateTripID.String())
		require.NoError(t, err)
		record, err := queries.GetTrip(ctx, pgUUID)
		require.NoError(t, err, "更新したTripがDBに存在しない")

		assert.Equal(t, updatedName, record.Name, "Nameが更新されていること")
		actualCreatedAt, err := mapper.FromTimestamp(record.CreatedAt)
		require.NoError(t, err)
		assert.WithinDuration(t, originalNow, actualCreatedAt, time.Second, "CreatedAtは変わらないこと")
		actualUpdatedAt, err := mapper.FromTimestamp(record.UpdatedAt)
		require.NoError(t, err)
		assert.WithinDuration(t, updatedNow, actualUpdatedAt, time.Second, "UpdatedAtが更新されていること")
	})

	t.Run("nilのTripを渡すとInternalErrorが返されること", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.True(t, errors.Is(err, apperr.NewInternalError("", nil)), "InternalErrorが返されるべき")
	})
}

// TestTripPostgresRepository_Delete はDeleteメソッドのテスト
func TestTripPostgresRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx) // テスト用DB接続の取得
	repo := NewTripPostgresRepository(db)
	queries := postgres.New(db)
	mapper := mapper.NewPostgreSQLTypeMapper()

	// テストデータ準備
	now := time.Now().UTC().Truncate(time.Microsecond)
	deleteTripID := domain.NewTripID(uuid.New().String())
	deleteTripName := "削除対象旅行"
	deleteTrip := domain.NewTrip(deleteTripID, deleteTripName, now, now)

	err := repo.Create(ctx, deleteTrip)
	require.NoError(t, err, "テストデータの作成に失敗")

	t.Run("存在するIDのTripを正常に削除できること", func(t *testing.T) {
		err := repo.Delete(ctx, deleteTripID)
		require.NoError(t, err, "Deleteでエラーが発生してはならない")

		// データベースから直接取得して検証
		pgUUID, err := mapper.ToUUID(deleteTripID.String())
		require.NoError(t, err)
		_, err = queries.GetTrip(ctx, pgUUID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows), "TripがDBから削除されているべき")
	})

	t.Run("存在しないIDのTripを削除しようとしてもエラーにならないこと", func(t *testing.T) {
		nonExistentID := domain.NewTripID(uuid.New().String())
		err := repo.Delete(ctx, nonExistentID)
		assert.NoError(t, err, "存在しないIDの削除はエラーにならないべき")
	})

	t.Run("不正な形式のIDでTripを削除しようとするとInternalErrorが返されること", func(t *testing.T) {
		invalidID := domain.NewTripID("invalid-uuid")
		err := repo.Delete(ctx, invalidID)
		assert.True(t, errors.Is(err, apperr.NewInternalError("", nil)), "InternalErrorが返されるべき")
	})
}
