package postgres

import (
	"context"
	"errors"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
)

// RefreshTokenPostgresRepository はRefreshTokenエンティティのPostgreSQL実装
type RefreshTokenPostgresRepository struct {
	*BasePostgresRepository
}

// NewRefreshTokenPostgresRepository は新しいRefreshTokenPostgresRepositoryを作成する
func NewRefreshTokenPostgresRepository(db postgres.DBTX) domain.RefreshTokenRepository {
	return &RefreshTokenPostgresRepository{
		BasePostgresRepository: NewBasePostgresRepository(db),
	}
}

// Create は新しいRefreshTokenを作成する
func (r *RefreshTokenPostgresRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	if token == nil {
		return apperr.NewInternalError("RefreshToken entity cannot be nil", nil)
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgID, err := mapper.ToUUID(token.ID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert refresh token ID to UUID for creation", err)
	}

	pgUserID, err := mapper.ToUUID(token.UserID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert user ID to UUID for creation", err)
	}

	pgExpiresAt, err := mapper.ToTimestamp(token.ExpiresAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert expires_at to timestamp", err)
	}

	pgCreatedAt, err := mapper.ToTimestamp(token.CreatedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert created_at to timestamp", err)
	}

	params := postgres.CreateRefreshTokenParams{
		ID:        pgID,
		UserID:    pgUserID,
		Token:     token.Token(),
		ExpiresAt: pgExpiresAt,
		CreatedAt: pgCreatedAt,
	}

	if err := queries.CreateRefreshToken(ctx, params); err != nil {
		return apperr.NewInternalError("Failed to create refresh token in database", err)
	}

	return nil
}

// FindByToken は指定されたTokenのRefreshTokenを取得する
func (r *RefreshTokenPostgresRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	queries := r.GetQueries(ctx)

	record, err := queries.FindRefreshTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrRefreshTokenNotFound
		}
		return nil, apperr.NewInternalError("Failed to fetch refresh token by token from database", err)
	}

	refreshToken, err := r.mapToRefreshToken(record)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to map database record to refresh token domain object", err)
	}

	return refreshToken, nil
}

// Delete は指定されたRefreshTokenを削除する
func (r *RefreshTokenPostgresRepository) Delete(ctx context.Context, id domain.RefreshTokenID) error {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgID, err := mapper.ToUUID(id.String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert refresh token ID to UUID for deletion", err)
	}

	rows, err := queries.DeleteRefreshToken(ctx, pgID)
	if err != nil {
		return apperr.NewInternalError("Failed to delete refresh token from database", err)
	}

	if rows == 0 {
		return apperr.ErrRefreshTokenNotFound
	}

	return nil
}

// DeleteByUserID は指定されたUserIDのRefreshTokenを削除する
func (r *RefreshTokenPostgresRepository) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUserID, err := mapper.ToUUID(userID.String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert user ID to UUID for deletion by user ID", err)
	}

	if err = queries.DeleteRefreshTokenByUserID(ctx, pgUserID); err != nil {
		return apperr.NewInternalError("Failed to delete refresh token by user ID from database", err)
	}

	return nil
}

// mapToRefreshToken はデータベースレコードをドメインオブジェクトに変換する
func (r *RefreshTokenPostgresRepository) mapToRefreshToken(record postgres.RefreshToken) (*domain.RefreshToken, error) {
	mapper := r.GetTypeMapper()

	id, err := mapper.FromUUID(record.ID)
	if err != nil {
		return nil, err
	}

	userID, err := mapper.FromUUID(record.UserID)
	if err != nil {
		return nil, err
	}

	expiresAt, err := mapper.FromTimestamp(record.ExpiresAt)
	if err != nil {
		return nil, err
	}

	createdAt, err := mapper.FromTimestamp(record.CreatedAt)
	if err != nil {
		return nil, err
	}

	return domain.NewRefreshToken(
		domain.NewRefreshTokenID(id),
		domain.NewUserID(userID),
		record.Token,
		expiresAt,
		createdAt,
	), nil
}
