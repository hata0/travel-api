package output

import (
	"time"

	"github.com/hata0/travel-api/internal/domain"
)

type Trip struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetTripOutput struct {
	Trip *Trip
}

func NewGetTripOutput(trip *domain.Trip) *GetTripOutput {
	return &GetTripOutput{
		Trip: mapToTrip(trip),
	}
}

type ListTripOutput struct {
	Trips []*Trip
}

func NewListTripOutput(trips []*domain.Trip) *ListTripOutput {
	formattedTrips := make([]*Trip, len(trips))
	for _, trip := range trips {
		formattedTrips = append(formattedTrips, mapToTrip(trip))
	}

	return &ListTripOutput{
		Trips: formattedTrips,
	}
}

func mapToTrip(trip *domain.Trip) *Trip {
	return &Trip{
		ID:        trip.ID().String(),
		Name:      trip.Name(),
		CreatedAt: trip.CreatedAt(),
		UpdatedAt: trip.UpdatedAt(),
	}
}
