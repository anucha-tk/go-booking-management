package db

import (
	"context"
	"database/sql"
	"errors"
	"go-booking-management-init/internal/domain/auth"
	"strings"
)

type SQLCAuthRepository struct {
	queries *Queries
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
		Role:         user.Role,
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
		Role:         res.Role,
		CreatedAt:    res.CreatedAt.Time,
		UpdatedAt:    res.UpdatedAt.Time,
	}, nil
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
		Role:         res.Role,
		CreatedAt:    res.CreatedAt.Time,
		UpdatedAt:    res.UpdatedAt.Time,
	}, nil
}
