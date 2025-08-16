package user

import "context"

//go:generate mockgen -destination mock/user.go github.com/hata0/travel-api/internal/domain/user UserRepository
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id UserID) (*User, error)
}
