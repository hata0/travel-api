package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	mock_domain "github.com/hata0/travel-api/internal/domain/mock"
	"github.com/hata0/travel-api/internal/usecase/output"
	"github.com/hata0/travel-api/internal/usecase/service"
	mock_services "github.com/hata0/travel-api/internal/usecase/service/mock"
)

func TestTripInteractor_Get(t *testing.T) {
	type fields struct {
		repository   func(ctrl *gomock.Controller) domain.TripRepository
		timeProvider func(ctrl *gomock.Controller) service.TimeService
		idGenerator  func(ctrl *gomock.Controller) service.IDService
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *output.GetTripOutput
		wantErr     bool
		expectedErr *apperr.AppError
	}{
		{
			name: "正常系: Tripが正常に取得できる場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					expectedTrip := domain.NewTrip(
						domain.NewTripID("test-id"),
						"Tokyo Trip",
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("test-id")).Return(expectedTrip, nil)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "test-id",
			},
			want: output.NewGetTripOutput(domain.NewTrip(
				domain.NewTripID("test-id"),
				"Tokyo Trip",
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			)),
			wantErr: false,
		},
		{
			name: "異常系: Tripが見つからない場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("nonexistent-id")).
						Return(nil, apperr.ErrTripNotFound)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "nonexistent-id",
			},
			want:        nil,
			wantErr:     true,
			expectedErr: apperr.NewNotFoundError(""),
		},
		{
			name: "異常系: リポジトリでエラーが発生した場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("test-id")).
						Return(nil, apperr.NewInternalError(""))
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "test-id",
			},
			want:        nil,
			wantErr:     true,
			expectedErr: apperr.NewInternalError(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			i := &TripInteractor{
				repository:  tt.fields.repository(ctrl),
				timeService: tt.fields.timeProvider(ctrl),
				idService:   tt.fields.idGenerator(ctrl),
			}

			got, err := i.Get(tt.args.ctx, tt.args.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTripInteractor_List(t *testing.T) {
	type fields struct {
		repository   func(ctrl *gomock.Controller) domain.TripRepository
		timeProvider func(ctrl *gomock.Controller) service.TimeService
		idGenerator  func(ctrl *gomock.Controller) service.IDService
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *output.ListTripOutput
		wantErr     bool
		expectedErr *apperr.AppError
	}{
		{
			name: "正常系: Tripリストが正常に取得できる場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					trips := []*domain.Trip{
						domain.NewTrip(
							domain.NewTripID("trip-1"),
							"Tokyo Trip",
							time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						),
						domain.NewTrip(
							domain.NewTripID("trip-2"),
							"Osaka Trip",
							time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
							time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
						),
					}
					repo.EXPECT().FindMany(gomock.Any()).Return(trips, nil)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: output.NewListTripOutput([]*domain.Trip{
				domain.NewTrip(
					domain.NewTripID("trip-1"),
					"Tokyo Trip",
					time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				),
				domain.NewTrip(
					domain.NewTripID("trip-2"),
					"Osaka Trip",
					time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
				),
			}),
			wantErr: false,
		},
		{
			name: "正常系: 空のTripリストが返される場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindMany(gomock.Any()).Return([]*domain.Trip{}, nil)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:    output.NewListTripOutput([]*domain.Trip{}),
			wantErr: false,
		},
		{
			name: "異常系: リポジトリでエラーが発生した場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindMany(gomock.Any()).Return(nil, apperr.NewInternalError(""))
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want:        nil,
			wantErr:     true,
			expectedErr: apperr.NewInternalError(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			i := &TripInteractor{
				repository:  tt.fields.repository(ctrl),
				timeService: tt.fields.timeProvider(ctrl),
				idService:   tt.fields.idGenerator(ctrl),
			}

			got, err := i.List(tt.args.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTripInteractor_Create(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	generatedID := "generated-id"

	type fields struct {
		repository   func(ctrl *gomock.Controller) domain.TripRepository
		timeProvider func(ctrl *gomock.Controller) service.TimeService
		idGenerator  func(ctrl *gomock.Controller) service.IDService
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        string
		wantErr     bool
		expectedErr *apperr.AppError
	}{
		{
			name: "正常系: Tripが正常に作成される場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					expectedTrip := domain.NewTrip(
						domain.NewTripID(generatedID),
						"New Trip",
						fixedTime,
						fixedTime,
					)
					repo.EXPECT().Create(gomock.Any(), expectedTrip).Return(nil)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					provider := mock_services.NewMockTimeProvider(ctrl)
					provider.EXPECT().Now().Return(fixedTime)
					return provider
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					generator := mock_services.NewMockIDGenerator(ctrl)
					generator.EXPECT().Generate().Return(generatedID)
					return generator
				},
			},
			args: args{
				ctx:  context.Background(),
				name: "New Trip",
			},
			want:    generatedID,
			wantErr: false,
		},
		{
			name: "異常系: リポジトリでエラーが発生した場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(apperr.NewInternalError(""))
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					provider := mock_services.NewMockTimeProvider(ctrl)
					provider.EXPECT().Now().Return(fixedTime)
					return provider
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					generator := mock_services.NewMockIDGenerator(ctrl)
					generator.EXPECT().Generate().Return(generatedID)
					return generator
				},
			},
			args: args{
				ctx:  context.Background(),
				name: "New Trip",
			},
			want:        "",
			wantErr:     true,
			expectedErr: apperr.NewInternalError(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			i := &TripInteractor{
				repository:  tt.fields.repository(ctrl),
				timeService: tt.fields.timeProvider(ctrl),
				idService:   tt.fields.idGenerator(ctrl),
			}

			got, err := i.Create(tt.args.ctx, tt.args.name)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", got)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTripInteractor_Update(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	originalTrip := domain.NewTrip(
		domain.NewTripID("test-id"),
		"Original Name",
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	type fields struct {
		repository   func(ctrl *gomock.Controller) domain.TripRepository
		timeProvider func(ctrl *gomock.Controller) service.TimeService
		idGenerator  func(ctrl *gomock.Controller) service.IDService
	}
	type args struct {
		ctx  context.Context
		id   string
		name string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		expectedErr *apperr.AppError
	}{
		{
			name: "正常系: Tripが正常に更新される場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("test-id")).
						Return(originalTrip, nil)

					updatedTrip := originalTrip.Update("Updated Name", fixedTime)
					repo.EXPECT().Update(gomock.Any(), updatedTrip).Return(nil)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					provider := mock_services.NewMockTimeProvider(ctrl)
					provider.EXPECT().Now().Return(fixedTime)
					return provider
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   "test-id",
				name: "Updated Name",
			},
			wantErr: false,
		},
		{
			name: "異常系: 更新対象のTripが見つからない場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("nonexistent-id")).
						Return(nil, apperr.ErrTripNotFound)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					provider := mock_services.NewMockTimeProvider(ctrl)
					provider.EXPECT().Now().Return(fixedTime)
					return provider
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   "nonexistent-id",
				name: "Updated Name",
			},
			wantErr:     true,
			expectedErr: apperr.NewNotFoundError(""),
		},
		{
			name: "異常系: Trip取得時にリポジトリでエラーが発生した場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("test-id")).
						Return(nil, apperr.NewInternalError(""))
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					provider := mock_services.NewMockTimeProvider(ctrl)
					provider.EXPECT().Now().Return(fixedTime)
					return provider
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   "test-id",
				name: "Updated Name",
			},
			wantErr:     true,
			expectedErr: apperr.NewInternalError(""),
		},
		{
			name: "異常系: Trip更新時にリポジトリでエラーが発生した場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().FindByID(gomock.Any(), domain.NewTripID("test-id")).
						Return(originalTrip, nil)
					repo.EXPECT().Update(gomock.Any(), gomock.Any()).
						Return(apperr.NewInternalError(""))
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					provider := mock_services.NewMockTimeProvider(ctrl)
					provider.EXPECT().Now().Return(fixedTime)
					return provider
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   "test-id",
				name: "Updated Name",
			},
			wantErr:     true,
			expectedErr: apperr.NewInternalError(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			i := &TripInteractor{
				repository:  tt.fields.repository(ctrl),
				timeService: tt.fields.timeProvider(ctrl),
				idService:   tt.fields.idGenerator(ctrl),
			}

			err := i.Update(tt.args.ctx, tt.args.id, tt.args.name)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTripInteractor_Delete(t *testing.T) {
	type fields struct {
		repository   func(ctrl *gomock.Controller) domain.TripRepository
		timeProvider func(ctrl *gomock.Controller) service.TimeService
		idGenerator  func(ctrl *gomock.Controller) service.IDService
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		expectedErr *apperr.AppError
	}{
		{
			name: "正常系: Tripが正常に削除される場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().Delete(gomock.Any(), domain.NewTripID("test-id")).Return(nil)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "test-id",
			},
			wantErr: false,
		},
		{
			name: "異常系: 削除対象のTripが見つからない場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().Delete(gomock.Any(), domain.NewTripID("nonexistent-id")).
						Return(apperr.ErrTripNotFound)
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "nonexistent-id",
			},
			wantErr:     true,
			expectedErr: apperr.NewNotFoundError(""),
		},
		{
			name: "異常系: リポジトリでエラーが発生した場合",
			fields: fields{
				repository: func(ctrl *gomock.Controller) domain.TripRepository {
					repo := mock_domain.NewMockTripRepository(ctrl)
					repo.EXPECT().Delete(gomock.Any(), domain.NewTripID("test-id")).
						Return(apperr.NewInternalError(""))
					return repo
				},
				timeProvider: func(ctrl *gomock.Controller) service.TimeService {
					return mock_services.NewMockTimeProvider(ctrl)
				},
				idGenerator: func(ctrl *gomock.Controller) service.IDService {
					return mock_services.NewMockIDGenerator(ctrl)
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "test-id",
			},
			wantErr:     true,
			expectedErr: apperr.NewInternalError(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			i := &TripInteractor{
				repository:  tt.fields.repository(ctrl),
				timeService: tt.fields.timeProvider(ctrl),
				idService:   tt.fields.idGenerator(ctrl),
			}

			err := i.Delete(tt.args.ctx, tt.args.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewTripInteractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repository := mock_domain.NewMockTripRepository(ctrl)
	timeProvider := mock_services.NewMockTimeProvider(ctrl)
	idGenerator := mock_services.NewMockIDGenerator(ctrl)

	interactor := NewTripInteractor(repository, timeProvider, idGenerator)

	assert.NotNil(t, interactor)
	assert.Equal(t, repository, interactor.repository)
	assert.Equal(t, timeProvider, interactor.timeService)
	assert.Equal(t, idGenerator, interactor.idService)
}
