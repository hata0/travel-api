package application

import (
	"context"
	"errors"
	"testing"
	"time"
	"travel-api/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTripRepository is a mock implementation of the TripRepository interface for testing.
type MockTripRepository struct {
	mock.Mock
}

func (m *MockTripRepository) FindByID(ctx context.Context, id domain.TripID) (domain.Trip, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Trip), args.Error(1)
}

func (m *MockTripRepository) FindMany(ctx context.Context) ([]domain.Trip, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Trip), args.Error(1)
}

func (m *MockTripRepository) Create(ctx context.Context, trip domain.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockTripRepository) Update(ctx context.Context, trip domain.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockTripRepository) Delete(ctx context.Context, trip domain.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func TestNewTripService(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repository)
}

func TestTripServiceImpl_Get(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripID := domain.TripID("test-id")
	expectedTrip := domain.NewTrip(tripID, "Test Trip", time.Now(), time.Now())

	mockRepo.On("FindByID", mock.Anything, tripID).Return(expectedTrip, nil)

	trip, err := service.Get(context.Background(), string(tripID))

	assert.NoError(t, err)
	assert.Equal(t, expectedTrip, trip)
	mockRepo.AssertExpectations(t)
}

func TestTripServiceImpl_Get_Error(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripID := domain.TripID("test-id")
	expectedErr := errors.New("not found")

	mockRepo.On("FindByID", mock.Anything, tripID).Return(domain.Trip{}, expectedErr)

	_, err := service.Get(context.Background(), string(tripID))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestTripServiceImpl_List(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	expectedTrips := []domain.Trip{
		domain.NewTrip("1", "Trip 1", time.Now(), time.Now()),
		domain.NewTrip("2", "Trip 2", time.Now(), time.Now()),
	}

	mockRepo.On("FindMany", mock.Anything).Return(expectedTrips, nil)

	trips, err := service.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedTrips, trips)
	mockRepo.AssertExpectations(t)
}

func TestTripServiceImpl_List_Error(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	expectedErr := errors.New("db error")

	mockRepo.On("FindMany", mock.Anything).Return(nil, expectedErr)

	_, err := service.List(context.Background())

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestTripServiceImpl_Create(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripName := "New Trip"

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("domain.Trip")).Return(nil)

	err := service.Create(context.Background(), tripName)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTripServiceImpl_Update(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripID := domain.TripID("test-id")
	tripName := "Original Trip"
	updatedTripName := "Updated Trip"
	now := time.Now()
	originalTrip := domain.NewTrip(tripID, tripName, now, now)

	mockRepo.On("FindByID", mock.Anything, tripID).Return(originalTrip, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("domain.Trip")).Return(nil)

	err := service.Update(context.Background(), string(tripID), updatedTripName)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Check if the updated name is correct in the trip passed to Update
	mockRepo.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(trip domain.Trip) bool {
		return trip.Name == updatedTripName && trip.ID == tripID
	}))
}

func TestTripServiceImpl_Update_FindError(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripID := domain.TripID("test-id")
	updatedTripName := "Updated Trip"
	expectedErr := errors.New("not found")

	mockRepo.On("FindByID", mock.Anything, tripID).Return(domain.Trip{}, expectedErr)

	err := service.Update(context.Background(), string(tripID), updatedTripName)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestTripServiceImpl_Delete(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripID := domain.TripID("test-id")
	now := time.Now()
	tripToDelete := domain.NewTrip(tripID, "Trip to delete", now, now)

	mockRepo.On("FindByID", mock.Anything, tripID).Return(tripToDelete, nil)
	mockRepo.On("Delete", mock.Anything, tripToDelete).Return(nil)

	err := service.Delete(context.Background(), string(tripID))

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTripServiceImpl_Delete_FindError(t *testing.T) {
	mockRepo := new(MockTripRepository)
	service := NewTripService(mockRepo)

	tripID := domain.TripID("test-id")
	expectedErr := errors.New("not found")

	mockRepo.On("FindByID", mock.Anything, tripID).Return(domain.Trip{}, expectedErr)

	err := service.Delete(context.Background(), string(tripID))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
