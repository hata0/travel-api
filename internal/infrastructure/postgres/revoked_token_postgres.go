package postgres

import (
	"context"
	"time"
	"travel-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RevokedTokenPostgresRepository struct {
	queries *Queries
}

func NewRevokedTokenPostgresRepository(db DBTX) domain.RevokedTokenRepository {
	return &RevokedTokenPostgresRepository{
		queries: New(db),
	}
}

func (r *RevokedTokenPostgresRepository) Create(ctx context.Context, token domain.RevokedToken) error {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(token.ID.String())

	var validatedExpiresAt pgtype.Timestamptz
	_ = validatedExpiresAt.Scan(token.ExpiresAt)

	var validatedRevokedAt pgtype.Timestamptz
	_ = validatedRevokedAt.Scan(token.RevokedAt)

	if err := r.queries.CreateRevokedToken(ctx, CreateRevokedTokenParams{
		ID:        validatedId,
		TokenJti:  token.TokenJTI,
		ExpiresAt: validatedExpiresAt,
		RevokedAt: validatedRevokedAt,
	}); err != nil {
		return err
	}

	return nil
}

func (r *RevokedTokenPostgresRepository) FindByJTI(ctx context.Context, jti string) (domain.RevokedToken, error) {
	record, err := r.queries.GetRevokedTokenByJTI(ctx, jti)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.RevokedToken{}, domain.ErrTokenNotFound
		} else {
			return domain.RevokedToken{}, err
		}
	}

	return r.mapToRevokedToken(record), nil
}

func (r *RevokedTokenPostgresRepository) mapToRevokedToken(record RevokedToken) domain.RevokedToken {
	var id domain.RevokedTokenID
	if record.ID.Valid {
		id, _ = domain.NewRevokedTokenID(record.ID.String())
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
		record.TokenJti,
		expiresAt,
		revokedAt,
	)
}
