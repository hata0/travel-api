package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TripPostgresRepository struct {
	*BaseRepository
}

func NewTripPostgresRepository(db postgres.DBTX) domain.TripRepository {
	return &TripPostgresRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *TripPostgresRepository) FindByID(ctx context.Context, id domain.TripID) (domain.Trip, error) {
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(id.String())

	record, err := queries.GetTrip(ctx, validatedId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Trip{}, apperr.ErrTripNotFound
		} else {
			return domain.Trip{}, apperr.NewInternalError(fmt.Sprintf("failed to find trip: %s", id.String()), err)
		}
	}

	return r.mapToTrip(record), nil
}

func (r *TripPostgresRepository) FindMany(ctx context.Context) ([]domain.Trip, error) {
	queries := r.getQueries(ctx)
	records, err := queries.ListTrips(ctx)
	if err != nil {
		return nil, apperr.NewInternalError("failed to find trips", err)
	}

	trips := make([]domain.Trip, len(records))
	for i, record := range records {
		trips[i] = r.mapToTrip(record)
	}

	return trips, nil
}

func (r *TripPostgresRepository) mapToTrip(record postgres.Trip) domain.Trip {
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
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(trip.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	if err := queries.CreateTrip(ctx, postgres.CreateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		CreatedAt: validatedCreatedAt,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to create trip: %s", trip.ID.String()), err)
	}

	return nil
}

func (r *TripPostgresRepository) Update(ctx context.Context, trip domain.Trip) error {
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	if err := queries.UpdateTrip(ctx, postgres.UpdateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to update trip: %s", trip.ID), err)
	}

	return nil
}

func (r *TripPostgresRepository) Delete(ctx context.Context, trip domain.Trip) error {
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	if err := queries.DeleteTrip(ctx, validatedId); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to delete trip: %s", trip.ID), err)
	}

	return nil
}
