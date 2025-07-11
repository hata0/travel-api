package domain

import (
	"time"
)

type TripId string

type Trip struct {
	id        TripId
	name      string
	createdAt time.Time
	updatedAt time.Time
}

func NewTrip(id TripId, name string, createdAt time.Time, updatedAt time.Time) Trip {
	return Trip{
		id:        id,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (t Trip) Update(name string, updatedAt time.Time) Trip {
	return Trip{
		id:        t.id,
		name:      name,
		createdAt: t.createdAt,
		updatedAt: updatedAt,
	}
}
