package postgres

import (
	"context" // errorsパッケージのインポートを追加
	"fmt"
	"time"
	"travel-api/internal/domain"
	domain_errors "travel-api/internal/domain/shared/errors"
	postgres "travel-api/internal/infrastructure/postgres/generated"
	shared_errors "travel-api/internal/shared/errors"

	"github.com/jackc/pgx/v5" // pgconnパッケージのインポートを追加
	"github.com/jackc/pgx/v5/pgtype"
)

type RevokedTokenPostgresRepository struct {
	*BaseRepository
}

func NewRevokedTokenPostgresRepository(db postgres.DBTX) domain.RevokedTokenRepository {
	return &RevokedTokenPostgresRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *RevokedTokenPostgresRepository) Create(ctx context.Context, token domain.RevokedToken) error {
	queries := r.getQueries(ctx)

	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedUserID pgtype.UUID
	_ = validatedUserID.Scan(token.UserID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedRevokedAt pgtype.Timestamptz
	_ = validatedRevokedAt.Scan(token.RevokedAt)

	if err := queries.CreateRevokedToken(ctx, postgres.CreateRevokedTokenParams{
		ID:        validatedId,
		UserID:    validatedUserID,
		TokenJti:  token.TokenJTI,
		ExpiresAt: validatedExpiresAt,
		RevokedAt: validatedRevokedAt,
	}); err != nil {
		return shared_errors.NewInternalError(fmt.Sprintf("failed to create revoked token: %s", token.ID.String()), err)
	}

	return nil
}

func (r *RevokedTokenPostgresRepository) FindByJTI(ctx context.Context, jti string) (domain.RevokedToken, error) {
	queries := r.getQueries(ctx)

	record, err := queries.GetRevokedTokenByJTI(ctx, jti)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.RevokedToken{}, domain_errors.ErrTokenNotFound
		} else {
			return domain.RevokedToken{}, shared_errors.NewInternalError(fmt.Sprintf("failed to find revoked token: %s", jti), err)
		}
	}

	return r.mapToRevokedToken(record), nil
}

func (r *RevokedTokenPostgresRepository) mapToRevokedToken(record postgres.RevokedToken) domain.RevokedToken {
	var id domain.RevokedTokenID
	if record.ID.Valid {
		id, _ = domain.NewRevokedTokenID(record.ID.String())
	}

	var userID domain.UserID
	if record.UserID.Valid {
		userID, _ = domain.NewUserID(record.UserID.String())
	}

	var expiresAt time.Time
	if record.ExpiresAt.Valid {
		expiresAt = record.ExpiresAt.Time
	}

	var revokedAt time.Time
	if record.RevokedAt.Valid {
		revokedAt = record.RevokedAt.Time
	}

	return domain.NewRevokedToken(
		id,
		userID,
		record.TokenJti,
		expiresAt,
		revokedAt,
	)
}
