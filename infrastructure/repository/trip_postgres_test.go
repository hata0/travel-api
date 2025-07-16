package repository

import (
	"context"
	"errors"
	"testing"
	"time"
	"travel-api/domain"
	"travel-api/infrastructure/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTripQuerier is a mock implementation of the TripQuerier interface for testing.
type MockTripQuerier struct {
	mock.Mock
}

func (m *MockTripQuerier) GetTrip(ctx context.Context, id pgtype.UUID) (database.Trip, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Trip), args.Error(1)
}

func (m *MockTripQuerier) ListTrips(ctx context.Context) ([]database.Trip, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Trip), args.Error(1)
}

func (m *MockTripQuerier) CreateTrip(ctx context.Context, arg database.CreateTripParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockTripQuerier) UpdateTrip(ctx context.Context, arg database.UpdateTripParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockTripQuerier) DeleteTrip(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTripPostgresRepository_FindByID(t *testing.T) {
	mockQueries := new(MockTripQuerier)
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
		mockQueries.On("GetTrip", mock.Anything, pgUUID).Return(mockTrip, nil).Once()

		trip, err := repo.FindByID(context.Background(), tripID)

		assert.NoError(t, err)
		assert.Equal(t, tripID, trip.ID)
		assert.Equal(t, name, trip.Name)
		mockQueries.AssertExpectations(t)
	})

	t.Run("異常系: レコードが存在しない", func(t *testing.T) {
		mockQueries.On("GetTrip", mock.Anything, pgUUID).Return(database.Trip{}, pgx.ErrNoRows).Once()

		_, err := repo.FindByID(context.Background(), tripID)

		assert.ErrorIs(t, err, domain.ErrTripNotFound)
		mockQueries.AssertExpectations(t)
	})

	t.Run("異常系: 不明なエラー", func(t *testing.T) {
		expectedErr := errors.New("some error")
		mockQueries.On("GetTrip", mock.Anything, pgUUID).Return(database.Trip{}, expectedErr).Once()

		_, err := repo.FindByID(context.Background(), tripID)

		assert.ErrorIs(t, err, expectedErr)
		mockQueries.AssertExpectations(t)
	})
}

func TestTripPostgresRepository_FindMany(t *testing.T) {
	mockQueries := new(MockTripQuerier)
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

	mockQueries.On("ListTrips", mock.Anything).Return(mockTrips, nil)

	trips, err := repo.FindMany(context.Background())

	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, name, trips[0].Name)
	mockQueries.AssertExpectations(t)
}

func TestTripPostgresRepository_Create(t *testing.T) {
	mockQueries := new(MockTripQuerier)
	repo := NewTripPostgresRepository(mockQueries)

	// ドメインオブジェクトを作成
	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(tripID, "New Trip", now, now)

	mockQueries.On("CreateTrip", mock.Anything, mock.AnythingOfType("database.CreateTripParams")).Return(nil)

	err := repo.Create(context.Background(), trip)

	assert.NoError(t, err)
	mockQueries.AssertExpectations(t)
}

func TestTripPostgresRepository_Update(t *testing.T) {
	mockQueries := new(MockTripQuerier)
	repo := NewTripPostgresRepository(mockQueries)

	// ドメインオブジェクトを作成
	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(tripID, "Updated Trip", now, now)

	mockQueries.On("UpdateTrip", mock.Anything, mock.AnythingOfType("database.UpdateTripParams")).Return(nil)

	err := repo.Update(context.Background(), trip)

	assert.NoError(t, err)
	mockQueries.AssertExpectations(t)
}

func TestTripPostgresRepository_Delete(t *testing.T) {
	mockQueries := new(MockTripQuerier)
	repo := NewTripPostgresRepository(mockQueries)

	// ドメインオブジェクトを作成
	tripID := domain.TripID("00000000-0000-0000-0000-000000000001")
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	trip := domain.NewTrip(tripID, "Test Trip", now, now)

	var pgUUID pgtype.UUID
	err := pgUUID.Scan(string(tripID))
	require.NoError(t, err)

	mockQueries.On("DeleteTrip", mock.Anything, pgUUID).Return(nil)

	err = repo.Delete(context.Background(), trip)

	assert.NoError(t, err)
	mockQueries.AssertExpectations(t)
}
