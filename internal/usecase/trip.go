package usecase

import (
	"context"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/clock"
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
	uuidGenerator domain.UUIDGenerator
}

func NewTripInteractor(repository domain.TripRepository, clock clock.Clock, uuidGenerator domain.UUIDGenerator) *TripInteractor {
	return &TripInteractor{
		repository:    repository,
		clock:         clock,
		uuidGenerator: uuidGenerator,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (output.GetTripOutput, error) {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return output.GetTripOutput{}, err
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		return output.GetTripOutput{}, err
	}

	return output.NewGetTripOutput(trip), nil
}

func (i *TripInteractor) List(ctx context.Context) (output.ListTripOutput, error) {
	trips, err := i.repository.FindMany(ctx)
	if err != nil {
		return output.ListTripOutput{}, err
	}

	return output.NewListTripOutput(trips), nil
}

func (i *TripInteractor) Create(ctx context.Context, name string) (string, error) {
	newUUID := i.uuidGenerator.NewUUID()
	tripID, err := domain.NewTripID(newUUID)
	if err != nil {
		return "", err
	}

	trip := domain.NewTrip(
		tripID,
		name,
		i.clock.Now(),
		i.clock.Now(),
	)

	err = i.repository.Create(ctx, trip)
	if err != nil {
		return "", err
	}

	return tripID.String(), nil
}

func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return err
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		return err
	}

	trip = trip.Update(name, i.clock.Now())

	return i.repository.Update(ctx, trip)
}

func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	tripID, err := domain.NewTripID(id)
	if err != nil {
		return err
	}

	trip, err := i.repository.FindByID(ctx, tripID)
	if err != nil {
		return err
	}

	return i.repository.Delete(ctx, trip)
}
