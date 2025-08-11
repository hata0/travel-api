package postgres

import (
	"context"
	"errors"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
)

// TripPostgresRepository はTripエンティティのPostgreSQL実装
type TripPostgresRepository struct {
	*BasePostgresRepository
}

// NewTripPostgresRepository は新しいTripPostgresRepositoryを作成する
func NewTripPostgresRepository(db postgres.DBTX) domain.TripRepository {
	return &TripPostgresRepository{
		BasePostgresRepository: NewBasePostgresRepository(db),
	}
}

// FindByID は指定されたIDのTripを取得する
func (r *TripPostgresRepository) FindByID(ctx context.Context, id domain.TripID) (*domain.Trip, error) {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(id.String())
	if err != nil {
		return nil, apperr.NewInternalError("failed to convert trip ID to UUID", err)
	}

	record, err := queries.GetTrip(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrTripNotFound
		}
		return nil, apperr.NewInternalError("failed to fetch trip from database", err)
	}

	trip, err := r.mapToTrip(record)
	if err != nil {
		return nil, apperr.NewInternalError("failed to map database record to trip domain object", err)
	}

	return trip, nil
}

// FindMany はすべてのTripを取得する
func (r *TripPostgresRepository) FindMany(ctx context.Context) ([]*domain.Trip, error) {
	queries := r.GetQueries(ctx)

	records, err := queries.ListTrips(ctx)
	if err != nil {
		return nil, apperr.NewInternalError("failed to fetch trips list from database", err)
	}

	trips := make([]*domain.Trip, 0, len(records))
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
func (r *TripPostgresRepository) Create(ctx context.Context, trip *domain.Trip) error {
	if trip == nil {
		return apperr.NewInternalError("trip entity cannot be nil", nil)
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(trip.ID().String())
	if err != nil {
		return apperr.NewInternalError("failed to convert trip ID to UUID for creation", err)
	}

	pgCreatedAt, err := mapper.ToTimestamp(trip.CreatedAt())
	if err != nil {
		return apperr.NewInternalError("failed to convert trip created_at to timestamp", err)
	}

	pgUpdatedAt, err := mapper.ToTimestamp(trip.UpdatedAt())
	if err != nil {
		return apperr.NewInternalError("failed to convert trip updated_at to timestamp", err)
	}

	params := postgres.CreateTripParams{
		ID:        pgUUID,
		Name:      trip.Name(),
		CreatedAt: pgCreatedAt,
		UpdatedAt: pgUpdatedAt,
	}

	if err := queries.CreateTrip(ctx, params); err != nil {
		return apperr.NewInternalError("failed to create trip in database", err)
	}

	return nil
}

// Update は既存のTripを更新する
func (r *TripPostgresRepository) Update(ctx context.Context, trip *domain.Trip) error {
	if trip == nil {
		return apperr.NewInternalError("trip entity cannot be nil", nil)
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(trip.ID().String())
	if err != nil {
		return apperr.NewInternalError("failed to convert trip ID to UUID for update", err)
	}

	pgUpdatedAt, err := mapper.ToTimestamp(trip.UpdatedAt())
	if err != nil {
		return apperr.NewInternalError("failed to convert trip updated_at to timestamp for update", err)
	}

	params := postgres.UpdateTripParams{
		ID:        pgUUID,
		Name:      trip.Name(),
		UpdatedAt: pgUpdatedAt,
	}

	if err := queries.UpdateTrip(ctx, params); err != nil {
		return apperr.NewInternalError("failed to update trip in database", err)
	}

	return nil
}

// Delete は指定されたIDのTripを削除する
func (r *TripPostgresRepository) Delete(ctx context.Context, id domain.TripID) error {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(id.String())
	if err != nil {
		return apperr.NewInternalError("failed to convert trip ID to UUID for deletion", err)
	}

	if err := queries.DeleteTrip(ctx, pgUUID); err != nil {
		return apperr.NewInternalError("failed to delete trip from database", err)
	}

	return nil
}

// mapToTrip はデータベースレコードをドメインオブジェクトに変換する
func (r *TripPostgresRepository) mapToTrip(record postgres.Trip) (*domain.Trip, error) {
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

	return domain.NewTrip(
		domain.NewTripID(id),
		record.Name,
		createdAt,
		updatedAt,
	), nil
}
