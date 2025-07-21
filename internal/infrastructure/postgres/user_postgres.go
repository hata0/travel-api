package postgres

import (
	"context"
	"errors"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/domain/shared/app_error"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserPostgresRepository struct {
	*BaseRepository
}

func NewUserPostgresRepository(db DBTX) domain.UserRepository {
	return &UserPostgresRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *UserPostgresRepository) Create(ctx context.Context, user domain.User) error {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	var validatedId pgtype.UUID
	_ = validatedId.Scan(user.ID.String())

	var validatedCreatedAt pgtype.Timestamptz
	_ = validatedCreatedAt.Scan(user.CreatedAt)

	var validatedUpdatedAt pgtype.Timestamptz
	_ = validatedUpdatedAt.Scan(user.UpdatedAt)

	if err := queries.CreateUser(ctx, CreateUserParams{
		ID:           validatedId,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    validatedCreatedAt,
		UpdatedAt:    validatedUpdatedAt,
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			return app_error.ErrUserAlreadyExists
		}
		return app_error.NewInternalServerError(err)
	}

	return nil
}

func (r *UserPostgresRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	record, err := queries.GetUserByEmail(ctx, email) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, app_error.ErrUserNotFound
		} else {
			return domain.User{}, app_error.NewInternalServerError(err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	record, err := queries.GetUserByUsername(ctx, username) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, app_error.ErrUserNotFound
		} else {
			return domain.User{}, app_error.NewInternalServerError(err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	var validatedId pgtype.UUID
	_ = validatedId.Scan(id.String())

	record, err := queries.GetUser(ctx, validatedId) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, app_error.ErrUserNotFound
		} else {
			return domain.User{}, app_error.NewInternalServerError(err)
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
