package usecase

import (
	"context"

	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/domain/trip"
	"github.com/hata0/travel-api/internal/usecase/output"
	"github.com/hata0/travel-api/internal/usecase/service"
)

//go:generate mockgen -destination mock/trip.go github.com/hata0/travel-api/internal/usecase TripUsecase
type TripUsecase interface {
	Get(ctx context.Context, id string) (*output.GetTripOutput, error)
	List(ctx context.Context) (*output.ListTripOutput, error)
	Create(ctx context.Context, name string) (*output.CreateTripOutput, error)
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type TripInteractor struct {
	repository  trip.TripRepository
	timeService service.TimeService
	idService   service.IDService
}

func NewTripInteractor(repository trip.TripRepository, timeService service.TimeService, idService service.IDService) TripUsecase {
	return &TripInteractor{
		repository:  repository,
		timeService: timeService,
		idService:   idService,
	}
}

// Get は指定されたIDの旅行を取得する
func (i *TripInteractor) Get(ctx context.Context, id string) (*output.GetTripOutput, error) {
	tripID := trip.NewTripID(id)

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsAppError(err) {
			return nil, err
		}
		return nil, apperr.NewInternalError("Failed to get trip", apperr.WithCause(err))
	}

	return output.NewGetTripOutput(trip), nil
}

// List はすべての旅行を取得する
func (i *TripInteractor) List(ctx context.Context) (*output.ListTripOutput, error) {
	trips, err := i.repository.FindMany(ctx)
	if err != nil {
		if apperr.IsAppError(err) {
			return nil, err
		}
		return nil, apperr.NewInternalError("Failed to list trips", apperr.WithCause(err))
	}

	return output.NewListTripOutput(trips), nil
}

// Create は新しい旅行を作成する
func (i *TripInteractor) Create(ctx context.Context, name string) (*output.CreateTripOutput, error) {
	newID := i.idService.Generate()
	now := i.timeService.Now()

	tripID := trip.NewTripID(newID)

	trip := trip.NewTrip(
		tripID,
		name,
		now,
		now,
	)

	err := i.repository.Create(ctx, trip)
	if err != nil {
		if apperr.IsAppError(err) {
			return nil, err
		}
		return nil, apperr.NewInternalError("Failed to create trip", apperr.WithCause(err))
	}

	return output.NewCreateTripOutput(tripID), nil
}

// Update は既存の旅行を更新する
func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	now := i.timeService.Now()

	tripID := trip.NewTripID(id)

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsAppError(err) {
			return err
		}
		return apperr.NewInternalError("Failed to get trip for update", apperr.WithCause(err))
	}

	updatedTrip := trip.Update(name, now)

	if err := i.repository.Update(ctx, updatedTrip); err != nil {
		if apperr.IsAppError(err) {
			return err
		}
		return apperr.NewInternalError("Failed to update trip", apperr.WithCause(err))
	}

	return nil
}

// Delete は指定されたIDの旅行を削除する
func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	tripID := trip.NewTripID(id)

	if err := i.repository.Delete(ctx, tripID); err != nil {
		if apperr.IsAppError(err) {
			return err
		}
		return apperr.NewInternalError("Failed to delete trip", apperr.WithCause(err))
	}

	return nil
}
