package usecase

import (
	"context"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/usecase/output"
)

type TripInteractor struct {
	repository domain.TripRepository
}

func NewTripInteractor(repository domain.TripRepository) *TripInteractor {
	return &TripInteractor{
		repository: repository,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (output.GetTripOutput, error) {
	trip, err := i.repository.FindByID(ctx, domain.TripID(id))
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

func (i *TripInteractor) Create(ctx context.Context, name string) error {
	trip := domain.NewTrip(
		domain.TripID(domain.NewUUID()),
		name,
		time.Now(),
		time.Now(),
	)

	return i.repository.Create(ctx, trip)
}

func (i *TripInteractor) Update(ctx context.Context, id string, name string) error {
	trip, err := i.repository.FindByID(ctx, domain.TripID(id))
	if err != nil {
		return err
	}

	trip = trip.Update(name, time.Now())

	return i.repository.Update(ctx, trip)
}

func (i *TripInteractor) Delete(ctx context.Context, id string) error {
	trip, err := i.repository.FindByID(ctx, domain.TripID(id))
	if err != nil {
		return err
	}

	return i.repository.Delete(ctx, trip)
}
