package repository

import (
	"context"
	"errors"
	"testing"
	"time"
	"travel-api/domain"
	"travel-api/infrastructure/database"
	mock_repository "travel-api/infrastructure/repository/mock"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTripPostgresRepository_FindByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock_repository.NewMockTripQuerier(ctrl)
	repo := NewTripPostgresRepository(mockQueries)

	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	var pgUUID pgtype.UUID
	err := pgUUID.Scan(string(tripID))
	require.NoError(t, err)

	t.Run("正常系", func(t *testing.T) {
		name := "Test Trip"
		now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
		mockTrip := database.Trip{
			ID:        pgUUID,
			Name:      name,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}
		mockQueries.EXPECT().GetTrip(gomock.Any(), pgUUID).Return(mockTrip, nil).Times(1)

		trip, err := repo.FindByID(context.Background(), tripID)

		assert.NoError(t, err)
		assert.Equal(t, tripID, trip.ID)
		assert.Equal(t, name, trip.Name)
	})

	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		mockQueries.EXPECT().GetTrip(gomock.Any(), pgUUID).Return(database.Trip{}, pgx.ErrNoRows).Times(1)

		_, err := repo.FindByID(context.Background(), tripID)

		assert.ErrorIs(t, err, domain.ErrTripNotFound)
	})

	t.Run("異常系: 不明なエラー", func(t *testing.T) {
		expectedErr := errors.New("some error")
		mockQueries.EXPECT().GetTrip(gomock.Any(), pgUUID).Return(database.Trip{}, expectedErr).Times(1)

		_, err := repo.FindByID(context.Background(), tripID)

		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTripPostgresRepository_FindMany(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock_repository.NewMockTripQuerier(ctrl)
	repo := NewTripPostgresRepository(mockQueries)

	// mockされたtrip recordを作成
	var pgUUID pgtype.UUID
	err := pgUUID.Scan("00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)
	name := "Test Trip 1"
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	mockTrips := []database.Trip{
		{
			ID:        pgUUID,
			Name:      name,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockQueries.EXPECT().ListTrips(gomock.Any()).Return(mockTrips, nil).Times(1)

	trips, err := repo.FindMany(context.Background())

	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, name, trips[0].Name)
}

func TestTripPostgresRepository_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock_repository.NewMockTripQuerier(ctrl)
	repo := NewTripPostgresRepository(mockQueries)

	// ドメインオブジェクトを作成
	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(tripID, "New Trip", now, now)

	mockQueries.EXPECT().CreateTrip(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := repo.Create(context.Background(), trip)

	assert.NoError(t, err)
}

func TestTripPostgresRepository_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock_repository.NewMockTripQuerier(ctrl)
	repo := NewTripPostgresRepository(mockQueries)

	// ドメインオブジェクトを作成
	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(tripID, "Updated Trip", now, now)

	mockQueries.EXPECT().UpdateTrip(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := repo.Update(context.Background(), trip)

	assert.NoError(t, err)
}

func TestTripPostgresRepository_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mock_repository.NewMockTripQuerier(ctrl)
	repo := NewTripPostgresRepository(mockQueries)

	// ドメインオブジェクトを作成
	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(tripID, "Test Trip", now, now)

	var pgUUID pgtype.UUID
	err := pgUUID.Scan(string(tripID))
	require.NoError(t, err)

	mockQueries.EXPECT().DeleteTrip(gomock.Any(), pgUUID).Return(nil).Times(1)

	err = repo.Delete(context.Background(), trip)

	assert.NoError(t, err)
}
