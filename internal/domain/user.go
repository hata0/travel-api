package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/user.go travel-api/internal/domain UserRepository
type UserRepository interface {
	Create(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
}

type UserID struct {
	value string
}

func NewUserID(id string) (UserID, error) {
	if !IsValidUUID(id) {
		return UserID{}, ErrInvalidUUID
	}
	return UserID{value: id}, nil
}

func (id UserID) String() string {
	return id.value
}

type User struct {
	ID           UserID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id UserID, username, email, passwordHash string, createdAt, updatedAt time.Time) User {
	return User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
