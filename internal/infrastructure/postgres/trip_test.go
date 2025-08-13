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

// testTrip テスト用のTrip構造体
type testTrip struct {
	ID        domain.TripID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// newTestTrip テスト用のTripを生成する
func newTestTrip(name string) testTrip {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return testTrip{
		ID:        domain.NewTripID(uuid.New().String()),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// toDomainTrip ドメインオブジェクトに変換する
func (tt testTrip) toDomainTrip() *domain.Trip {
	return domain.NewTrip(tt.ID, tt.Name, tt.CreatedAt, tt.UpdatedAt)
}

// tripTestSuite テスト用の共通セットアップ
type tripTestSuite struct {
	ctx     context.Context
	tx      pgx.Tx
	repo    domain.TripRepository
	queries *postgres.Queries
	mapper  *mapper.PostgreSQLTypeMapper
}

// newTripTestSuite テストスイートを作成する（トランザクション分離）
func newTripTestSuite(t *testing.T) *tripTestSuite {
	t.Helper()

	ctx := context.Background()
	db := setupDB(t, ctx)

	// サブテスト用のトランザクションを開始
	tx, err := db.Begin(ctx)
	require.NoError(t, err, "トランザクション開始に失敗")

	// サブテスト終了時にロールバック
	t.Cleanup(func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Logf("トランザクションロールバック時の警告: %v", err)
		}
	})

	return &tripTestSuite{
		ctx:     ctx,
		tx:      tx,
		repo:    NewTripPostgresRepository(tx), // トランザクションを渡す
		queries: postgres.New(tx),
		mapper:  mapper.NewPostgreSQLTypeMapper(),
	}
}

// createTripInDB データベースに直接Tripを作成する
func (s *tripTestSuite) createTripInDB(t *testing.T, trip testTrip) {
	t.Helper()

	pgUUID, err := s.mapper.ToUUID(trip.ID.String())
	require.NoError(t, err, "UUID変換に失敗")
	pgCreatedAt, err := s.mapper.ToTimestamp(trip.CreatedAt)
	require.NoError(t, err, "CreatedAt変換に失敗")
	pgUpdatedAt, err := s.mapper.ToTimestamp(trip.UpdatedAt)
	require.NoError(t, err, "UpdatedAt変換に失敗")

	err = s.queries.CreateTrip(s.ctx, postgres.CreateTripParams{
		ID:        pgUUID,
		Name:      trip.Name,
		CreatedAt: pgCreatedAt,
		UpdatedAt: pgUpdatedAt,
	})
	require.NoError(t, err, "テストデータの作成に失敗")
}

// getTripFromDB データベースから直接Tripを取得する
func (s *tripTestSuite) getTripFromDB(t *testing.T, id domain.TripID) (*postgres.Trip, error) {
	t.Helper()

	pgUUID, err := s.mapper.ToUUID(id.String())
	require.NoError(t, err, "UUID変換に失敗")
	trip, err := s.queries.FindTrip(s.ctx, pgUUID)

	return &trip, err
}

// assertTripEquals Tripの等価性をアサートする
func assertTripEquals(t *testing.T, expected testTrip, actual *domain.Trip) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID(), "TripIDが一致すること")
	assert.Equal(t, expected.Name, actual.Name(), "TripNameが一致すること")
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt(), time.Second,
		"CreatedAtがほぼ一致すること (expected: %v, actual: %v)", expected.CreatedAt, actual.CreatedAt())
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt(), time.Second,
		"UpdatedAtがほぼ一致すること (expected: %v, actual: %v)", expected.UpdatedAt, actual.UpdatedAt())
}

// assertTripExistsInDB データベースにTripが存在することをアサートする
func (s *tripTestSuite) assertTripExistsInDB(t *testing.T, expected testTrip) {
	t.Helper()

	record, err := s.getTripFromDB(t, expected.ID)
	require.NoError(t, err, "データベースにTripが存在すること")

	actualID, err := s.mapper.FromUUID(record.ID)
	require.NoError(t, err, "UUID変換に失敗")
	assert.Equal(t, expected.ID.String(), actualID, "IDが一致すること")
	assert.Equal(t, expected.Name, record.Name, "Nameが一致すること")

	actualCreatedAt, err := s.mapper.FromTimestamp(record.CreatedAt)
	require.NoError(t, err, "CreatedAt変換に失敗")
	assert.WithinDuration(t, expected.CreatedAt, actualCreatedAt, time.Second, "CreatedAtがほぼ一致すること")

	actualUpdatedAt, err := s.mapper.FromTimestamp(record.UpdatedAt)
	require.NoError(t, err, "UpdatedAt変換に失敗")
	assert.WithinDuration(t, expected.UpdatedAt, actualUpdatedAt, time.Second, "UpdatedAtがほぼ一致すること")
}

// assertTripNotExistsInDB データベースにTripが存在しないことをアサートする
func (s *tripTestSuite) assertTripNotExistsInDB(t *testing.T, id domain.TripID) {
	t.Helper()

	_, err := s.getTripFromDB(t, id)
	assert.ErrorIs(t, err, pgx.ErrNoRows,
		"データベースにTripが存在しないこと")
}

// assertTripsContainAll 取得したTripリストに期待するTripがすべて含まれることをアサートする
func assertTripsContainAll(t *testing.T, expected []testTrip, actual []*domain.Trip) {
	t.Helper()

	require.Len(t, actual, len(expected), "取得したTrip数が期待値と一致すること")

	expectedIDs := make(map[domain.TripID]testTrip)
	for _, trip := range expected {
		expectedIDs[trip.ID] = trip
	}

	for _, actualTrip := range actual {
		expectedTrip, exists := expectedIDs[actualTrip.ID()]
		require.True(t, exists, "期待されるTripが含まれること (ID: %s)", actualTrip.ID())

		assert.Equal(t, expectedTrip.Name, actualTrip.Name(), "Tripの名前が一致すること")
		assert.WithinDuration(t, expectedTrip.CreatedAt, actualTrip.CreatedAt(), time.Second, "CreatedAtがほぼ一致すること")
		assert.WithinDuration(t, expectedTrip.UpdatedAt, actualTrip.UpdatedAt(), time.Second, "UpdatedAtがほぼ一致すること")
	}
}

func TestTripPostgresRepository_NewTripPostgresRepository(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t, ctx)

	repo := NewTripPostgresRepository(db)
	assert.NotNil(t, repo, "リポジトリインスタンスがnilであってはならない")
}

func TestTripPostgresRepository_FindByID(t *testing.T) {
	t.Run("存在するIDでTripを取得できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: データベースにTripが存在する
		testTrip := newTestTrip("テスト旅行")
		suite.createTripInDB(t, testTrip)

		// When: FindByIDでTripを取得する
		foundTrip, err := suite.repo.FindByID(suite.ctx, testTrip.ID)

		// Then: Tripが正常に取得できる
		require.NoError(t, err, "FindByIDでエラーが発生してはならない")
		require.NotNil(t, foundTrip, "取得したTripがnilであってはならない")
		assertTripEquals(t, testTrip, foundTrip)
	})

	t.Run("存在しないIDでErrTripNotFoundが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 存在しないID
		nonExistentID := domain.NewTripID(uuid.New().String())

		// When: 存在しないIDでTripを取得する
		_, err := suite.repo.FindByID(suite.ctx, nonExistentID)

		// Then: ErrTripNotFoundが返される
		assert.ErrorIs(t, err, apperr.ErrTripNotFound,
			"ErrTripNotFoundが返されるべき")
	})

	t.Run("不正な形式のIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 不正な形式のID
		invalidID := domain.NewTripID("invalid-uuid-format")

		// When: 不正なIDでTripを取得する
		_, err := suite.repo.FindByID(suite.ctx, invalidID)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("空文字列のIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 空文字列のID
		emptyID := domain.NewTripID("")

		// When: 空のIDでTripを取得する
		_, err := suite.repo.FindByID(suite.ctx, emptyID)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})
}

func TestTripPostgresRepository_FindMany(t *testing.T) {
	t.Run("Tripが存在しない場合空のリストが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: データベースが空（他のテストの影響を受けない）

		// When: FindManyでTripを取得する
		foundTrips, err := suite.repo.FindMany(suite.ctx)

		// Then: 空のリストが返される
		require.NoError(t, err, "FindManyでエラーが発生してはならない")
		assert.Empty(t, foundTrips, "空のリストが返されるべき")
	})

	t.Run("複数のTripが存在する場合すべて取得できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: データベースに複数のTripが存在する
		trip1 := newTestTrip("北海道旅行")
		trip2 := newTestTrip("沖縄旅行")
		trip3 := newTestTrip("京都旅行")
		expectedTrips := []testTrip{trip1, trip2, trip3}

		suite.createTripInDB(t, trip1)
		suite.createTripInDB(t, trip2)
		suite.createTripInDB(t, trip3)

		// When: FindManyでTripを取得する
		foundTrips, err := suite.repo.FindMany(suite.ctx)

		// Then: すべてのTripが取得される
		require.NoError(t, err, "FindManyでエラーが発生してはならない")
		assertTripsContainAll(t, expectedTrips, foundTrips)
	})

	t.Run("単一のTripが存在する場合正常に取得できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: データベースに1つのTripが存在する（トランザクション分離により他のテストの影響なし）
		trip := newTestTrip("単一旅行")
		suite.createTripInDB(t, trip)
		expectedTrips := []testTrip{trip}

		// When: FindManyでTripを取得する
		foundTrips, err := suite.repo.FindMany(suite.ctx)

		// Then: 1つのTripが取得される
		require.NoError(t, err, "FindManyでエラーが発生してはならない")
		assertTripsContainAll(t, expectedTrips, foundTrips)
	})
}

func TestTripPostgresRepository_Create(t *testing.T) {
	t.Run("新しいTripを正常に作成できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 新しいTrip
		testTrip := newTestTrip("新規旅行")
		domainTrip := testTrip.toDomainTrip()

		// When: Tripを作成する
		err := suite.repo.Create(suite.ctx, domainTrip)

		// Then: Tripが正常に作成される
		require.NoError(t, err, "Createでエラーが発生してはならない")
		suite.assertTripExistsInDB(t, testTrip)
	})

	t.Run("nilのTripでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// When: nilのTripを作成する
		err := suite.repo.Create(suite.ctx, nil)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("重複するIDでエラーが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 既存のTrip
		existingTrip := newTestTrip("既存旅行")
		suite.createTripInDB(t, existingTrip)

		// When: 同じIDで別のTripを作成する
		duplicateTrip := testTrip{
			ID:        existingTrip.ID, // 同じID
			Name:      "重複旅行",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err := suite.repo.Create(suite.ctx, duplicateTrip.toDomainTrip())

		// Then: エラーが返される
		assert.Error(t, err, "重複IDでの作成はエラーになるべき")
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("不正なIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 不正な形式のIDを持つTrip
		invalidTrip := testTrip{
			ID:        domain.NewTripID("invalid-uuid-format"),
			Name:      "不正ID旅行",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		// When: 不正なIDでTripを作成する
		err := suite.repo.Create(suite.ctx, invalidTrip.toDomainTrip())

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("空文字列の名前でTripを作成できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 空文字列の名前を持つTrip
		emptyNameTrip := newTestTrip("")

		// When: 空文字列名でTripを作成する
		err := suite.repo.Create(suite.ctx, emptyNameTrip.toDomainTrip())

		// Then: 正常に作成される（ビジネスロジックでの検証は別途実装）
		require.NoError(t, err, "空文字列名でも作成できるべき")
		suite.assertTripExistsInDB(t, emptyNameTrip)
	})

	t.Run("複数のTripを連続作成できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 複数の新しいTrip
		trips := []testTrip{
			newTestTrip("連続作成1"),
			newTestTrip("連続作成2"),
			newTestTrip("連続作成3"),
		}

		// When: 複数のTripを連続で作成する
		for _, trip := range trips {
			err := suite.repo.Create(suite.ctx, trip.toDomainTrip())
			require.NoError(t, err, "連続作成でエラーが発生してはならない")
		}

		// Then: すべてのTripが正常に作成される
		for _, trip := range trips {
			suite.assertTripExistsInDB(t, trip)
		}
	})
}

func TestTripPostgresRepository_Update(t *testing.T) {
	t.Run("既存のTripを正常に更新できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 既存のTrip
		originalTrip := newTestTrip("更新前旅行")
		suite.createTripInDB(t, originalTrip)

		// When: Tripを更新する
		updatedTime := time.Now().UTC().Truncate(time.Microsecond)
		updatedTrip := testTrip{
			ID:        originalTrip.ID,
			Name:      "更新後旅行",
			CreatedAt: originalTrip.CreatedAt, // CreatedAtは変わらない
			UpdatedAt: updatedTime,
		}
		err := suite.repo.Update(suite.ctx, updatedTrip.toDomainTrip())

		// Then: Tripが正常に更新される
		require.NoError(t, err, "Updateでエラーが発生してはならない")
		suite.assertTripExistsInDB(t, updatedTrip)
	})

	t.Run("nilのTripでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// When: nilのTripを更新する
		err := suite.repo.Update(suite.ctx, nil)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("存在しないIDでもエラーにならないこと", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 存在しないTrip
		nonExistentTrip := newTestTrip("存在しない旅行")

		// When: 存在しないTripを更新する
		err := suite.repo.Update(suite.ctx, nonExistentTrip.toDomainTrip())

		// Then: エラーにならない（PostgreSQLのUPDATEは影響行数0でもエラーにならない）
		assert.NoError(t, err, "存在しないIDの更新はエラーにならないべき")
	})

	t.Run("不正なIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 不正な形式のIDを持つTrip
		invalidTrip := testTrip{
			ID:        domain.NewTripID("invalid-uuid-format"),
			Name:      "不正ID旅行",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		// When: 不正なIDでTripを更新する
		err := suite.repo.Update(suite.ctx, invalidTrip.toDomainTrip())

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("名前のみを更新できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 既存のTrip
		originalTrip := newTestTrip("名前変更前")
		suite.createTripInDB(t, originalTrip)

		updatedTime := time.Now().UTC().Truncate(time.Microsecond)

		// When: 名前のみを更新する
		updatedTrip := testTrip{
			ID:        originalTrip.ID,
			Name:      "名前変更後",
			CreatedAt: originalTrip.CreatedAt,
			UpdatedAt: updatedTime,
		}
		err := suite.repo.Update(suite.ctx, updatedTrip.toDomainTrip())

		// Then: 名前が更新されCreatedAtは変わらない
		require.NoError(t, err, "名前の更新でエラーが発生してはならない")
		suite.assertTripExistsInDB(t, updatedTrip)
	})
}

func TestTripPostgresRepository_Delete(t *testing.T) {
	t.Run("存在するIDのTripを正常に削除できること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 既存のTrip
		existingTrip := newTestTrip("削除対象旅行")
		suite.createTripInDB(t, existingTrip)

		// When: Tripを削除する
		err := suite.repo.Delete(suite.ctx, existingTrip.ID)

		// Then: Tripが正常に削除される
		require.NoError(t, err, "Deleteでエラーが発生してはならない")
		suite.assertTripNotExistsInDB(t, existingTrip.ID)
	})

	t.Run("存在しないIDでErrTripNotFoundが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 存在しないID
		nonExistentID := domain.NewTripID(uuid.New().String())

		// When: 存在しないIDでTripを削除する
		err := suite.repo.Delete(suite.ctx, nonExistentID)

		// Then: ErrTripNotFoundが返される
		assert.ErrorIs(t, err, apperr.ErrTripNotFound,
			"ErrTripNotFoundが返されるべき")
	})

	t.Run("不正な形式のIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 不正な形式のID
		invalidID := domain.NewTripID("invalid-uuid-format")

		// When: 不正なIDでTripを削除する
		err := suite.repo.Delete(suite.ctx, invalidID)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""),
			"InternalErrorが返されるべき")
	})

	t.Run("空文字列のIDでInternalErrorが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 空文字列のID
		emptyID := domain.NewTripID("")

		// When: 空のIDでTripを削除する
		err := suite.repo.Delete(suite.ctx, emptyID)

		// Then: InternalErrorが返される
		assert.ErrorIs(t, err, apperr.NewInternalError(""), "InternalErrorが返されるべき")
	})

	t.Run("複数のTripが存在する場合指定したもののみ削除されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 複数のTripが存在する
		trip1 := newTestTrip("旅行1")
		trip2 := newTestTrip("旅行2")
		trip3 := newTestTrip("旅行3")

		suite.createTripInDB(t, trip1)
		suite.createTripInDB(t, trip2)
		suite.createTripInDB(t, trip3)

		// When: 1つのTripを削除する
		err := suite.repo.Delete(suite.ctx, trip2.ID)

		// Then: 指定したTripのみが削除される
		require.NoError(t, err, "Deleteでエラーが発生してはならない")
		suite.assertTripExistsInDB(t, trip1)
		suite.assertTripNotExistsInDB(t, trip2.ID)
		suite.assertTripExistsInDB(t, trip3)
	})

	t.Run("同じIDを複数回削除した場合2回目はErrTripNotFoundが返されること", func(t *testing.T) {
		suite := newTripTestSuite(t)

		// Given: 既存のTrip
		existingTrip := newTestTrip("重複削除対象")
		suite.createTripInDB(t, existingTrip)

		// When: 同じIDを2回削除する
		err1 := suite.repo.Delete(suite.ctx, existingTrip.ID)
		err2 := suite.repo.Delete(suite.ctx, existingTrip.ID)

		// Then: 1回目は正常に削除される
		assert.NoError(t, err1, "1回目の削除でエラーが発生してはならない")

		// 2回目はErrTripNotFoundが返される
		assert.ErrorIs(t, err2, apperr.ErrTripNotFound, "2回目はErrTripNotFoundが返されるべき")

		suite.assertTripNotExistsInDB(t, existingTrip.ID)
	})
}
