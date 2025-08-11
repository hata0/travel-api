package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/trip.go github.com/hata0/travel-api/internal/domain TripRepository
type TripRepository interface {
	FindByID(ctx context.Context, id TripID) (*Trip, error)
	FindMany(ctx context.Context) ([]*Trip, error)
	Create(ctx context.Context, trip *Trip) error
	Update(ctx context.Context, trip *Trip) error
	Delete(ctx context.Context, id TripID) error
}

// TripID は旅行IDを表現する値オブジェクト
type TripID struct {
	value string
}

func NewTripID(id string) TripID {
	return TripID{value: id}
}

func (id TripID) String() string {
	return id.value
}

func (id TripID) Equals(other TripID) bool {
	return id.value == other.value
}

// Trip は旅行を表現するエンティティ
type Trip struct {
	id        TripID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

// NewTrip は新しい旅行を作成する
func NewTrip(id TripID, name string, createdAt, updatedAt time.Time) *Trip {
	return &Trip{
		id:        id,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters
func (t *Trip) ID() TripID           { return t.id }
func (t *Trip) Name() string         { return t.name }
func (t *Trip) CreatedAt() time.Time { return t.createdAt }
func (t *Trip) UpdatedAt() time.Time { return t.updatedAt }

// Update は旅行情報を更新する
func (t *Trip) Update(name string, updatedAt time.Time) *Trip {
	return &Trip{
		id:        t.id,
		name:      name,
		createdAt: t.createdAt,
		updatedAt: updatedAt,
	}
}

func (t *Trip) Equals(other *Trip) bool {
	if other == nil {
		return false
	}
	return t.id.Equals(other.id)
}
