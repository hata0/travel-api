package postgres

import (
	"context"
	"errors"

	apperr "github.com/hata0/travel-api/internal/domain/errors"
	revokedtoken "github.com/hata0/travel-api/internal/domain/revoked_token"
	"github.com/hata0/travel-api/internal/domain/user"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
)

// RevokedTokenPostgresRepository はRevokedTokenエンティティのPostgreSQL実装
type RevokedTokenPostgresRepository struct {
	*BasePostgresRepository
}

// NewRevokedTokenPostgresRepository は新しいRevokedTokenPostgresRepositoryを作成する
func NewRevokedTokenPostgresRepository(db postgres.DBTX) revokedtoken.RevokedTokenRepository {
	return &RevokedTokenPostgresRepository{
		BasePostgresRepository: NewBasePostgresRepository(db),
	}
}

// Create は新しいRevokedTokenを作成する
func (r *RevokedTokenPostgresRepository) Create(ctx context.Context, token *revokedtoken.RevokedToken) error {
	if token == nil {
		return apperr.NewInternalError("RevokedToken entity cannot be nil")
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgID, err := mapper.ToUUID(token.ID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert revoked token ID to UUID for creation", apperr.WithCause(err))
	}

	pgUserID, err := mapper.ToUUID(token.UserID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert user ID to UUID for creation", apperr.WithCause(err))
	}

	pgExpiresAt, err := mapper.ToTimestamp(token.ExpiresAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert expires_at to timestamp", apperr.WithCause(err))
	}

	pgRevokedAt, err := mapper.ToTimestamp(token.RevokedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert revoked_at to timestamp", apperr.WithCause(err))
	}

	params := postgres.CreateRevokedTokenParams{
		ID:        pgID,
		UserID:    pgUserID,
		TokenJti:  token.TokenJTI(),
		ExpiresAt: pgExpiresAt,
		RevokedAt: pgRevokedAt,
	}

	if err := queries.CreateRevokedToken(ctx, params); err != nil {
		return apperr.NewInternalError("Failed to create revoked token in database", apperr.WithCause(err))
	}

	return nil
}

// FindByJTI は指定されたJTIのRevokedTokenを取得する
func (r *RevokedTokenPostgresRepository) FindByJTI(ctx context.Context, jti string) (*revokedtoken.RevokedToken, error) {
	queries := r.GetQueries(ctx)

	record, err := queries.FindRevokedTokenByJTI(ctx, jti)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, revokedtoken.NewRevokedTokenNotFoundError()
		}
		return nil, apperr.NewInternalError("Failed to fetch revoked token by JTI from database", apperr.WithCause(err))
	}

	revokedToken, err := r.mapToRevokedToken(record)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to map database record to revoked token domain object", apperr.WithCause(err))
	}

	return revokedToken, nil
}

// mapToRevokedToken はデータベースレコードをドメインオブジェクトに変換する
func (r *RevokedTokenPostgresRepository) mapToRevokedToken(record postgres.RevokedToken) (*revokedtoken.RevokedToken, error) {
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

	revokedAt, err := mapper.FromTimestamp(record.RevokedAt)
	if err != nil {
		return nil, err
	}

	return revokedtoken.NewRevokedToken(
		revokedtoken.NewRevokedTokenID(id),
		user.NewUserID(userID),
		record.TokenJti,
		expiresAt,
		revokedAt,
	), nil
}
