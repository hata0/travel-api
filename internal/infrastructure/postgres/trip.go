package postgres

import (
	"context"
	"errors"

	apperr "github.com/hata0/travel-api/internal/domain/errors"
	"github.com/hata0/travel-api/internal/domain/trip"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
)

// TripPostgresRepository はTripエンティティのPostgreSQL実装
type TripPostgresRepository struct {
	*BasePostgresRepository
}

// NewTripPostgresRepository は新しいTripPostgresRepositoryを作成する
func NewTripPostgresRepository(db postgres.DBTX) trip.TripRepository {
	return &TripPostgresRepository{
		BasePostgresRepository: NewBasePostgresRepository(db),
	}
}

// FindByID は指定されたIDのTripを取得する
func (r *TripPostgresRepository) FindByID(ctx context.Context, id trip.TripID) (*trip.Trip, error) {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(id.String())
	if err != nil {
		return nil, apperr.NewInternalError("Failed to convert trip ID to UUID", apperr.WithCause(err))
	}

	record, err := queries.FindTrip(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, trip.NewTripNotFoundError()
		}
		return nil, apperr.NewInternalError("Failed to fetch trip from database", apperr.WithCause(err))
	}

	trip, err := r.mapToTrip(record)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to map database record to trip domain object", apperr.WithCause(err))
	}

	return trip, nil
}

// FindMany はすべてのTripを取得する
func (r *TripPostgresRepository) FindMany(ctx context.Context) ([]*trip.Trip, error) {
	queries := r.GetQueries(ctx)

	records, err := queries.ListTrips(ctx)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to fetch trips list from database", apperr.WithCause(err))
	}

	trips := make([]*trip.Trip, 0, len(records))
	for _, record := range records {
		trip, err := r.mapToTrip(record)
		if err != nil {
			// ログに記録して続行（一つのエラーで全体を失敗させない）
			// TODO: ロガーを追加してエラーをログに記録
			continue
		}
		trips = append(trips, trip)
	}

	return trips, nil
}

// Create は新しいTripを作成する
func (r *TripPostgresRepository) Create(ctx context.Context, trip *trip.Trip) error {
	if trip == nil {
		return apperr.NewInternalError("Trip entity cannot be nil")
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(trip.ID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert trip ID to UUID for creation", apperr.WithCause(err))
	}

	pgCreatedAt, err := mapper.ToTimestamp(trip.CreatedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert trip created_at to timestamp", apperr.WithCause(err))
	}

	pgUpdatedAt, err := mapper.ToTimestamp(trip.UpdatedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert trip updated_at to timestamp", apperr.WithCause(err))
	}

	params := postgres.CreateTripParams{
		ID:        pgUUID,
		Name:      trip.Name(),
		CreatedAt: pgCreatedAt,
		UpdatedAt: pgUpdatedAt,
	}

	if err := queries.CreateTrip(ctx, params); err != nil {
		return apperr.NewInternalError("Failed to create trip in database", apperr.WithCause(err))
	}

	return nil
}

// Update は既存のTripを更新する
func (r *TripPostgresRepository) Update(ctx context.Context, trip *trip.Trip) error {
	if trip == nil {
		return apperr.NewInternalError("Trip entity cannot be nil")
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(trip.ID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert trip ID to UUID for update", apperr.WithCause(err))
	}

	pgUpdatedAt, err := mapper.ToTimestamp(trip.UpdatedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert trip updated_at to timestamp for update", apperr.WithCause(err))
	}

	params := postgres.UpdateTripParams{
		ID:        pgUUID,
		Name:      trip.Name(),
		UpdatedAt: pgUpdatedAt,
	}

	if err := queries.UpdateTrip(ctx, params); err != nil {
		return apperr.NewInternalError("Failed to update trip in database", apperr.WithCause(err))
	}

	return nil
}

// Delete は指定されたIDのTripを削除する
func (r *TripPostgresRepository) Delete(ctx context.Context, id trip.TripID) error {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(id.String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert trip ID to UUID for deletion", apperr.WithCause(err))
	}

	rows, err := queries.DeleteTrip(ctx, pgUUID)
	if err != nil {
		return apperr.NewInternalError("Failed to delete trip from database", apperr.WithCause(err))
	}

	if rows == 0 {
		return trip.NewTripNotFoundError()
	}

	return nil
}

// mapToTrip はデータベースレコードをドメインオブジェクトに変換する
func (r *TripPostgresRepository) mapToTrip(record postgres.Trip) (*trip.Trip, error) {
	mapper := r.GetTypeMapper()

	id, err := mapper.FromUUID(record.ID)
	if err != nil {
		return nil, err
	}

	createdAt, err := mapper.FromTimestamp(record.CreatedAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := mapper.FromTimestamp(record.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return trip.NewTrip(
		trip.NewTripID(id),
		record.Name,
		createdAt,
		updatedAt,
	), nil
}
