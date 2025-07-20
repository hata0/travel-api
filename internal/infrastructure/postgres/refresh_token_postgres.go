package postgres

import (
	"context"
	"errors"
	"time"
	"travel-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type RefreshTokenPostgresRepository struct {
	queries *Queries
}

func NewRefreshTokenPostgresRepository(db DBTX) domain.RefreshTokenRepository {
	return &RefreshTokenPostgresRepository{
		queries: New(db),
	}
}

func (r *RefreshTokenPostgresRepository) Create(ctx context.Context, token domain.RefreshToken) error {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedUserID pgtype.UUID
	_ = validatedUserID.Scan(token.UserID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(token.CreatedAt)

	if err := r.queries.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		ID:        validatedId,
		UserID:    validatedUserID,
		Token:     token.Token,
		ExpiresAt: validatedExpiresAt,
		CreatedAt: validatedCreatedAt,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			return domain.ErrTokenAlreadyExists
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}

func (r *RefreshTokenPostgresRepository) FindByToken(ctx context.Context, token string) (domain.RefreshToken, error) {
	record, err := r.queries.FindRefreshTokenByToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.RefreshToken{}, domain.ErrTokenNotFound
		} else {
			return domain.RefreshToken{}, domain.NewInternalServerError(err)
		}
	}

	return r.mapToRefreshToken(record), nil
}

func (r *RefreshTokenPostgresRepository) Delete(ctx context.Context, token string) error {
	if err := r.queries.DeleteRefreshToken(ctx, token); err != nil {
		return domain.NewInternalServerError(err)
	}
	return nil
}

func (r *RefreshTokenPostgresRepository) mapToRefreshToken(record RefreshToken) domain.RefreshToken {
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
