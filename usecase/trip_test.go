package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
	"travel-api/domain"
	mock_domain "travel-api/domain/mock"
	"travel-api/usecase/output"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTripInteractor_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripID := domain.TripID("test-id")
	now := time.Now()
	expectedTrip := domain.NewTrip(tripID, "Test Trip", now, now)

	mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(expectedTrip, nil).Times(1)

	tripOutput, err := interactor.Get(context.Background(), string(tripID))

	assert.NoError(t, err)
	assert.Equal(t, output.NewGetTripOutput(expectedTrip), tripOutput)
}

func TestTripInteractor_Get_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripID := domain.TripID("test-id")
	expectedErr := errors.New("not found")

	mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(domain.Trip{}, expectedErr).Times(1)

	_, err := interactor.Get(context.Background(), string(tripID))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTripInteractor_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	now := time.Now()
	expectedTrips := []domain.Trip{
		domain.NewTrip("1", "Trip 1", now, now),
		domain.NewTrip("2", "Trip 2", now, now),
	}

	mockRepo.EXPECT().FindMany(gomock.Any()).Return(expectedTrips, nil).Times(1)

	tripsOutput, err := interactor.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, output.NewListTripOutput(expectedTrips), tripsOutput)
}

func TestTripInteractor_List_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	expectedErr := errors.New("db error")

	mockRepo.EXPECT().FindMany(gomock.Any()).Return(nil, expectedErr).Times(1)

	_, err := interactor.List(context.Background())

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTripInteractor_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripName := "New Trip"

	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := interactor.Create(context.Background(), tripName)

	assert.NoError(t, err)
}

func TestTripInteractor_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripID := domain.TripID("test-id")
	tripName := "Original Trip"
	updatedTripName := "Updated Trip"
	now := time.Now()
	originalTrip := domain.NewTrip(tripID, tripName, now, now)

	mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(originalTrip, nil).Times(1)
	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err := interactor.Update(context.Background(), string(tripID), updatedTripName)

	assert.NoError(t, err)
}

func TestTripInteractor_Update_FindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripID := domain.TripID("test-id")
	updatedTripName := "Updated Trip"
	expectedErr := errors.New("not found")

	mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(domain.Trip{}, expectedErr).Times(1)

	err := interactor.Update(context.Background(), string(tripID), updatedTripName)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTripInteractor_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripID := domain.TripID("test-id")
	now := time.Now()
	tripToDelete := domain.NewTrip(tripID, "Trip to delete", now, now)

	mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(tripToDelete, nil).Times(1)
	mockRepo.EXPECT().Delete(gomock.Any(), tripToDelete).Return(nil).Times(1)

	err := interactor.Delete(context.Background(), string(tripID))

	assert.NoError(t, err)
}

func TestTripInteractor_Delete_FindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockTripRepository(ctrl)
	interactor := NewTripInteractor(mockRepo)

	tripID := domain.TripID("test-id")
	expectedErr := errors.New("not found")

	mockRepo.EXPECT().FindByID(gomock.Any(), tripID).Return(domain.Trip{}, expectedErr).Times(1)

	err := interactor.Delete(context.Background(), string(tripID))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}