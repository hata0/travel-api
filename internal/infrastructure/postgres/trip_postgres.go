package postgres

import (
	"context"
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
	if err := validatedId.Scan(string(id)); err != nil {
		return domain.Trip{}, err
	}

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
	return domain.NewTrip(
		domain.TripID(record.ID.String()),
		record.Name,
		record.CreatedAt.Time,
		record.UpdatedAt.Time,
	)
}

func (r *TripPostgresRepository) Create(ctx context.Context, trip domain.Trip) error {
	var validatedId pgtype.UUID
	if err := validatedId.Scan(string(trip.ID)); err != nil {
		return err
	}

	var validatedCreatedAt pgtype.Timestamptz
	if err := validatedCreatedAt.Scan(trip.CreatedAt); err != nil {
		return err
	}

	var validatedUpdatedAt pgtype.Timestamptz
	if err := validatedUpdatedAt.Scan(trip.UpdatedAt); err != nil {
		return err
	}

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
	if err := validatedId.Scan(string(trip.ID)); err != nil {
		return err
	}

	var validatedUpdatedAt pgtype.Timestamptz
	if err := validatedUpdatedAt.Scan(trip.UpdatedAt); err != nil {
		return err
	}

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
	if err := validatedId.Scan(string(trip.ID)); err != nil {
		return err
	}

	if err := r.queries.DeleteTrip(ctx, validatedId); err != nil {
		return err
	}

	return nil
}
