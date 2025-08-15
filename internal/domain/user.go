package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination mock/user.go github.com/hata0/travel-api/internal/domain UserRepository
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id UserID) (*User, error)
}

type UserID struct {
	value string
}

func NewUserID(id string) UserID {
	return UserID{value: id}
}

func (id UserID) String() string {
	return id.value
}

func (id UserID) Equals(other UserID) bool {
	return id.value == other.value
}

type User struct {
	id           UserID
	username     string
	email        string
	passwordHash []byte
	createdAt    time.Time
	updatedAt    time.Time
}

func NewUser(id UserID, username, email string, passwordHash []byte, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Getters
func (u *User) ID() UserID           { return u.id }
func (u *User) Username() string     { return u.username }
func (u *User) Email() string        { return u.email }
func (u *User) PasswordHash() []byte { return u.passwordHash }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Update はユーザー情報を更新する
func (u *User) Update(username, email string, passwordHash []byte, updatedAt time.Time) *User {
	return &User{
		id:           u.id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    u.createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.id.Equals(other.id)
}
