package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserPostgresRepository struct {
	*BaseRepository
}

func NewUserPostgresRepository(db postgres.DBTX) domain.UserRepository {
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

	if err := queries.CreateUser(ctx, postgres.CreateUserParams{
		ID:           validatedId,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    validatedCreatedAt,
		UpdatedAt:    validatedUpdatedAt,
	}); err != nil {
		return apperr.NewInternalError(fmt.Sprintf("failed to create user: %s", user.ID.String()), err)
	}

	return nil
}

func (r *UserPostgresRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	record, err := queries.GetUserByEmail(ctx, email) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, apperr.ErrUserNotFound
		} else {
			return domain.User{}, apperr.NewInternalError(fmt.Sprintf("failed to find user: %s", email), err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	queries := r.getQueries(ctx) // ここで適切なQueriesインスタンスを取得

	record, err := queries.GetUserByUsername(ctx, username) // 取得したqueriesを使用
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, apperr.ErrUserNotFound
		} else {
			return domain.User{}, apperr.NewInternalError(fmt.Sprintf("failed to find user: %s", username), err)
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
			return domain.User{}, apperr.ErrUserNotFound
		} else {
			return domain.User{}, apperr.NewInternalError(fmt.Sprintf("failed to find user: %s", id.String()), err)
		}
	}

	return r.mapToUser(record), nil
}

func (r *UserPostgresRepository) mapToUser(record postgres.User) domain.User {
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
