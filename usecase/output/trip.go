package output

import (
	"time"
	"travel-api/domain"
)

type Trip struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetTripOutput struct {
	Trip Trip
}

func NewGetTripOutput(trip domain.Trip) GetTripOutput {
	return GetTripOutput{
		Trip: mapToTrip(trip),
	}
}

type ListTripOutput struct {
	Trips []Trip
}

func NewListTripOutput(trips []domain.Trip) ListTripOutput {
	formattedTrips := make([]Trip, len(trips))
	for i, trip := range trips {
		formattedTrips[i] = mapToTrip(trip)
	}

	return ListTripOutput{
		Trips: formattedTrips,
	}
}

func mapToTrip(trip domain.Trip) Trip {
	return Trip{
		ID:        string(trip.ID),
		Name:      trip.Name,
		CreatedAt: trip.CreatedAt,
		UpdatedAt: trip.UpdatedAt,
	}
}
