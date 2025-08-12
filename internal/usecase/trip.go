package usecase

import (
	"context"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/usecase/output"
	"github.com/hata0/travel-api/internal/usecase/services"
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
	repository   domain.TripRepository
	timeProvider services.TimeProvider
	idGenerator  services.IDGenerator
}

func NewTripInteractor(repository domain.TripRepository, timeProvider services.TimeProvider, idGenerator services.IDGenerator) *TripInteractor {
	return &TripInteractor{
		repository:   repository,
		timeProvider: timeProvider,
		idGenerator:  idGenerator,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (*output.GetTripOutput, error) {
	tripID := domain.NewTripID(id)

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			// TODO: messageを設定する
			return nil, apperr.NewNotFoundError("")
		}

		return nil, apperr.NewInternalError("", err)
	}

	return output.NewGetTripOutput(trip), nil
}

func (i *TripInteractor) List(ctx context.Context) (*output.ListTripOutput, error) {
	trips, err := i.repository.FindMany(ctx)
	if err != nil {
		return nil, apperr.NewInternalError("", err)
	}

	return output.NewListTripOutput(trips), nil
}

func (i *TripInteractor) Create(ctx context.Context, name string) (string, error) {
	newID := i.idGenerator.Generate()
	now := i.timeProvider.Now()

	tripID := domain.NewTripID(newID)

	trip := domain.NewTrip(
		tripID,
		name,
		now,
		now,
	)

	err := i.repository.Create(ctx, trip)
	if err != nil {
		return "", apperr.NewInternalError("", err)
	}

	return tripID.String(), nil
}

func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	now := i.timeProvider.Now()

	tripID := domain.NewTripID(id)

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			return apperr.NewNotFoundError("")
		}
		return apperr.NewInternalError("", err)
	}

	updatedTrip := trip.Update(name, now)

	if err := i.repository.Update(ctx, updatedTrip); err != nil {
		return apperr.NewInternalError("", err)
	}

	return nil
}

func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	tripID := domain.NewTripID(id)

	if err := i.repository.Delete(ctx, tripID); err != nil {
		if apperr.IsTripNotFound(err) {
			return apperr.NewNotFoundError("")
		}
		return apperr.NewInternalError("", err)
	}

	return nil
}
