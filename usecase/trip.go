package usecase

import (
	"context"
	"time"
	"travel-api/domain"
)

type TripInteractor struct {
	repository domain.TripRepository
}

func NewTripInteractor(repository domain.TripRepository) *TripInteractor {
	return &TripInteractor{
		repository: repository,
	}
}

func (i *TripInteractor) Get(ctx context.Context, id string) (domain.Trip, error) {
	return i.repository.FindByID(ctx, domain.TripID(id))
}

func (i *TripInteractor) List(ctx context.Context) ([]domain.Trip, error) {
	return i.repository.FindMany(ctx)
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
