package output

import (
	"time"

	"github.com/hata0/travel-api/internal/domain/trip"
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

func NewGetTripOutput(trip *trip.Trip) *GetTripOutput {
	return &GetTripOutput{
		Trip: mapToTrip(trip),
	}
}

type ListTripOutput struct {
	Trips []*Trip
}

func NewListTripOutput(trips []*trip.Trip) *ListTripOutput {
	formattedTrips := make([]*Trip, len(trips))
	for _, trip := range trips {
		formattedTrips = append(formattedTrips, mapToTrip(trip))
	}

	return &ListTripOutput{
		Trips: formattedTrips,
	}
}

type CreateTripOutput struct {
	ID string
}

func NewCreateTripOutput(id trip.TripID) *CreateTripOutput {
	return &CreateTripOutput{
		ID: id.String(),
	}
}

func mapToTrip(trip *trip.Trip) *Trip {
	return &Trip{
		ID:        trip.ID().String(),
		Name:      trip.Name(),
		CreatedAt: trip.CreatedAt(),
		UpdatedAt: trip.UpdatedAt(),
	}
}
