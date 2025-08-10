package usecase

import (
	"context"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/clock"
	domain_errors "travel-api/internal/domain/shared/errors"
	"travel-api/internal/domain/shared/uuid"
	shared_errors "travel-api/internal/shared/errors"
	"travel-api/internal/usecase/output"
)

//go:generate mockgen -destination mock/trip.go travel-api/internal/usecase TripUsecase
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
		return output.GetTripOutput{}, shared_errors.NewInternalError("trip id creation failed", err)
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if domain_errors.IsTripNotFound(err) {
			return output.GetTripOutput{}, shared_errors.NewNotFoundError("trip")
		}

		return output.GetTripOutput{}, shared_errors.NewInternalError("failed to find trip", err)
	}

	return output.NewGetTripOutput(trip), nil
}

func (i *TripInteractor) List(ctx context.Context) (output.ListTripOutput, error) {
	trips, err := i.repository.FindMany(ctx)
	if err != nil {
		return output.ListTripOutput{}, shared_errors.NewInternalError("failed to find trips", err)
	}

	return output.NewListTripOutput(trips), nil
}

func (i *TripInteractor) Create(ctx context.Context, name string) (string, error) {
	newUUID := i.uuidGenerator.NewUUID()
	tripID, err := domain.NewTripID(newUUID)
	if err != nil {
		return "", shared_errors.NewInternalError("trip id generation failed", err)
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
		return "", shared_errors.NewInternalError("failed to create trip", err)
	}

	return tripID.String(), nil
}

func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return shared_errors.NewInternalError("trip id creation failed", err)
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if domain_errors.IsTripNotFound(err) {
			return shared_errors.NewNotFoundError("trip")
		}
		return shared_errors.NewInternalError("failed to find trip", err)
	}

	updatedTrip := trip.Update(name, i.clock.Now())

	if err := i.repository.Update(ctx, updatedTrip); err != nil {
		return shared_errors.NewInternalError("failed to update trip", err)
	}

	return nil
}

func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return shared_errors.NewInternalError("trip id creation failed", err)
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		if domain_errors.IsTripNotFound(err) {
			return shared_errors.NewNotFoundError("trip")
		}
		return shared_errors.NewInternalError("failed to find trip", err)
	}

	if err := i.repository.Delete(ctx, trip); err != nil {
		return shared_errors.NewInternalError("failed to delete trip", err)
	}

	return nil
}
