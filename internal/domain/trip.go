package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/trip.go travel-api/internal/domain TripRepository
type TripRepository interface {
	FindByID(ctx context.Context, id TripID) (Trip, error)
	FindMany(ctx context.Context) ([]Trip, error)
	Create(ctx context.Context, trip Trip) error
	Update(ctx context.Context, trip Trip) error
	Delete(ctx context.Context, trip Trip) error
}

type TripID struct {
	value string
}

func NewTripID(id string) (TripID, error) {
	if !IsValidUUID(id) {
		return TripID{}, ErrInvalidUUID
	}
	return TripID{value: id}, nil
}

func (id TripID) String() string {
	return id.value
}

type Trip struct {
	ID        TripID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTrip(id TripID, name string, createdAt time.Time, updatedAt time.Time) Trip {
	return Trip{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (t Trip) Update(name string, updatedAt time.Time) Trip {
	return Trip{
		ID:        t.ID,
		Name:      name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: updatedAt,
	}
}
