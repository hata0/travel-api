package postgres

import (
	"context"
	"errors"

	"github.com/hata0/travel-api/internal/domain"
	apperr "github.com/hata0/travel-api/internal/domain/errors"
	postgres "github.com/hata0/travel-api/internal/infrastructure/postgres/generated"
	"github.com/jackc/pgx/v5"
)

// UserPostgresRepository はUserエンティティのPostgreSQL実装
type UserPostgresRepository struct {
	*BasePostgresRepository
}

// NewUserPostgresRepository は新しいUserPostgresRepositoryを作成する
func NewUserPostgresRepository(db postgres.DBTX) domain.UserRepository {
	return &UserPostgresRepository{
		BasePostgresRepository: NewBasePostgresRepository(db),
	}
}

// Create は新しいUserを作成する
func (r *UserPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return apperr.NewInternalError("User entity cannot be nil", nil)
	}

	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(user.ID().String())
	if err != nil {
		return apperr.NewInternalError("Failed to convert user ID to UUID for creation", err)
	}

	pgCreatedAt, err := mapper.ToTimestamp(user.CreatedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert user created_at to timestamp", err)
	}

	pgUpdatedAt, err := mapper.ToTimestamp(user.UpdatedAt())
	if err != nil {
		return apperr.NewInternalError("Failed to convert user updated_at to timestamp", err)
	}

	params := postgres.CreateUserParams{
		ID:           pgUUID,
		Username:     user.Username(),
		Email:        user.Email(),
		PasswordHash: user.PasswordHash(),
		CreatedAt:    pgCreatedAt,
		UpdatedAt:    pgUpdatedAt,
	}

	if err := queries.CreateUser(ctx, params); err != nil {
		return apperr.NewInternalError("Failed to create user in database", err)
	}

	return nil
}

// FindByEmail は指定されたEmailのUserを取得する
func (r *UserPostgresRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	queries := r.GetQueries(ctx)

	record, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, apperr.NewInternalError("Failed to fetch user by email from database", err)
	}

	user, err := r.mapToUser(record)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to map database record to user domain object", err)
	}

	return user, nil
}

// FindByUsername は指定されたUsernameのUserを取得する
func (r *UserPostgresRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	queries := r.GetQueries(ctx)

	record, err := queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, apperr.NewInternalError("Failed to fetch user by username from database", err)
	}

	user, err := r.mapToUser(record)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to map database record to user domain object", err)
	}

	return user, nil
}

// FindByID は指定されたIDのUserを取得する
func (r *UserPostgresRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	queries := r.GetQueries(ctx)
	mapper := r.GetTypeMapper()

	pgUUID, err := mapper.ToUUID(id.String())
	if err != nil {
		return nil, apperr.NewInternalError("Failed to convert user ID to UUID", err)
	}

	record, err := queries.GetUser(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrUserNotFound
		}
		return nil, apperr.NewInternalError("Failed to fetch user from database", err)
	}

	user, err := r.mapToUser(record)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to map database record to user domain object", err)
	}

	return user, nil
}

// mapToUser はデータベースレコードをドメインオブジェクトに変換する
func (r *UserPostgresRepository) mapToUser(record postgres.User) (*domain.User, error) {
	mapper := r.GetTypeMapper()

	id, err := mapper.FromUUID(record.ID)
	if err != nil {
		return nil, err
	}

	createdAt, err := mapper.FromTimestamp(record.CreatedAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := mapper.FromTimestamp(record.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return domain.NewUser(
		domain.NewUserID(id),
		record.Username,
		record.Email,
		record.PasswordHash,
		createdAt,
		updatedAt,
	), nil
}
