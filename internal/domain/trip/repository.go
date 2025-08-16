package trip

import "context"

//go:generate mockgen -destination mock/trip.go github.com/hata0/travel-api/internal/domain/trip TripRepository
type TripRepository interface {
	FindByID(ctx context.Context, id TripID) (*Trip, error)
	FindMany(ctx context.Context) ([]*Trip, error)
	Create(ctx context.Context, trip *Trip) error
	Update(ctx context.Context, trip *Trip) error
	Delete(ctx context.Context, id TripID) error
}
