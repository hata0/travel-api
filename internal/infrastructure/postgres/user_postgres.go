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

type UserPostgresRepository struct {
	queries *Queries
}

func NewUserPostgresRepository(db DBTX) domain.UserRepository {
	return &UserPostgresRepository{
		queries: New(db),
	}
}

func (r *UserPostgresRepository) Create(ctx context.Context, user domain.User) error {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(user.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(user.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(user.UpdatedAt)

	if err := r.queries.CreateUser(ctx, CreateUserParams{
		ID:           validatedId,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    validatedCreatedAt,
		UpdatedAt:    validatedUpdatedAt,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			return domain.ErrUserAlreadyExists
		}
		return domain.NewInternalServerError(err)
	}

	return nil
}

func (r *UserPostgresRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	record, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, domain.ErrUserNotFound
		} else {
			return domain.User{}, domain.NewInternalServerError(err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	record, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, domain.ErrUserNotFound
		} else {
			return domain.User{}, domain.NewInternalServerError(err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	var validatedId pgtype.UUID
	_ = validatedId.Scan(id.String())

	record, err := r.queries.GetUser(ctx, validatedId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, domain.ErrUserNotFound
		} else {
			return domain.User{}, domain.NewInternalServerError(err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) mapToUser(record User) domain.User {
	var id domain.UserID
	if record.ID.Valid {
		id, _ = domain.NewUserID(record.ID.String())
	}

	var createdAt time.Time
	if record.CreatedAt.Valid {
		createdAt = record.CreatedAt.Time
	}

	var updatedAt time.Time
	if record.UpdatedAt.Valid {
		updatedAt = record.UpdatedAt.Time
	}

	return domain.NewUser(
		id,
		record.Username,
		record.Email,
		record.PasswordHash,
		createdAt,
		updatedAt,
	)
}
