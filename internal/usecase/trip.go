package usecase

import (
	"context"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/domain/shared/clock"
	"github.com/hata0/travel-api/internal/domain/shared/uuid"
	"github.com/hata0/travel-api/internal/usecase/output"
)

//go:generate mockgen -destination mock/trip.go github.com/hata0/travel-api/internal/usecase TripUsecase
type TripUsecase interface {
	Get(ctx context.Context, id string) (output.GetTripOutput, error)
	List(ctx context.Context) (output.ListTripOutput, error)
	Create(ctx context.Context, name string) (string, error)
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type TripInteractor struct {
	repository    domain.TripRepository
	clock         clock.Clock
	uuidGenerator uuid.UUIDGenerator
}

func NewTripInteractor(repository domain.TripRepository, clock clock.Clock, uuidGenerator uuid.UUIDGenerator) *TripInteractor {
	return &TripInteractor{
		repository:    repository,
		clock:         clock,
		uuidGenerator: uuidGenerator,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (output.GetTripOutput, error) {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return output.GetTripOutput{}, apperr.NewInternalError("trip id creation failed", err)
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			return output.GetTripOutput{}, apperr.NewNotFoundError("trip")
		}

		return output.GetTripOutput{}, apperr.NewInternalError("failed to find trip", err)
	}

	return output.NewGetTripOutput(trip), nil
}

func (i *TripInteractor) List(ctx context.Context) (output.ListTripOutput, error) {
	trips, err := i.repository.FindMany(ctx)
	if err != nil {
		return output.ListTripOutput{}, apperr.NewInternalError("failed to find trips", err)
	}

	return output.NewListTripOutput(trips), nil
}

func (i *TripInteractor) Create(ctx context.Context, name string) (string, error) {
	newUUID := i.uuidGenerator.NewUUID()
	tripID, err := domain.NewTripID(newUUID)
	if err != nil {
		return "", apperr.NewInternalError("trip id generation failed", err)
	}

	now := i.clock.Now()
	trip := domain.NewTrip(
		tripID,
		name,
		now,
		now,
	)

	err = i.repository.Create(ctx, trip)
	if err != nil {
		return "", apperr.NewInternalError("failed to create trip", err)
	}

	return tripID.String(), nil
}

func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return apperr.NewInternalError("trip id creation failed", err)
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			return apperr.NewNotFoundError("trip")
		}
		return apperr.NewInternalError("failed to find trip", err)
	}

	updatedTrip := trip.Update(name, i.clock.Now())

	if err := i.repository.Update(ctx, updatedTrip); err != nil {
		return apperr.NewInternalError("failed to update trip", err)
	}

	return nil
}

func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return apperr.NewInternalError("trip id creation failed", err)
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if apperr.IsTripNotFound(err) {
			return apperr.NewNotFoundError("trip")
		}
		return apperr.NewInternalError("failed to find trip", err)
	}

	if err := i.repository.Delete(ctx, trip); err != nil {
		return apperr.NewInternalError("failed to delete trip", err)
	}

	return nil
}
