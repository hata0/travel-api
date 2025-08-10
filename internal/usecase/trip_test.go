package usecase

import (
	"context"
	"testing"
	"time"
	"travel-api/internal/domain"
	apperr "travel-api/internal/domain/errors"
	mock_domain "travel-api/internal/domain/mock"
	mock_clock "travel-api/internal/domain/shared/clock/mock"
	mock_uuid "travel-api/internal/domain/shared/uuid/mock"
	"travel-api/internal/usecase/output"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTripInteractor_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_clock.NewMockClock(ctrl)
	mockUUIDGenerator := mock_uuid.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	t.Run("正常系: Tripが取得できる", func(t *testing.T) {
		tripID, err := domain.NewTripID("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
		assert.NoError(t, err)
		now := time.Now()
		expectedTrip := domain.NewTrip(tripID, "Test Trip", now, now)

		mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(expectedTrip, nil).Times(1)

		tripOutput, err := interactor.Get(context.Background(), tripID.String())

		assert.NoError(t, err)
		assert.Equal(t, output.NewGetTripOutput(expectedTrip), tripOutput)
	})

	t.Run("異常系: 無効なUUIDの場合", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		_, err := interactor.Get(context.Background(), invalidUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
	})
}

func TestTripInteractor_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_clock.NewMockClock(ctrl)
	mockUUIDGenerator := mock_uuid.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	t.Run("正常系: 複数のTripが取得できる", func(t *testing.T) {
		now := time.Now()
		tripID1, err := domain.NewTripID("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
		assert.NoError(t, err)
		tripID2, err := domain.NewTripID("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
		assert.NoError(t, err)
		expectedTrips := []domain.Trip{
			domain.NewTrip(tripID1, "Trip 1", now, now),
			domain.NewTrip(tripID2, "Trip 2", now, now),
		}

		mockRepo.EXPECT().FindMany(gomock.Any()).Return(expectedTrips, nil).Times(1)

		tripsOutput, err := interactor.List(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, output.NewListTripOutput(expectedTrips), tripsOutput)
	})
}

func TestTripInteractor_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_clock.NewMockClock(ctrl)
	mockUUIDGenerator := mock_uuid.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	tripName := "New Trip"
	generatedUUID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	mockUUIDGenerator.EXPECT().NewUUID().Return(generatedUUID).Times(1)
	mockClock.EXPECT().Now().Return(now).Times(1)

	tripID, err := domain.NewTripID(generatedUUID)
	assert.NoError(t, err)
	expectedTrip := domain.NewTrip(tripID, tripName, now, now)

	mockRepo.EXPECT().Create(gomock.Any(), expectedTrip).Return(nil).Times(1)

	createdID, err := interactor.Create(context.Background(), tripName)

	assert.NoError(t, err)
	assert.Equal(t, generatedUUID, createdID)
}

func TestTripInteractor_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_clock.NewMockClock(ctrl)
	mockUUIDGenerator := mock_uuid.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	t.Run("正常系: Tripが更新できる", func(t *testing.T) {
		tripID, err := domain.NewTripID("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
		assert.NoError(t, err)
		tripName := "Original Trip"
		updatedTripName := "Updated Trip"
		originalCreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		originalUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		newUpdatedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		originalTrip := domain.NewTrip(tripID, tripName, originalCreatedAt, originalUpdatedAt)
		updatedTrip := originalTrip.Update(updatedTripName, newUpdatedAt)

		mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(originalTrip, nil).Times(1)
		mockClock.EXPECT().Now().Return(newUpdatedAt).Times(1)
		mockRepo.EXPECT().Update(gomock.Any(), updatedTrip).Return(nil).Times(1)

		err = interactor.Update(context.Background(), tripID.String(), updatedTripName)

		assert.NoError(t, err)
	})

	t.Run("異常系: 無効なUUIDの場合", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		err := interactor.Update(context.Background(), invalidUUID, "any name")
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
	})
}

func TestTripInteractor_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	mockClock := mock_clock.NewMockClock(ctrl)
	mockUUIDGenerator := mock_uuid.NewMockUUIDGenerator(ctrl)
	interactor := NewTripInteractor(mockRepo, mockClock, mockUUIDGenerator)

	t.Run("正常系: Tripが削除できる", func(t *testing.T) {
		tripID, err := domain.NewTripID("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")
		assert.NoError(t, err)
		now := time.Now()
		tripToDelete := domain.NewTrip(tripID, "Trip to delete", now, now)

		mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(tripToDelete, nil).Times(1)
		mockRepo.EXPECT().Delete(gomock.Any(), tripToDelete).Return(nil).Times(1)

		err = interactor.Delete(context.Background(), tripID.String())

		assert.NoError(t, err)
	})

	t.Run("異常系: 無効なUUIDの場合", func(t *testing.T) {
		invalidUUID := "invalid-uuid"
		err := interactor.Delete(context.Background(), invalidUUID)
		assert.ErrorIs(t, err, apperr.ErrInvalidUUID)
	})
}
