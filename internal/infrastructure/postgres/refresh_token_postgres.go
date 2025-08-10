package postgres

import (
	"context"
	"fmt"
	"time"
	"travel-api/internal/domain"
	apperr "travel-api/internal/domain/errors"
	postgres "travel-api/internal/infrastructure/postgres/generated"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RefreshTokenPostgresRepository struct {
	*BaseRepository
}

func NewRefreshTokenPostgresRepository(db postgres.DBTX) domain.RefreshTokenRepository {
	return &RefreshTokenPostgresRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *RefreshTokenPostgresRepository) Create(ctx context.Context, token domain.RefreshToken) error {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedUserID pgtype.UUID
	_ = validatedUserID.Scan(token.UserID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(token.CreatedAt)

	if err := queries.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{ // 取得したqueriesを使用
		ID:        validatedId,
		UserID:    validatedUserID,
		Token:     token.Token,
		ExpiresAt: validatedExpiresAt,
		CreatedAt: validatedCreatedAt,
	}); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to create refresh token: %s", token.ID.String()), err)
	}

	return nil
}

func (r *RefreshTokenPostgresRepository) FindByToken(ctx context.Context, token string) (domain.RefreshToken, error) {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	record, err := queries.FindRefreshTokenByToken(ctx, token) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.RefreshToken{}, apperr.ErrTokenNotFound
		} else {
			return domain.RefreshToken{}, apperr.NewInternalError(fmt.Sprintf("failed to find refresh token: %s", token), err)
		}
	}

	return r.mapToRefreshToken(record), nil
}

func (r *RefreshTokenPostgresRepository) Delete(ctx context.Context, token domain.RefreshToken) error {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	if err := queries.DeleteRefreshToken(ctx, validatedId); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to delete refresh token: %s", token.ID.String()), err)
	}
	return nil
}

func (r *RefreshTokenPostgresRepository) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	queries := r.getQueries(ctx)

	var validatedUserID pgtype.UUID
	_ = validatedUserID.Scan(userID.String())

	if err := queries.DeleteRefreshTokensByUserID(ctx, validatedUserID); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to delete refresh token for user: %s", userID.String()), err)
	}
	return nil
}

func (r *RefreshTokenPostgresRepository) mapToRefreshToken(record postgres.RefreshToken) domain.RefreshToken {
	var id domain.RefreshTokenID
	if record.ID.Valid {
		id, _ = domain.NewRefreshTokenID(record.ID.String())
	}

	var userID domain.UserID
	if record.UserID.Valid {
		userID, _ = domain.NewUserID(record.UserID.String())
	}

	var expiresAt time.Time
	if record.ExpiresAt.Valid {
		expiresAt = record.ExpiresAt.Time
	}

	var createdAt time.Time
	if record.CreatedAt.Valid {
		createdAt = record.CreatedAt.Time
	}

	return domain.NewRefreshToken(
		id,
		userID,
		record.Token,
		expiresAt,
		createdAt,
	)
}
