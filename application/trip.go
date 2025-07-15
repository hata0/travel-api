package application

import (
	"context"
	"time"
	"travel-api/domain"
)

type TripServiceImpl struct {
	repository domain.TripRepository
}

func NewTripService(repository domain.TripRepository) TripServiceImpl {
	return TripServiceImpl{
		repository: repository,
	}
}

func (s TripServiceImpl) Get(ctx context.Context, id string) (domain.Trip, error) {
	return s.repository.FindByID(ctx, domain.TripID(id))
}

func (s TripServiceImpl) List(ctx context.Context) ([]domain.Trip, error) {
	return s.repository.FindMany(ctx)
}

func (s TripServiceImpl) Create(ctx context.Context, name string) error {
	trip := domain.NewTrip(
		domain.TripID(domain.NewUUID()),
		name,
		time.Now(),
		time.Now(),
	)

	return s.repository.Create(ctx, trip)
}

func (s TripServiceImpl) Update(ctx context.Context, id string, name string) error {
	trip, err := s.repository.FindByID(ctx, domain.TripID(id))
	if err != nil {
		return err
	}

	trip = trip.Update(name, time.Now())

	return s.repository.Update(ctx, trip)
}

func (s TripServiceImpl) Delete(ctx context.Context, id string) error {
	trip, err := s.repository.FindByID(ctx, domain.TripID(id))
	if err != nil {
		return err
	}

	return s.repository.Delete(ctx, trip)
}
