package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/domain/trip"
	mock_trip "github.com/hata0/travel-api/internal/domain/trip/mock" // repository mock
	"github.com/hata0/travel-api/internal/usecase/output"
	mock_service "github.com/hata0/travel-api/internal/usecase/service/mock" // service mocks
)

func TestTripInteractor_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_trip.NewMockTripRepository(ctrl)
	mockTimeService := mock_service.NewMockTimeService(ctrl)
	mockIDService := mock_service.NewMockIDService(ctrl)

	interactor := NewTripInteractor(mockRepo, mockTimeService, mockIDService)

	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	tripID := trip.NewTripID("test-id")
	testTrip := trip.NewTrip(tripID, "Test Trip", fixedTime, fixedTime)

	tests := []struct {
		name    string
		id      string
		setup   func()
		want    *output.GetTripOutput
		wantErr error
	}{
		{
			name: "正常系: 旅行が正常に取得できる",
			id:   "test-id",
			setup: func() {
				mockRepo.EXPECT().
					FindByID(gomock.Any(), tripID).
					Return(testTrip, nil).
					Times(1)
			},
			want:    output.NewGetTripOutput(testTrip),
			wantErr: nil,
		},
		{
			name: "異常系: リポジトリからアプリケーションエラーが返される",
			id:   "not-found-id",
			setup: func() {
				notFoundID := trip.NewTripID("not-found-id")
				appErr := trip.NewTripNotFoundError()
				mockRepo.EXPECT().
					FindByID(gomock.Any(), notFoundID).
					Return(nil, appErr).
					Times(1)
			},
			want:    nil,
			wantErr: trip.NewTripNotFoundError(),
		},
		{
			name: "異常系: リポジトリから予期しないエラーが返される",
			id:   "error-id",
			setup: func() {
				errorID := trip.NewTripID("error-id")
				unexpectedErr := errors.New("database connection error")
				mockRepo.EXPECT().
					FindByID(gomock.Any(), errorID).
					Return(nil, unexpectedErr).
					Times(1)
			},
			want:    nil,
			wantErr: apperr.NewInternalError("Failed to get trip", apperr.WithCause(errors.New("database connection error"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := interactor.Get(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, got)
				// エラーの型とメッセージを検証
				if appErr, ok := tt.wantErr.(*apperr.AppError); ok {
					gotAppErr, ok := err.(*apperr.AppError)
					require.True(t, ok, "Expected AppError but got %T", err)
					assert.Equal(t, appErr.Code, gotAppErr.Code)
					assert.Equal(t, appErr.Message, gotAppErr.Message)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTripInteractor_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_trip.NewMockTripRepository(ctrl)
	mockTimeService := mock_service.NewMockTimeService(ctrl)
	mockIDService := mock_service.NewMockIDService(ctrl)

	interactor := NewTripInteractor(mockRepo, mockTimeService, mockIDService)

	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	testTrips := []*trip.Trip{
		trip.NewTrip(trip.NewTripID("id1"), "Trip 1", fixedTime, fixedTime),
		trip.NewTrip(trip.NewTripID("id2"), "Trip 2", fixedTime, fixedTime),
	}

	tests := []struct {
		name    string
		setup   func()
		want    *output.ListTripOutput
		wantErr error
	}{
		{
			name: "正常系: 旅行一覧が正常に取得できる",
			setup: func() {
				mockRepo.EXPECT().
					FindMany(gomock.Any()).
					Return(testTrips, nil).
					Times(1)
			},
			want:    output.NewListTripOutput(testTrips),
			wantErr: nil,
		},
		{
			name: "正常系: 空の旅行一覧が取得できる",
			setup: func() {
				mockRepo.EXPECT().
					FindMany(gomock.Any()).
					Return([]*trip.Trip{}, nil).
					Times(1)
			},
			want:    output.NewListTripOutput([]*trip.Trip{}),
			wantErr: nil,
		},
		{
			name: "異常系: リポジトリからアプリケーションエラーが返される",
			setup: func() {
				appErr := apperr.NewInternalError("Database error")
				mockRepo.EXPECT().
					FindMany(gomock.Any()).
					Return(nil, appErr).
					Times(1)
			},
			want:    nil,
			wantErr: apperr.NewInternalError("Database error"),
		},
		{
			name: "異常系: リポジトリから予期しないエラーが返される",
			setup: func() {
				unexpectedErr := errors.New("connection timeout")
				mockRepo.EXPECT().
					FindMany(gomock.Any()).
					Return(nil, unexpectedErr).
					Times(1)
			},
			want:    nil,
			wantErr: apperr.NewInternalError("Failed to list trips", apperr.WithCause(errors.New("connection timeout"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := interactor.List(context.Background())

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, got)
				if appErr, ok := tt.wantErr.(*apperr.AppError); ok {
					gotAppErr, ok := err.(*apperr.AppError)
					require.True(t, ok, "Expected AppError but got %T", err)
					assert.Equal(t, appErr.Code, gotAppErr.Code)
					assert.Equal(t, appErr.Message, gotAppErr.Message)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTripInteractor_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_trip.NewMockTripRepository(ctrl)
	mockTimeService := mock_service.NewMockTimeService(ctrl)
	mockIDService := mock_service.NewMockIDService(ctrl)

	interactor := NewTripInteractor(mockRepo, mockTimeService, mockIDService)

	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	generatedID := "generated-id"

	tests := []struct {
		name     string
		tripName string
		setup    func()
		want     *output.CreateTripOutput
		wantErr  error
	}{
		{
			name:     "正常系: 旅行が正常に作成できる",
			tripName: "New Trip",
			setup: func() {
				mockIDService.EXPECT().
					Generate().
					Return(generatedID).
					Times(1)
				mockTimeService.EXPECT().
					Now().
					Return(fixedTime).
					Times(1)

				expectedTrip := trip.NewTrip(
					trip.NewTripID(generatedID),
					"New Trip",
					fixedTime,
					fixedTime,
				)
				mockRepo.EXPECT().
					Create(gomock.Any(), expectedTrip).
					Return(nil).
					Times(1)
			},
			want:    output.NewCreateTripOutput(trip.NewTripID(generatedID)),
			wantErr: nil,
		},
		{
			name:     "異常系: リポジトリからアプリケーションエラーが返される",
			tripName: "Error Trip",
			setup: func() {
				mockIDService.EXPECT().
					Generate().
					Return(generatedID).
					Times(1)
				mockTimeService.EXPECT().
					Now().
					Return(fixedTime).
					Times(1)

				appErr := apperr.NewInternalError("Database error")
				expectedTrip := trip.NewTrip(
					trip.NewTripID(generatedID),
					"Error Trip",
					fixedTime,
					fixedTime,
				)
				mockRepo.EXPECT().
					Create(gomock.Any(), expectedTrip).
					Return(appErr).
					Times(1)
			},
			want:    nil,
			wantErr: apperr.NewInternalError("Database error"),
		},
		{
			name:     "異常系: リポジトリから予期しないエラーが返される",
			tripName: "Unexpected Error Trip",
			setup: func() {
				mockIDService.EXPECT().
					Generate().
					Return(generatedID).
					Times(1)
				mockTimeService.EXPECT().
					Now().
					Return(fixedTime).
					Times(1)

				unexpectedErr := errors.New("database write error")
				expectedTrip := trip.NewTrip(
					trip.NewTripID(generatedID),
					"Unexpected Error Trip",
					fixedTime,
					fixedTime,
				)
				mockRepo.EXPECT().
					Create(gomock.Any(), expectedTrip).
					Return(unexpectedErr).
					Times(1)
			},
			want:    nil,
			wantErr: apperr.NewInternalError("Failed to create trip", apperr.WithCause(errors.New("database write error"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := interactor.Create(context.Background(), tt.tripName)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, got)
				if appErr, ok := tt.wantErr.(*apperr.AppError); ok {
					gotAppErr, ok := err.(*apperr.AppError)
					require.True(t, ok, "Expected AppError but got %T", err)
					assert.Equal(t, appErr.Code, gotAppErr.Code)
					assert.Equal(t, appErr.Message, gotAppErr.Message)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTripInteractor_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_trip.NewMockTripRepository(ctrl)
	mockTimeService := mock_service.NewMockTimeService(ctrl)
	mockIDService := mock_service.NewMockIDService(ctrl)

	interactor := NewTripInteractor(mockRepo, mockTimeService, mockIDService)

	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updateTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	tripID := trip.NewTripID("test-id")
	originalTrip := trip.NewTrip(tripID, "Original Trip", fixedTime, fixedTime)

	tests := []struct {
		name     string
		id       string
		tripName string
		setup    func()
		wantErr  error
	}{
		{
			name:     "正常系: 旅行が正常に更新できる",
			id:       "test-id",
			tripName: "Updated Trip",
			setup: func() {
				mockTimeService.EXPECT().
					Now().
					Return(updateTime).
					Times(1)

				mockRepo.EXPECT().
					FindByID(gomock.Any(), tripID).
					Return(originalTrip, nil).
					Times(1)

				updatedTrip := originalTrip.Update("Updated Trip", updateTime)
				mockRepo.EXPECT().
					Update(gomock.Any(), updatedTrip).
					Return(nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name:     "異常系: 取得時にアプリケーションエラーが返される",
			id:       "not-found-id",
			tripName: "Updated Trip",
			setup: func() {
				mockTimeService.EXPECT().
					Now().
					Return(updateTime).
					Times(1)

				notFoundID := trip.NewTripID("not-found-id")
				appErr := trip.NewTripNotFoundError()
				mockRepo.EXPECT().
					FindByID(gomock.Any(), notFoundID).
					Return(nil, appErr).
					Times(1)
			},
			wantErr: trip.NewTripNotFoundError(),
		},
		{
			name:     "異常系: 取得時に予期しないエラーが返される",
			id:       "error-id",
			tripName: "Updated Trip",
			setup: func() {
				mockTimeService.EXPECT().
					Now().
					Return(updateTime).
					Times(1)

				errorID := trip.NewTripID("error-id")
				unexpectedErr := errors.New("database connection error")
				mockRepo.EXPECT().
					FindByID(gomock.Any(), errorID).
					Return(nil, unexpectedErr).
					Times(1)
			},
			wantErr: apperr.NewInternalError("Failed to get trip for update", apperr.WithCause(errors.New("database connection error"))),
		},
		{
			name:     "異常系: 更新時にアプリケーションエラーが返される",
			id:       "test-id",
			tripName: "Updated Trip",
			setup: func() {
				mockTimeService.EXPECT().
					Now().
					Return(updateTime).
					Times(1)

				mockRepo.EXPECT().
					FindByID(gomock.Any(), tripID).
					Return(originalTrip, nil).
					Times(1)

				appErr := apperr.NewInternalError("Database error")
				updatedTrip := originalTrip.Update("Updated Trip", updateTime)
				mockRepo.EXPECT().
					Update(gomock.Any(), updatedTrip).
					Return(appErr).
					Times(1)
			},
			wantErr: apperr.NewInternalError("Database error"),
		},
		{
			name:     "異常系: 更新時に予期しないエラーが返される",
			id:       "test-id",
			tripName: "Updated Trip",
			setup: func() {
				mockTimeService.EXPECT().
					Now().
					Return(updateTime).
					Times(1)

				mockRepo.EXPECT().
					FindByID(gomock.Any(), tripID).
					Return(originalTrip, nil).
					Times(1)

				unexpectedErr := errors.New("database update error")
				updatedTrip := originalTrip.Update("Updated Trip", updateTime)
				mockRepo.EXPECT().
					Update(gomock.Any(), updatedTrip).
					Return(unexpectedErr).
					Times(1)
			},
			wantErr: apperr.NewInternalError("Failed to update trip", apperr.WithCause(errors.New("database update error"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := interactor.Update(context.Background(), tt.id, tt.tripName)

			if tt.wantErr != nil {
				require.Error(t, err)
				if appErr, ok := tt.wantErr.(*apperr.AppError); ok {
					gotAppErr, ok := err.(*apperr.AppError)
					require.True(t, ok, "Expected AppError but got %T", err)
					assert.Equal(t, appErr.Code, gotAppErr.Code)
					assert.Equal(t, appErr.Message, gotAppErr.Message)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTripInteractor_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_trip.NewMockTripRepository(ctrl)
	mockTimeService := mock_service.NewMockTimeService(ctrl)
	mockIDService := mock_service.NewMockIDService(ctrl)

	interactor := NewTripInteractor(mockRepo, mockTimeService, mockIDService)

	tests := []struct {
		name    string
		id      string
		setup   func()
		wantErr error
	}{
		{
			name: "正常系: 旅行が正常に削除できる",
			id:   "test-id",
			setup: func() {
				tripID := trip.NewTripID("test-id")
				mockRepo.EXPECT().
					Delete(gomock.Any(), tripID).
					Return(nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name: "異常系: リポジトリからアプリケーションエラーが返される",
			id:   "not-found-id",
			setup: func() {
				notFoundID := trip.NewTripID("not-found-id")
				appErr := trip.NewTripNotFoundError()
				mockRepo.EXPECT().
					Delete(gomock.Any(), notFoundID).
					Return(appErr).
					Times(1)
			},
			wantErr: trip.NewTripNotFoundError(),
		},
		{
			name: "異常系: リポジトリから予期しないエラーが返される",
			id:   "error-id",
			setup: func() {
				errorID := trip.NewTripID("error-id")
				unexpectedErr := errors.New("database delete error")
				mockRepo.EXPECT().
					Delete(gomock.Any(), errorID).
					Return(unexpectedErr).
					Times(1)
			},
			wantErr: apperr.NewInternalError("Failed to delete trip", apperr.WithCause(errors.New("database delete error"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := interactor.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if appErr, ok := tt.wantErr.(*apperr.AppError); ok {
					gotAppErr, ok := err.(*apperr.AppError)
					require.True(t, ok, "Expected AppError but got %T", err)
					assert.Equal(t, appErr.Code, gotAppErr.Code)
					assert.Equal(t, appErr.Message, gotAppErr.Message)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// NewTripInteractor のテスト
func TestNewTripInteractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_trip.NewMockTripRepository(ctrl)
	mockTimeService := mock_service.NewMockTimeService(ctrl)
	mockIDService := mock_service.NewMockIDService(ctrl)

	interactor := NewTripInteractor(mockRepo, mockTimeService, mockIDService)

	assert.NotNil(t, interactor)
}
