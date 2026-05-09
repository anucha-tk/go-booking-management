package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int32) (*User, error)
	RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}
