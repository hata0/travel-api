package postgres

import (
	"context"
	"errors"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/app_error"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type TripPostgresRepository struct {
	*BaseRepository
}

func NewTripPostgresRepository(db DBTX) domain.TripRepository {
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
			return domain.Trip{}, app_error.ErrTripNotFound
		} else {
			return domain.Trip{}, app_error.NewInternalServerError(err)
		}
	}

	return r.mapToTrip(record), nil
}

func (r *TripPostgresRepository) FindMany(ctx context.Context) ([]domain.Trip, error) {
	queries := r.getQueries(ctx)
	records, err := queries.ListTrips(ctx)
	if err != nil {
		return nil, app_error.NewInternalServerError(err)
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
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(trip.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	if err := queries.CreateTrip(ctx, CreateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		CreatedAt: validatedCreatedAt,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			return app_error.ErrTripAlreadyExists
		}
		return app_error.NewInternalServerError(err)
	}

	return nil
}

func (r *TripPostgresRepository) Update(ctx context.Context, trip domain.Trip) error {
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(trip.UpdatedAt)

	if err := queries.UpdateTrip(ctx, UpdateTripParams{
		ID:        validatedId,
		Name:      trip.Name,
		UpdatedAt: validatedUpdatedAt,
	}); err != nil {
		return app_error.NewInternalServerError(err)
	}

	return nil
}

func (r *TripPostgresRepository) Delete(ctx context.Context, trip domain.Trip) error {
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(trip.ID.String())

	if err := queries.DeleteTrip(ctx, validatedId); err != nil {
		return app_error.NewInternalServerError(err)
	}

	return nil
}
