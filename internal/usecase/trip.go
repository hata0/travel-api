package usecase

import (
	"context"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/usecase/output"
	"github.com/hata0/travel-api/internal/usecase/service"
)

//go:generate mockgen -destination mock/trip.go github.com/hata0/travel-api/internal/usecase TripUsecase
type TripUsecase interface {
	Get(ctx context.Context, id string) (*output.GetTripOutput, error)
	List(ctx context.Context) (*output.ListTripOutput, error)
	Create(ctx context.Context, name string) (string, error)
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type TripInteractor struct {
	repository  domain.TripRepository
	timeService service.TimeService
	idService   service.IDService
}

func NewTripInteractor(repository domain.TripRepository, timeService service.TimeService, idService service.IDService) *TripInteractor {
	return &TripInteractor{
		repository:  repository,
		timeService: timeService,
		idService:   idService,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (*output.GetTripOutput, error) {
	tripID := domain.NewTripID(id)

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			return nil, apperr.NewNotFoundError("Trip not found")
		}

		return nil, apperr.NewInternalError("Failed to retrieve trip", err)
	}

	return output.NewGetTripOutput(trip), nil
}

func (i *TripInteractor) List(ctx context.Context) (*output.ListTripOutput, error) {
	trips, err := i.repository.FindMany(ctx)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to retrieve trips", err)
	}

	return output.NewListTripOutput(trips), nil
}

func (i *TripInteractor) Create(ctx context.Context, name string) (string, error) {
	newID := i.idService.Generate()
	now := i.timeService.Now()

	tripID := domain.NewTripID(newID)

	trip := domain.NewTrip(
		tripID,
		name,
		now,
		now,
	)

	err := i.repository.Create(ctx, trip)
	if err != nil {
		return "", apperr.NewInternalError("Failed to create trip", err)
	}

	return tripID.String(), nil
}

func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	now := i.timeService.Now()

	tripID := domain.NewTripID(id)

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			return apperr.NewNotFoundError("Trip not found")
		}
		return apperr.NewInternalError("Failed to retrieve trip for update", err)
	}

	updatedTrip := trip.Update(name, now)

	if err := i.repository.Update(ctx, updatedTrip); err != nil {
		return apperr.NewInternalError("Failed to update trip", err)
	}

	return nil
}

func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	tripID := domain.NewTripID(id)

	if err := i.repository.Delete(ctx, tripID); err != nil {
		if apperr.IsTripNotFound(err) {
			return apperr.NewNotFoundError("Trip not found")
		}
		return apperr.NewInternalError("Failed to delete trip", err)
	}

	return nil
}
