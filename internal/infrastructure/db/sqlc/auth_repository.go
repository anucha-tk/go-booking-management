package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-booking-management-init/internal/domain/auth"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SQLCAuthRepository struct {
	queries Querier
	db      *sql.DB
}

func NewSQLCAuthRepository(db *sql.DB) *SQLCAuthRepository {
	return &SQLCAuthRepository{
		queries: New(db),
		db:      db,
	}
}

func (r *SQLCAuthRepository) Create(ctx context.Context, user *auth.User) (*auth.User, error) {
	params := CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
	}

	res, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		// Check for unique constraint violation (PostgreSQL code 23505)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, auth.ErrUserAlreadyExists
		}
		return nil, err
	}

	return &auth.User{
		ID:           res.ID,
		Email:        res.Email,
		PasswordHash: res.PasswordHash,
		Role:         auth.UserRole(res.Role),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}

func (r *SQLCAuthRepository) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	u, err := uuid.Parse(jti)
	if err != nil {
		return fmt.Errorf("invalid jti format: %w", err)
	}

	params := RevokeTokenParams{
		Jti:       u,
		ExpiresAt: expiresAt,
	}

	return r.queries.RevokeToken(ctx, params)
}

func (r *SQLCAuthRepository) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	u, err := uuid.Parse(jti)
	if err != nil {
		return false, fmt.Errorf("invalid jti format: %w", err)
	}

	return r.queries.IsTokenRevoked(ctx, u)
}

func (r *SQLCAuthRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	res, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &auth.User{
		ID:           res.ID,
		Email:        res.Email,
		PasswordHash: res.PasswordHash,
		Role:         auth.UserRole(res.Role),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}

func (r *SQLCAuthRepository) GetByID(ctx context.Context, id int32) (*auth.User, error) {
	res, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("database error in GetByID: %w", err)
	}

	return &auth.User{
		ID:           res.ID,
		Email:        res.Email,
		PasswordHash: res.PasswordHash,
		Role:         auth.UserRole(res.Role),
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}, nil
}
