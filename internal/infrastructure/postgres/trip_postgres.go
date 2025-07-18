package postgres

import (
	"context"
	"time"
	"travel-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TripPostgresRepository struct {
	queries *Queries
}

func NewTripPostgresRepository(db DBTX) domain.TripRepository {
	return &TripPostgresRepository{
		queries: New(db),
	}
}

func (r *TripPostgresRepository) FindByID(ctx context.Context, id domain.TripID) (domain.Trip, error) {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(id.String())

	record, err := r.queries.GetTrip(ctx, validatedId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Trip{}, domain.ErrTripNotFound
		} else {
			return domain.Trip{}, err
		}
	}

	return r.mapToTrip(record), nil
}

func (r *TripPostgresRepository) FindMany(ctx context.Context) ([]domain.Trip, error) {
	records, err := r.queries.ListTrips(ctx)
	if err != nil {
		return nil, err
	}

	trips := make([]domain.Trip, len(records))
	for i, record := range records {
		trips[i] = r.mapToTrip(record)
	}

	return trips, nil
}

func (r *TripPostgresRepository) mapToTrip(record Trip) domain.Trip {
	var id domain.TripID
	if record.ID.Valid {
		id, _ = domain.NewTripID(record.ID.String())
	}

	var createdAt time.Time
	if record.CreatedAt.Valid {
		createdAt = record.CreatedAt.Time
	}

	var updatedAt time.Time
	if record.UpdatedAt.Valid {
		updatedAt = record.UpdatedAt.Time
	}

	return domain.NewTrip(
		id,
		record.Name,
		createdAt,
		updatedAt,
	)
}

func (r *TripPostgresRepository) Create(ctx context.Context, trip domain.Trip) error {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(trip.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	if err := r.queries.CreateTrip(ctx, CreateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		CreatedAt: validatedCreatedAt,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
		return err
	}

	return nil
}

func (r *TripPostgresRepository) Update(ctx context.Context, trip domain.Trip) error {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	if err := r.queries.UpdateTrip(ctx, UpdateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
		return err
	}

	return nil
}

func (r *TripPostgresRepository) Delete(ctx context.Context, trip domain.Trip) error {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	if err := r.queries.DeleteTrip(ctx, validatedId); err != nil {
		return err
	}

	return nil
}
