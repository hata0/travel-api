package postgres

import (
	"context"
	"errors" // errorsパッケージのインポートを追加
	"time"
	"travel-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn" // pgconnパッケージのインポートを追加
	"github.com/jackc/pgx/v5/pgtype"
)

type RevokedTokenPostgresRepository struct {
	*BaseRepository
}

func NewRevokedTokenPostgresRepository(db DBTX) domain.RevokedTokenRepository {
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

	if err := queries.CreateRevokedToken(ctx, CreateRevokedTokenParams{
		ID:        validatedId,
		UserID:    validatedUserID,
		TokenJti:  token.TokenJTI,
		ExpiresAt: validatedExpiresAt,
		RevokedAt: validatedRevokedAt,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			return domain.ErrTokenAlreadyExists
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}

func (r *RevokedTokenPostgresRepository) FindByJTI(ctx context.Context, jti string) (domain.RevokedToken, error) {
	queries := r.getQueries(ctx)

	record, err := queries.GetRevokedTokenByJTI(ctx, jti)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.RevokedToken{}, domain.ErrTokenNotFound
		} else {
			return domain.RevokedToken{}, domain.NewInternalServerError(err)
		}
	}

	return r.mapToRevokedToken(record), nil
}

func (r *RevokedTokenPostgresRepository) mapToRevokedToken(record RevokedToken) domain.RevokedToken {
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
